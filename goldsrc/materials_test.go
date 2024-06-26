package goldsrc_test

import (
	"goldutil/goldsrc"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaterials(t *testing.T) {
	var (
		src = `
// Starting with a comment, then an empty line.

D TEX_DIRT
G TEX_GRATE
M TEX_METAL

// Another set of garbage for good measure.

P TEX_COMPUTER
S TEX_LIQUID
T TEX_TILE
V TEX_VENTS
W TEX_WOOD
Y TEX_GLASS
`
		expected = goldsrc.Materials(map[string]goldsrc.MaterialType{
			"TEX_DIRT":     'D',
			"TEX_GRATE":    'G',
			"TEX_METAL":    'M',
			"TEX_COMPUTER": 'P',
			"TEX_LIQUID":   'S',
			"TEX_TILE":     'T',
			"TEX_VENTS":    'V',
			"TEX_WOOD":     'W',
			"TEX_GLASS":    'Y',
		})
	)

	actual, err := goldsrc.LoadMaterials(strings.NewReader(src))
	require.NoError(t, err)

	require.Equal(t, expected, actual)
}

func TestInvalidMaterials(t *testing.T) {
	var cases = []string{
		"D TEX_DIRT // line with garbage, ie. this comment",
		"A TEX_DIRT",                // invalid type
		"D TEXTURENAME_IS_TOO_LONG", // texture name is too long
		"D badchars",
		"D TEX_DIRT garbage",
		" D TEX_DIRT", // space before
		"D  TEX_DIRT", // two spaces
	}

	for _, v := range cases {
		_, err := goldsrc.LoadMaterials(strings.NewReader(v))
		require.Error(t, err, "should be an error: %s", v)
	}
}
