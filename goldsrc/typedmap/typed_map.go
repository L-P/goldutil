// Package typed_map parses Quake .map files in the Valve 220 format version
// 220 as struct containers.
package typedmap

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type TypedMap []any

type BrushEntity struct {
	brushes []Brush
}

func (ent *BrushEntity) AddBrush(brush Brush) {
	ent.brushes = append(ent.brushes, brush)
}

type AnonymousEntity struct {
	BrushEntity
	kvs map[string]string
}

func NewAnonymousEntity(kvs map[string]string) *AnonymousEntity {
	if kvs == nil {
		kvs = make(map[string]string)
	}

	return &AnonymousEntity{kvs: kvs}
}

type Brush []string // raw planes, unparsed

func LoadFromFile(path string, types []any) (TypedMap, error) {
	f, err := os.Open(path)
	if err != nil {
		return TypedMap{}, err
	}
	defer f.Close()

	return LoadFromReader(f, types)
}

func LoadFromReader(r io.Reader, types []any) (TypedMap, error) {
	parser := newParser(r, types)
	tmap, err := parser.run()
	if err != nil {
		return TypedMap{}, fmt.Errorf("unable to parse qmap: %w", err)
	}

	return tmap, nil
}

func (tmap *TypedMap) String() string {
	var b strings.Builder

	b.WriteString("// Game: Half-Life\n")
	b.WriteString("// Format: Valve\n")

	for i, v := range *tmap {
		fmt.Fprintf(&b, "// entity %d\n", i)
		data, err := Marshal(v)
		if err != nil {
			panic(err)
		}
		b.Write(data)
		fmt.Fprintf(&b, "%#v\n", v)
	}

	return b.String()
}
