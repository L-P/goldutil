// Package typed_map parses Quake .map files in the Valve 220 format as struct
// containers.
package qmap

import (
	"fmt"
	"io"
	"iter"
	"os"
	"strings"

	"github.com/google/uuid"
)

// QMap holds the entities of a .map file. While it essentially is an array
// of entities, it is stored into a UUID-keyed map to allow CRUD operations
// without having to deal with moving indexes.
// Because we have lots of "keys" and "values" floating around, by convention
// variables containing keys of this map are called "index" (not "i", not "k").
// KVs being stored as maps, obtaining an entity and updating its KVs will
// update the contents of the QMap.
type QMap struct {
	entities map[uuid.UUID]AnonymousEntity
	order    []uuid.UUID
}

// Returns an iterator over the entities in their original order.
func (qm *QMap) Entities() iter.Seq[AnonymousEntity] {
	return func(yield func(AnonymousEntity) bool) {
		for _, index := range qm.order {
			if ent, ok := qm.entities[index]; ok {
				if !yield(ent) {
					return
				}
			}
		}
	}
}

func New() *QMap {
	return &QMap{
		entities: make(map[uuid.UUID]AnonymousEntity),
	}
}

type BrushEntity struct {
	Brushes []Brush
}

func (ent *BrushEntity) AddBrush(brush Brush) {
	ent.Brushes = append(ent.Brushes, brush)
}

type AnonymousEntity struct {
	BrushEntity

	KVs map[string]string
}

func NewAnonymousEntity() AnonymousEntity {
	return AnonymousEntity{KVs: make(map[string]string)}
}

func (ent *AnonymousEntity) Clear() {
	*ent = AnonymousEntity{}
}

func (ent *AnonymousEntity) IsZero() bool {
	return ent.KVs == nil
}

type Brush []string // raw planes, unparsed

func LoadFromFile(path string) (*QMap, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck // readonly

	return LoadFromReader(f)
}

func LoadFromReader(r io.Reader) (*QMap, error) {
	parser := newParser(r)
	qm, err := parser.run()
	if err != nil {
		return nil, fmt.Errorf("unable to parse qmap: %w", err)
	}

	return qm, nil
}

func (ent *AnonymousEntity) String() string {
	var b strings.Builder

	b.WriteString("{\n")
	for k, v := range ent.KVs {
		fmt.Fprintf(&b, `"%s" "%s"`, k, v)
		b.WriteRune('\n')
	}

	for i, brush := range ent.Brushes {
		fmt.Fprintf(&b, "// brush %d\n", i)
		b.WriteString("{\n")
		for _, v := range brush {
			b.WriteString(v)
			b.WriteRune('\n')
		}
		b.WriteString("}\n")
	}

	b.WriteString("}\n")

	return b.String()
}

func (qm *QMap) String() string {
	var b strings.Builder

	b.WriteString("// Game: Half-Life\n")
	b.WriteString("// Format: Valve\n")

	var i int
	for ent := range qm.Entities() { // ensure we keep the original order
		fmt.Fprintf(&b, "// entity %d\n", i)
		b.WriteString(ent.String())
		i++
	}

	return b.String()
}

func (qm *QMap) AddEntities(ents []any) error {
	for _, v := range ents {
		index, err := uuid.NewRandom()
		if err != nil {
			return fmt.Errorf("unable to generate UUID as entity index: %w", err)
		}

		anon, err := NewAnonymousEntityFromStruct(v)
		if err != nil {
			return fmt.Errorf("unable to convert back to AnonymousEntity: %w", err)
		}

		qm.entities[index] = anon
		qm.order = append(qm.order, index)
	}

	return nil
}

func (qm *QMap) AddAnonymousEntities(ents ...AnonymousEntity) error {
	for _, ent := range ents {
		index, err := uuid.NewRandom()
		if err != nil {
			return fmt.Errorf("unable to generate UUID as entity index: %w", err)
		}

		qm.entities[index] = ent
		qm.order = append(qm.order, index)
	}

	return nil
}

func (qm *QMap) Delete(index uuid.UUID) {
	delete(qm.entities, index)
}
