package neat_test

import (
	"embed"
	"goldutil/goldsrc/typedmap"
	"goldutil/neat"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed test_cases/*.map
var cases embed.FS

func TestNeatify(t *testing.T) {
	inputs, err := fs.Glob(cases, "test_cases/*.input.map")
	require.NoError(t, err)
	require.NotEmpty(t, inputs)

	for _, inputPath := range inputs {
		expectedPath, _ := strings.CutSuffix(inputPath, ".input.map")
		expectedPath += ".expected.map"

		t.Run(inputPath, func(t *testing.T) {
			input, err := cases.Open(inputPath)
			require.NoError(t, err)
			tmap, err := typedmap.LoadFromReader(input)
			require.NoError(t, err)

			require.NoError(t, neat.Neatify(tmap))

			expected, err := cases.Open(expectedPath)
			require.NoError(t, err)
			expectedTMap, err := typedmap.LoadFromReader(expected)
			require.NoError(t, err)

			require.ElementsMatch(
				t,
				slices.Collect(maps.Values(expectedTMap)),
				slices.Collect(maps.Values(tmap)),
			)
		})
	}
}
