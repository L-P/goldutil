package goldsrc

import (
	"bufio"
	"fmt"
	"goldutil/set"
	"io"
	"log"
	"strconv"
	"strings"
)

type TitleEffect int

const (
	TitleEffectFade TitleEffect = iota
	TitleEffectFlicker
	TitleEffectTypewriter
)

// A single entry from titles.txt.
type Title struct {
	Name    string
	Message string

	// I won't bother parsing position/colors for now. Don't need them.
	Position       string      // $position "0.0 0.0"
	Effect         TitleEffect // $effect 0/1/2
	TextColor      string      // $color "255 255 255"
	HighlightColor string      // $color2 "255 255 255"
	FadeIn         float32     // $fadein
	FadeOut        float32     // $fadeout
	FXTime         float32     // $fxtime
	HoldTime       float32     // $holdtime
}

func NewTitlesFromReader(r io.Reader) ([]Title, error) {
	parser := newTitlesParser(r)
	return parser.run()
}

type titlesParserState int

const (
	tpsNone titlesParserState = iota
	tpsOutside
	tpsInMessage
)

type titlesParser struct {
	scanner      *bufio.Scanner
	currentTitle Title
	output       []Title
	names        set.PresenceSet[string]
}

func newTitlesParser(r io.Reader) titlesParser {
	return titlesParser{
		scanner: bufio.NewScanner(r),
		names:   set.NewPresenceSet[string](0),
	}
}

func (parser *titlesParser) run() ([]Title, error) {
	var curLineNumber int
	state := tpsOutside

	for parser.scanner.Scan() {
		curLineNumber++
		curLine := parser.scanner.Text()
		if strings.HasPrefix(curLine, "//") {
			continue
		}

		var err error
		switch state {
		case tpsOutside:
			state, err = parser.parseOutside(curLine, curLineNumber)
		case tpsInMessage:
			state, err = parser.parseMessage(curLine, curLineNumber)
		case tpsNone:
			return nil, ParseError{"reached an invalid state", curLineNumber, curLine}
		}

		if err != nil {
			return nil, err
		}
	}

	return parser.output, nil
}

func (parser *titlesParser) parseOutside(line string, lineNumber int) (titlesParserState, error) {
	if strings.TrimSpace(line) == "" {
		return tpsOutside, nil
	}

	if line == "{" {
		return tpsInMessage, nil
	}

	if line[0] == '$' {
		return tpsOutside, parser.parseParameter(line, lineNumber)
	}

	parser.currentTitle.Name = line

	return tpsOutside, nil
}

func (parser *titlesParser) parseMessage(line string, lineNumber int) (titlesParserState, error) {
	if line == "}" {
		parser.output = append(parser.output, parser.currentTitle)
		parser.currentTitle.Name = ""
		parser.currentTitle.Message = ""
		return tpsOutside, nil
	}

	if parser.currentTitle.Message != "" {
		parser.currentTitle.Message += "\n"
	}

	parser.currentTitle.Message += line
	if len(line) > 256 {
		return tpsNone, ParseError{"message exceeds 256 chars", lineNumber, line}
	}

	return tpsInMessage, nil
}

func (parser *titlesParser) parseParameter(line string, lineNumber int) error {
	key, value, hasValue := strings.Cut(line, " ")
	if !hasValue {
		return ParseError{"parameter has no value", lineNumber, line}
	}

	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	log.Printf("'%s' '%s'", key, value)

	switch key {
	case "$position":
		parser.currentTitle.Position = value
		return nil
	case "$effect":
		fx, err := strconv.Atoi(value)
		if err != nil {
			return ParseError{err.Error(), lineNumber, line}
		}
		parser.currentTitle.Effect = TitleEffect(fx)
		return nil
	case "$color":
		parser.currentTitle.TextColor = value
		return nil
	case "$color2":
		parser.currentTitle.HighlightColor = value
		return nil
	}

	floatValue, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return ParseError{err.Error(), lineNumber, line}
	}

	switch key {
	case "$fadein":
		parser.currentTitle.FadeIn = float32(floatValue)
	case "$fxtime":
		parser.currentTitle.FXTime = float32(floatValue)
	case "$holdtime":
		parser.currentTitle.HoldTime = float32(floatValue)
	case "$fadeout":
		parser.currentTitle.FadeOut = float32(floatValue)
	default:
		return fmt.Errorf("unknown parameter: %s", line)
	}

	return nil
}

type ParseError struct {
	message      string
	lineNumber   int
	lineContents string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("parse error on line #%d: %s", e.lineNumber, e.message)
}
