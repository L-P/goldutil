package fgd

import (
	"errors"
	"fmt"
	"image/color"
	"io"
	"strings"
)

type DefinitionType string

const (
	DefinitionTypeBase  DefinitionType = "BaseClass"
	DefinitionTypePoint DefinitionType = "PointClass"
	DefinitionTypeSolid DefinitionType = "SolidClass"
)

type PropertyType string

const (
	PropertyTypeAudio             PropertyType = "sound"    // path, stored as string
	PropertyTypeChoices           PropertyType = "choices"  // read as string or integer
	PropertyTypeColor255          PropertyType = "color255" // stored as string
	PropertyTypeDecal             PropertyType = "decal"    // texture name, stored as string
	PropertyTypeFlags             PropertyType = "flags"    // stored as integer
	PropertyTypeFloat             PropertyType = "float"    // stored as string
	PropertyTypeInteger           PropertyType = "integer"
	PropertyTypeSprite            PropertyType = "sprite" // path, stored as string
	PropertyTypeString            PropertyType = "string"
	PropertyTypeStudio            PropertyType = "studio"             // path, stored as string
	PropertyTypeTargetDestination PropertyType = "target_destination" // stored as string
	PropertyTypeTargetSource      PropertyType = "target_source"      // stored as string
)

type Property struct {
	Name               string
	Description        string
	ExtraDocumentation string
	Type               PropertyType

	// Ultimately all values are encoded as strings in the final .map, so we
	// lose nothing by only storing strings ourselves.
	// Still, for integer types the writer should ensure that the value is not
	// quoted, other parsers can be finicky.
	DefaultValue string

	// For Flags and Choices.
	Choices []ChoiceEntry
}

type ChoiceEntry struct {
	Value              string
	Name               string
	DefaultValue       string
	ExtraDocumentation string
}

type Definition struct {
	Name        string
	Type        DefinitionType
	Description string
	Bases       []string
	Size        [6]int // minX, minY, minZ, maxX, maxY, maxZ
	Offset      [3]int
	Color       color.NRGBA
	Flags       []string // only ever used in sven-coop with Light or Path

	Body       string
	Sequence   string
	IconSprite string

	HasDecal, HasSprite, HasStudio bool

	Properties []Property
}

//nolint:funlen // big ass-switch, there's nothing better
func (def *Definition) setOption(name string, values []string) error {
	switch strings.ToLower(name) {
	case "studio":
		def.HasStudio = true
	case "sprite":
		def.HasSprite = true
	case "decal":
		def.HasDecal = true
	case "iconsprite":
		if len(values) != 1 {
			return errors.New("iconsprite() requires exactly one value")
		}
		def.IconSprite = values[0]
	case "base":
		def.Bases = values
	case "flags":
		def.Flags = values
	case "color":
		if len(values) != 3 {
			return errors.New("color() requires exactly 3 values")
		}

		color, err := parseColor(values[0], values[1], values[2])
		if err != nil {
			return fmt.Errorf("cannot parse color(): %w", err)
		}

		def.Color = color
	case "sequence":
		if len(values) != 1 {
			return errors.New("sequence() requires exactly one value")
		}

		def.Sequence = values[0]
	case "body":
		if len(values) != 1 {
			return errors.New("body() requires exactly one value")
		}

		def.Body = values[0]
	case "offset":
		if len(values) != 3 {
			return errors.New("offset() requires exactly 3 values")
		}

		ints, err := parseInts(values)
		if err != nil {
			return fmt.Errorf("unable to parse offset(): %w", err)
		}

		copy(def.Offset[:], ints)
	case "size":
		size, err := parseSize(values)
		if err != nil {
			return fmt.Errorf("unable to parse size(): %w", err)
		}
		def.Size = size
	default:
		return fmt.Errorf("unhandled entity option: %s", name)
	}

	return nil
}

type FGD []Definition

func NewFromReader(r io.Reader) (FGD, error) {
	fgd, err := parse(r)
	if err != nil {
		return nil, fmt.Errorf("unable to parse FGD: %w", err)
	}

	return fgd, nil
}
