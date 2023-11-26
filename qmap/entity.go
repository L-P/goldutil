package qmap

import "fmt"

// Shared key names.
const (
	KAngles     = "angles"
	KClass      = "classname"
	KFlags      = "spawnflags"
	KKillTarget = "killtarget"
	KName       = "targetname"
	KOrigin     = "origin"
	KTarget     = "target"

	// Common but not strictly shared.
	KMaster = "master"

	// AI.
	KTriggerCondition = "TriggerCondition"
	KTriggerTarget    = "TriggerTarget"
)

type Entity struct {
	startLine, endLine int
	props              []Property
	brushes            []Brush
	keyLookup          map[string]string
}

func (e Entity) Name() string {
	name, ok := e.GetProperty(KName)
	if ok {
		return name
	}

	return fmt.Sprintf("__%s_L%d", e.Class(), e.startLine)
}

func (e Entity) Class() string {
	class, ok := e.GetProperty(KClass)
	if ok {
		return class
	}

	return "invalid"
}

// Properties defined last override properties defined first.
func (e *Entity) finalize() {
	e.keyLookup = make(map[string]string, len(e.props))
	for _, v := range e.props {
		e.keyLookup[v.key] = v.value
	}
}

func (e *Entity) GetProperty(key string) (string, bool) {
	v, ok := e.keyLookup[key]
	return v, ok
}

func (e *Entity) PropertyMap() map[string]string {
	return e.keyLookup
}

func (e Entity) Brushes() []Brush {
	return e.brushes
}

func (e *Entity) addProperty(p Property) {
	e.props = append(e.props, p)
}

func (e *Entity) addBrush(b Brush) {
	e.brushes = append(e.brushes, b)
}

type Property struct {
	line       int
	key, value string
}

type Brush struct {
	startLine, endLine int
	planes             []string // raw planes, unparsed
}

func (b *Brush) addPlane(plane string) {
	b.planes = append(b.planes, plane)
}
