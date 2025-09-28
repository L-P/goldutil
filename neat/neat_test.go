package neat_test

import (
	"embed"
	"goldutil/goldsrc/qmap"
	"goldutil/neat"
	"io/fs"
	"os"
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
			qm, err := qmap.LoadFromReader(input)
			require.NoError(t, err)

			mod, err := os.OpenRoot("test_cases")
			require.NoError(t, err)

			require.NoError(t, neat.Neatify(qm, mod))

			expected, err := cases.Open(expectedPath)
			require.NoError(t, err)
			expectedQM, err := qmap.LoadFromReader(expected)
			require.NoError(t, err)

			require.ElementsMatch(
				t,
				slices.Collect(expectedQM.Entities()),
				slices.Collect(qm.Entities()),
			)
		})
	}
}
