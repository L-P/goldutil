package qmap

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

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
	qm      QMap

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

func (p *parser) run() (QMap, error) {
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
			return QMap{}, ParseError{"reached an invalid state", curLineNumber, curLine}
		}

		if err != nil {
			return QMap{}, err
		}
	}

	if err := p.scanner.Err(); err != nil {
		return QMap{}, fmt.Errorf("unable to read file: %w", err)
	}

	if state != psOutside {
		return QMap{}, ParseError{"reached EOF before closing entity or brush", -1, ""}
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

		p.qm[index] = *p.curEntity
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

func parseProp(line string, lineNumber int) (string, string, error) {
	parts := strings.SplitN(strings.Trim(line, " \t"), " ", 2)
	if len(parts) != 2 {
		return "", "", ParseError{"unexpected property format", lineNumber, line}
	}

	key, err := parsePropertyString(parts[0])
	if err != nil {
		return "", "", ParseError{
			fmt.Sprintf("could not parse key: %s", err),
			lineNumber, line,
		}
	}

	value, err := parsePropertyString(parts[1])
	if err != nil {
		return "", "", ParseError{
			fmt.Sprintf("could not parse value: %s", err),
			lineNumber, line,
		}
	}

	return key, value, nil
}

func parsePropertyString(str string) (string, error) {
	if len(str) < 2 {
		return "", errors.New("too short (< 2 chars) to be a valid property string")
	}

	if !strings.HasPrefix(str, `"`) {
		return "", errors.New("not starting with double-quotes")
	}

	if !strings.HasSuffix(str, `"`) {
		return "", errors.New("not ending with double-quotes")
	}

	return str[1 : len(str)-1], nil
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
