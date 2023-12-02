// Package qmap parses Quake .map files in the Valve 220 format version 220.
// 'map' being a reserved go keyword, I added a q in there.
package qmap

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type QMap struct {
	entities []Entity

	targetNameLookup map[string]int
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
	qmap.targetNameLookup = make(map[string]int, len(qmap.entities))

	for i, v := range qmap.entities {
		v.finalize()
		qmap.entities[i] = v

		if targetName, ok := v.GetProperty(KName); ok {
			qmap.targetNameLookup[targetName] = i
		}
	}
}

func (qmap *QMap) GetEntityByName(name string) (Entity, bool) {
	i, ok := qmap.targetNameLookup[name]
	if !ok {
		return Entity{}, false
	}

	return qmap.entities[i], true
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
