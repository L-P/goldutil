package qmap

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

type parserState int

const (
	psNone parserState = iota
	psOutside
	psInEntity
	psInBrush
)

type parser struct {
	scanner *bufio.Scanner
	state   parserState
	qm      *QMap

	curEntity *AnonymousEntity
	curBrush  Brush
}

func newParser(r io.Reader) parser {
	return parser{
		state:   psOutside,
		scanner: bufio.NewScanner(r),
		qm:      New(),
	}
}

func (p *parser) run() (*QMap, error) {
	var (
		curLineNumber int
		state         = psOutside
	)

	for p.scanner.Scan() {
		curLineNumber++
		curLine := p.scanner.Text()
		if strings.HasPrefix(curLine, "//") {
			continue
		}

		var err error
		switch state {
		case psOutside:
			state, err = p.parseOutside(curLine, curLineNumber)
		case psInEntity:
			state, err = p.parseEntity(curLine, curLineNumber)
		case psInBrush:
			state = p.parseBrush(curLine, curLineNumber)
		case psNone:
			return nil, ParseError{"reached an invalid state", curLineNumber, curLine}
		}

		if err != nil {
			return nil, err
		}
	}

	if err := p.scanner.Err(); err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}

	if state != psOutside {
		return nil, ParseError{"reached EOF before closing entity or brush", -1, ""}
	}

	return p.qm, nil
}

func (p *parser) parseOutside(line string, lineNumber int) (parserState, error) {
	if line != "{" {
		return psNone, ParseError{"expected start of entity", lineNumber, line}
	}

	newEnt := NewAnonymousEntity()
	p.curEntity = &newEnt

	return psInEntity, nil
}

func (p *parser) parseEntity(line string, lineNumber int) (parserState, error) {
	switch line {
	case "}":
		index, err := uuid.NewRandom()
		if err != nil {
			return psNone, fmt.Errorf("unable to generate UUID as entity index: %w", err)
		}

		p.qm.entities[index] = *p.curEntity
		p.qm.order = append(p.qm.order, index)
		p.curEntity = nil

		return psOutside, nil

	case "{":
		p.curBrush = Brush{}
		return psInBrush, nil
	}

	pKey, pValue, err := parseProp(line, lineNumber)
	if err != nil {
		return psNone, err
	}

	// Only keep last value.
	// TODO: Double-check that it's what the engine does.
	p.curEntity.KVs[pKey] = pValue

	return psInEntity, nil
}

// Splits a raw line containing a property into its key and value.
// There's no escaping double-quotes and a property cannot span multiple lines.
func parseProp(line string, lineNumber int) (string, string, error) {
	parts := make([]string, 0, 2)
	var inString bool
	var cur string

	for _, c := range line {
		if !inString {
			if unicode.IsSpace(c) {
				continue
			}

			if c != '"' {
				return "", "", ParseError{fmt.Sprintf("expected \", got: %c", c), lineNumber, line}
			}
			inString = true
			continue
		}

		if c == '"' {
			inString = false
			parts = append(parts, cur)
			cur = ""
			continue
		}

		cur += string(c)
	}

	if inString {
		return "", "", ParseError{"missing terminating double-quote", lineNumber, line}
	}

	if len(parts) != 2 {
		return "", "", ParseError{"too many string tokens", lineNumber, line}
	}

	return parts[0], parts[1], nil
}

func (p *parser) parseBrush(line string, _lineNumber int) parserState {
	if line == "}" {
		p.curEntity.AddBrush(p.curBrush)
		p.curBrush = nil

		return psInEntity
	}

	p.curBrush = append(p.curBrush, line)

	return psInBrush
}

type ParseError struct {
	message      string
	lineNumber   int
	lineContents string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("parse error on line #%d: %s", e.lineNumber, e.message)
}
