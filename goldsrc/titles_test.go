//nolint:gofmt // BUG
package goldsrc_test

import (
	"goldutil/goldsrc"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTitles(t *testing.T) {
	input := `
// Some comment.

// Initial state.
$position 0.05 0.85
$effect 0
$color 255 255 255
$color2 255 255 255
$fadein 0.1
$fxtime 0
$holdtime 4
$fadeout 0.25

foo
{
This is good content.
With a break.
}

$holdtime 12
bar
{
This should only differ by name, message, and holdtime.
}
`

	parsed, err := goldsrc.NewTitlesFromReader(strings.NewReader(input))
	require.NoError(t, err)

	expected := map[string]goldsrc.Title{
		"foo": goldsrc.Title{
			Name:           "foo",
			Message:        "This is good content.\nWith a break.",
			Position:       "0.05 0.85",
			Effect:         goldsrc.TitleEffectFade,
			TextColor:      "255 255 255",
			HighlightColor: "255 255 255",
			FadeIn:         0.1,
			FadeOut:        0.25,
			FXTime:         0,
			HoldTime:       4,
		},
		"bar": goldsrc.Title{
			Name:           "bar",
			Message:        "This should only differ by name, message, and holdtime.",
			Position:       "0.05 0.85",
			Effect:         goldsrc.TitleEffectFade,
			TextColor:      "255 255 255",
			HighlightColor: "255 255 255",
			FadeIn:         0.1,
			FadeOut:        0.25,
			FXTime:         0,
			HoldTime:       12,
		},
	}

	require.Equal(t, expected, parsed)
}
