package palette_test

import (
	"goldutil/palette"
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaletteSize(t *testing.T) {
	require.Equal(t, 256, palette.Size, "this is hardcoded in engine and will never change")
}

func TestAsColorPalette(t *testing.T) {
	expected := make(color.Palette, 256)
	for i := range expected {
		expected[i] = color.NRGBA{uint8(i), 0, 0, 0xFF}
	}

	var input palette.Palette
	for i := range input {
		input[i] = palette.RGB{uint8(i), 0, 0}
	}

	actual := input.AsColorPalette()
	require.Equal(t, expected, actual)
}

func TestFromImageSmallPalette(t *testing.T) {
	img := createTestImage(t, 256, 256, createSmallTestPalette())
	pal, lastIndex, shouldRemap, err := palette.FromImage(img)
	require.NoError(t, err)
	require.Equal(t, uint8(127), lastIndex)
	require.True(t, shouldRemap)

	for i := range lastIndex + 1 {
		//nolint:gosec // nope, not out of bounds
		require.Equal(t, palette.RGB{i, 0, 0}, pal[i])
	}

	require.Contains(t, img.Pix, lastIndex)
	palette.RemapLastColor(img, lastIndex)
	require.NotContains(t, img.Pix, lastIndex)
}

func createSmallTestPalette() color.Palette {
	palette := make(color.Palette, 128)
	for i := range palette {
		palette[i] = color.NRGBA{uint8(i), 0, 0, 0xFF}
	}

	return palette
}

func createTestImage(t *testing.T, width, height int, pal color.Palette) *image.Paletted {
	t.Helper()

	img := image.NewPaletted(image.Rect(0, 0, width, height), pal)
	for y := range img.Rect.Max.Y {
		for x := range img.Rect.Max.X {
			offset := img.PixOffset(x, y)
			img.Pix[offset] = byte(offset % len(pal))
		}
	}

	return img
}
