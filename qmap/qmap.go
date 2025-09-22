// Package qmap parses Quake .map files in the Valve 220 format version 220.
// 'map' being a reserved go keyword, I added a q in there.
package qmap

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"unicode"
)

type QMap struct {
	entities []Entity
}

func LoadFromFile(path string) (QMap, error) {
	f, err := os.Open(path)
	if err != nil {
		return QMap{}, err
	}
	defer f.Close()

	return LoadFromReader(f)
}

func LoadFromReader(r io.Reader) (QMap, error) {
	parser := newParser(r)
	qmap, err := parser.run()
	if err != nil {
		return QMap{}, fmt.Errorf("unable to parse qmap: %w", err)
	}

	return qmap, nil
}

func (qmap *QMap) AddEntity(ent Entity) {
	qmap.entities = append(qmap.entities, ent)
}

func (qmap *QMap) RawEntities() []Entity {
	return qmap.entities
}

type Stats struct {
	NumEntities int
	NumProps    int
	NumBrushes  int
	NumPlanes   int
}

func (qmap *QMap) ComputeStats() Stats {
	var ret Stats

	for _, e := range qmap.entities {
		ret.NumEntities += 1
		ret.NumProps += len(e.props)

		for _, b := range e.brushes {
			ret.NumBrushes += 1
			ret.NumPlanes += len(b.planes)
		}
	}

	return ret
}

func (qmap *QMap) finalize() {
	for _, v := range qmap.entities {
		v.finalize()
	}
}

func (qmap *QMap) String() string {
	var b strings.Builder

	b.WriteString("// Game: Half-Life\n")
	b.WriteString("// Format: Valve\n")

	for i, v := range qmap.entities {
		fmt.Fprintf(&b, "// entity %d\n", i)
		b.WriteString(v.String())
	}

	return b.String()
}

// Marshalls structs into QMap entities.
// The qmap: field tags is of the form: property_name[,hardcoded_value].
func Marshal(in any) (string, error) {
	typ := reflect.TypeOf(in)
	numFields := typ.NumField()
	var out strings.Builder

	fmt.Fprint(&out, "{\n")

	for i := 0; i < numFields; i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		propName, propValue, ok := strings.Cut(field.Tag.Get("qmap"), ",")
		if !ok {
			propValue = reflect.ValueOf(in).Field(i).String()
		}

		if propName == "" {
			propName = toSnakeCase(typ.Field(i).Name)
		}

		if strings.Contains(propName, `"`) {
			return "", fmt.Errorf("property name cannot contain double-quotes: %s", propName)
		}

		if strings.Contains(propValue, `"`) {
			return "", fmt.Errorf("property value cannot contain double-quotes: %s", propValue)
		}

		fmt.Fprintf(&out, `"%s" "%s"`, propName, propValue)
		out.WriteRune('\n')
	}

	fmt.Fprint(&out, "}\n")

	return out.String(), nil
}

func toSnakeCase(in string) string {
	var b strings.Builder
	b.Grow(len(in))

	for i, c := range in {
		if unicode.IsUpper(c) && i != 0 {
			b.WriteRune('_')
		}

		b.WriteRune(unicode.ToLower(c))
	}

	return b.String()
}
