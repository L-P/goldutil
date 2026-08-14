package fgd_test

import (
	"strings"
	"testing"

	"github.com/L-P/goldutil/fgd"
	"github.com/L-P/goldutil/neat"
	"github.com/stretchr/testify/require"
)

func TestFGDParser(t *testing.T) {
	actual, err := fgd.NewFromReader(strings.NewReader(input))
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestDogfood(t *testing.T) {
	_, err := fgd.NewFromReader(strings.NewReader(neat.FGD))
	require.NoError(t, err)
}

var input = `
// Comment.
@PointClass = test_point_entity : "Test point entity"
[
    // Comment with leading whitespace.
    targetname(target_source) : "Name"
    target(target_destination) : "Target"
]

@SolidClass = test_brush_entity : "Test brush entity" // Comment after a line.
[
    targetname(target_source) : "Name"
    target(target_destination) : "Target"
]
`

var expected = fgd.FGD{
	fgd.Definition{
		Name:        "test_point_entity",
		Type:        fgd.DefinitionTypePoint,
		Description: "Test point entity",
		Properties: []fgd.Property{
			fgd.Property{
				Name:        "targetname",
				Description: "Name",
				Type:        fgd.PropertyTypeTargetSource,
			},
			fgd.Property{
				Name:        "target",
				Description: "Target",
				Type:        fgd.PropertyTypeTargetDestination,
			},
		},
	},
	fgd.Definition{
		Name:        "test_brush_entity",
		Type:        fgd.DefinitionTypeSolid,
		Description: "Test brush entity",
		Properties: []fgd.Property{
			fgd.Property{
				Name:        "targetname",
				Description: "Name",
				Type:        fgd.PropertyTypeTargetSource,
			},
			fgd.Property{
				Name:        "target",
				Description: "Target",
				Type:        fgd.PropertyTypeTargetDestination,
			},
		},
	},
}
