package fgd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bzick/tokenizer"
)

type propParserState int

const (
	ppsNone propParserState = iota
	ppsBeforeName
	ppsBeforeType
	ppsBeforeDescription
	ppsInChoices
	ppsEnd
)

func parseEntityProperty(stream *tokenizer.Stream) (Property, error) {
	var (
		cur   Property
		state = ppsBeforeName
	)

	for stream.IsValid() {
		var err error
		switch state {
		case ppsBeforeName:
			state, err = parseEntityPropertyName(stream, &cur)
		case ppsBeforeType:
			state, err = parseEntityPropertyType(stream, &cur)
		case ppsBeforeDescription:
			state, err = parseEntityPropertyDescription(stream, &cur)
		case ppsInChoices:
			state, err = parseEntityPropertyChoices(stream, &cur)
		case ppsEnd:
			return cur, nil
		case ppsNone:
			err = errors.New("entity property parser reached an invalid state")
		}

		if err != nil {
			return Property{}, err
		}
	}

	return Property{}, errors.New("unexpected EOF while parsing entity property")
}

func parseEntityPropertyChoices(stream *tokenizer.Stream, cur *Property) (propParserState, error) {
	if stream.CurrentToken().Key() == TokenBracketClose {
		return ppsEnd, nil
	}

	if stream.CurrentToken().StringKey() == TokenComment {
		stream.GoNext()
		return ppsInChoices, nil
	}

	var i int
	values := make([]string, 4)
	for {
		parsed, err := parseValue(stream)
		if err != nil {
			return ppsNone, fmt.Errorf("unable to parse value: %w", err)
		}

		values[i] = parsed
		i++
		stream.GoNext()

		if stream.CurrentToken().Key() != TokenColon {
			break
		}
		stream.GoNext()
	}

	cur.Choices = append(cur.Choices, ChoiceEntry{
		Value:              values[0],
		Name:               values[1],
		DefaultValue:       values[2],
		ExtraDocumentation: values[3],
	})

	return ppsInChoices, nil
}

func parseEntityPropertyName(stream *tokenizer.Stream, cur *Property) (propParserState, error) {
	var builder strings.Builder

	for stream.CurrentToken().Key() != TokenParenOpen {
		builder.WriteString(stream.CurrentToken().ValueString())
		stream.GoNext()
	}

	cur.Name = builder.String()

	return ppsBeforeType, nil
}

func parseEntityPropertyType(stream *tokenizer.Stream, cur *Property) (propParserState, error) {
	stream.GoPrev()

	if !stream.IsNextSequence(TokenParenOpen, tokenizer.TokenKeyword, TokenParenClose) {
		return ppsNone, errors.New(`unexpected token sequence when parsing entity property type`)
	}

	advance(stream, 2)
	cur.Type = PropertyType(strings.ToLower(stream.CurrentToken().ValueString()))
	advance(stream, 2)

	return ppsBeforeDescription, nil
}

func parseEntityPropertyDescription(stream *tokenizer.Stream, cur *Property) (propParserState, error) {
	var values []string

	for stream.CurrentToken().Key() == TokenColon {
		stream.GoNext()

		if stream.CurrentToken().Key() == TokenColon {
			values = append(values, "")
			continue
		}

		value, err := parseValue(stream)
		if err != nil {
			return ppsNone, fmt.Errorf("unable to parse value: %w", err)
		}
		values = append(values, value)
		stream.GoNext()
	}

	// All optional, in order: name, default, description, extra description (jack extension)
	switch len(values) {
	case 3:
		cur.ExtraDocumentation = values[2]
		fallthrough
	case 2:
		cur.DefaultValue = values[1]
		fallthrough
	case 1:
		cur.Description = values[0]
	case 0:
		// NOOP
	default:
		return ppsNone, fmt.Errorf("expected no more than 3 colon-separated values, got %s", strings.Join(values, ", "))
	}

	skipComments(stream)
	if stream.CurrentToken().Key() == TokenEquals {
		stream.GoNext()
		skipComments(stream)
		if stream.CurrentToken().Key() == TokenBracketOpen {
			stream.GoNext()

			if cur.Type != PropertyTypeFlags && cur.Type != PropertyTypeChoices {
				return ppsNone, fmt.Errorf("non choices/flags property with choices list: %s", cur.Name)
			}

			return ppsInChoices, nil
		}
	}

	stream.GoPrev()

	return ppsEnd, nil
}

func skipComments(stream *tokenizer.Stream) bool {
	for stream.IsValid() {
		if stream.CurrentToken().Key() == tokenizer.TokenString &&
			stream.CurrentToken().StringKey() == TokenComment {
			stream.GoNext()
			continue
		}

		break
	}

	return stream.IsValid()
}

func advance(stream *tokenizer.Stream, count int) {
	for range count {
		if !stream.IsValid() {
			return
		}

		stream.GoNext()
	}
}
