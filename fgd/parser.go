package fgd

import (
	"errors"
	"fmt"
	"image/color"
	"io"
	"strconv"

	"github.com/bzick/tokenizer"
)

const (
	TokenBracketClose tokenizer.TokenKey = iota + 1
	TokenBracketOpen
	TokenColon
	TokenComma
	TokenComment
	TokenEquals
	TokenParenClose
	TokenParenOpen
	TokenQuotedString
	TokenTypePrefix
	TokenMinus
)

func newTokenizer() *tokenizer.Tokenizer {
	ret := tokenizer.New()

	ret.DefineTokens(TokenTypePrefix, []string{"@"})
	ret.DefineTokens(TokenEquals, []string{"="})
	ret.DefineTokens(TokenParenOpen, []string{"("})
	ret.DefineTokens(TokenParenClose, []string{")"})
	ret.DefineTokens(TokenColon, []string{":"})
	ret.DefineTokens(TokenBracketOpen, []string{"["})
	ret.DefineTokens(TokenBracketClose, []string{"]"})
	ret.DefineTokens(TokenComma, []string{","})
	ret.DefineTokens(TokenMinus, []string{"-"})
	ret.DefineStringToken(TokenQuotedString, `"`, `"`)
	ret.DefineStringToken(TokenComment, "//", "\n")
	ret.AllowKeywordSymbols(tokenizer.Underscore, tokenizer.Numbers)

	return ret
}

type parserState int

const (
	psNone parserState = iota

	psOutside                // waiting for @
	psBeforeDefinitionType   // waiting for {Base,Point,Solid}Class
	psInDefinitionProperties // either "prop(value)"×n or "=" to end the list
	psBeforeEntityName
	psBeforeEntityDescription
	psBeforeEntityProperties // waiting for "["
	psInEntityProperties     // waiting for properties or "]" to terminate
)

type parseError struct {
	Err        error
	LineNumber int
	LineOffset int
}

func (e parseError) Error() string {
	return fmt.Sprintf("parse error at line %d, offset %d: %s", e.LineNumber, e.LineOffset, e.Err.Error())
}

func parse(r io.Reader) (FGD, error) {
	var (
		fgd   FGD
		cur   Definition
		state = psOutside
		tkn   = newTokenizer()
	)

	stream := tkn.ParseStream(r, 1024) // arbitrary
	defer stream.Close()

	for stream.IsValid() {
		if !skipComments(stream) {
			break
		}

		var err error
		token := stream.CurrentToken()

		switch state {
		case psOutside:
			state, err = parseOutside(token, &cur)
		case psBeforeDefinitionType:
			state, err = parseBeforeDefinitionType(token, &cur)
		case psInDefinitionProperties:
			state, err = parseInDefinitionProperties(stream, &cur)
		case psBeforeEntityName:
			state, err = parseBeforeEntityName(token, &cur)
		case psBeforeEntityDescription:
			state, err = parseBeforeEntityDescription(stream, &cur)
		case psBeforeEntityProperties:
			state, err = parseBeforeEntityProperties(token, &cur)
		case psInEntityProperties:
			state, err = parseInEntityProperties(stream, &cur)
			if err == nil && state == psOutside {
				fgd = append(fgd, cur)
			}
		case psNone:
			err = errors.New("parser reached an invalid state")
		}

		if err != nil {
			return nil, parseError{
				Err:        err,
				LineNumber: stream.CurrentToken().Line(),
				LineOffset: stream.CurrentToken().Offset(),
			}
		}

		stream.GoNext()
	}

	return fgd, nil
}

func parseOutside(token *tokenizer.Token, cur *Definition) (parserState, error) {
	if token.Key() != TokenTypePrefix {
		return psNone, fmt.Errorf("expected entity definition prefix \"@\", got '%s'", token.Value())
	}

	*cur = Definition{}

	return psBeforeDefinitionType, nil
}

func parseBeforeEntityProperties(token *tokenizer.Token, cur *Definition) (parserState, error) {
	if token.Key() != TokenBracketOpen {
		return psNone, fmt.Errorf("expected opening bracket before properties list, got '%s'", token.Value())
	}

	return psInEntityProperties, nil
}

func parseInEntityProperties(stream *tokenizer.Stream, cur *Definition) (parserState, error) {
	if stream.CurrentToken().Key() == TokenBracketClose {
		return psOutside, nil
	}

	prop, err := parseEntityProperty(stream)
	if err != nil {
		return psNone, err
	}
	cur.Properties = append(cur.Properties, prop)

	return psInEntityProperties, nil
}

func parseBeforeEntityDescription(stream *tokenizer.Stream, cur *Definition) (parserState, error) {
	token := stream.CurrentToken()
	// Description is optional, the properties list can start now.
	if token.Key() == TokenBracketOpen {
		return psInEntityProperties, nil
	}

	if token.Key() != TokenColon {
		return psNone, fmt.Errorf("expected colon before entity description, got '%s'", token.Value())
	}

	stream.GoNext()
	if !stream.IsValid() {
		return psNone, errors.New("unexpected EOF")
	}
	token = stream.CurrentToken()

	if token.StringKey() != TokenQuotedString {
		return psNone, fmt.Errorf("expected quoted entity description, got '%s'", token.Value())
	}
	cur.Description = token.ValueUnescapedString()

	return psBeforeEntityProperties, nil
}

func parseBeforeEntityName(token *tokenizer.Token, cur *Definition) (parserState, error) {
	if !token.IsKeyword() {
		return psNone, fmt.Errorf("expected entity name, got '%s'", token.Value())
	}

	cur.Name = token.ValueString()

	return psBeforeEntityDescription, nil
}

func parseInDefinitionProperties(stream *tokenizer.Stream, cur *Definition) (parserState, error) {
	for stream.CurrentToken().Key() != TokenEquals {
		name, err := parseValue(stream)
		if err != nil {
			return psNone, fmt.Errorf("unable to obtain entity definition option name: %w", err)
		}
		stream.GoNext()

		if stream.CurrentToken().Key() != TokenParenOpen {
			return psNone, fmt.Errorf("expected opening paren after option name, got '%s'", stream.CurrentToken().Value())
		}
		stream.GoNext()

		var optionValues []string
		for stream.CurrentToken().Key() != TokenParenClose {
			value, err := parseValue(stream)
			if err != nil {
				return psNone, fmt.Errorf("unable to parse entity definition option value: %w", err)
			}
			optionValues = append(optionValues, value)
			stream.GoNextIfNextIs(TokenComma)
			stream.GoNext()
		}
		stream.GoNext()

		if err := cur.setOption(name, optionValues); err != nil {
			return psNone, fmt.Errorf("unable to set entity definition option %s: %w", name, err)
		}
	}

	return psBeforeEntityName, nil
}

func parseBeforeDefinitionType(token *tokenizer.Token, cur *Definition) (parserState, error) {
	if !token.IsKeyword() {
		return psNone, fmt.Errorf("expected entity definition type, got '%s'", token.Value())
	}

	switch DefinitionType(token.ValueString()) {
	case DefinitionTypeBase:
		cur.Type = DefinitionTypeBase
	case DefinitionTypePoint:
		cur.Type = DefinitionTypePoint
	case DefinitionTypeSolid:
		cur.Type = DefinitionTypeSolid
	default:
		return psNone, fmt.Errorf("unknown entity type: %s", token.Value())
	}

	return psInDefinitionProperties, nil
}

func parseValue(stream *tokenizer.Stream) (string, error) {
	var negative bool
	if stream.CurrentToken().Key() == TokenMinus {
		if !stream.IsAnyNextSequence([]tokenizer.TokenKey{tokenizer.TokenInteger, tokenizer.TokenFloat}) {
			return "", fmt.Errorf("unexpected token after TokenMinus: '%s'", stream.NextToken().Value())
		}
		stream.GoNext()
		negative = true
	}

	token := stream.CurrentToken()
	switch token.Key() { //nolint: exhaustive // there's a default, duh
	case tokenizer.TokenInteger, tokenizer.TokenFloat:
		if negative {
			return "-" + token.ValueString(), nil
		}
		return token.ValueString(), nil
	case tokenizer.TokenKeyword:
		return token.ValueString(), nil
	case tokenizer.TokenString:
		if token.StringKey() != TokenQuotedString {
			return "", fmt.Errorf("expected a quoted string value, got '%s'", token.Value())
		}
		return token.ValueUnescapedString(), nil
	default:
		return "", fmt.Errorf("unexpected token, expected a valid key/value, got '%s'", token.Value())
	}
}

func parseColor(strR, strG, strB string) (color.NRGBA, error) {
	var zero color.NRGBA

	r, err := strconv.ParseUint(strR, 10, 8)
	if err != nil {
		return zero, fmt.Errorf("cannot parse R component: %w", err)
	}
	g, err := strconv.ParseUint(strG, 10, 8)
	if err != nil {
		return zero, fmt.Errorf("cannot parse G component: %w", err)
	}
	b, err := strconv.ParseUint(strB, 10, 8)
	if err != nil {
		return zero, fmt.Errorf("cannot parse B component: %w", err)
	}

	return color.NRGBA{uint8(r), uint8(g), uint8(b), 0xFF}, nil
}

func parseInts(in []string) ([]int, error) {
	ret := make([]int, len(in))
	errs := make([]error, len(in))

	for i := range in {
		ret[i], errs[i] = strconv.Atoi(in[i])
	}

	return ret, errors.Join(errs...)
}

func parseSize(strs []string) ([6]int, error) {
	ints, err := parseInts(strs)
	if err != nil {
		return [6]int{}, fmt.Errorf("unable to parse values as int: %w", err)
	}

	var ret [6]int

	switch len(ints) {
	case 3:
		for i := range 3 {
			ret[i] = -ints[i] / 2
			ret[i+3] = ints[i] / 2
		}
	case 6:
		copy(ret[:], ints)
	default:
		return [6]int{}, fmt.Errorf("expected 3 or 6 values, got %d", len(ints))
	}

	return ret, nil
}
