package fgd_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/L-P/goldutil/fgd"
	"github.com/L-P/goldutil/neat"
	"github.com/stretchr/testify/require"
)

func TestFGDParser(t *testing.T) {
	for i, v := range cases {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			actual, err := fgd.NewFromReader(strings.NewReader(v.input))
			require.NoError(t, err)
			require.Equal(t, v.expected, actual)
		})
	}
}

func TestDogfood(t *testing.T) {
	_, err := fgd.NewFromReader(strings.NewReader(neat.FGD))
	require.NoError(t, err)
}

type testCase struct {
	input    string
	expected fgd.FGD
	err      error
}

var cases = []testCase{
	{
		``, nil, nil,
	},
	{
		`
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
`,

		fgd.FGD{
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
		},
		nil,
	},
}
