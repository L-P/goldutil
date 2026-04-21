package sprite_test

import (
	"goldutil/goldsrc/sprite"
	"goldutil/palette"
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSprite(t *testing.T) {
	path := createTestImage(t, 256, 256)
	spr, err := sprite.NewFromImage(
		sprite.TypeParallelUpright,
		sprite.TextureFormatNormal,
		path,
	)
	require.NoError(t, err)
	require.NotNil(t, spr)

	expected := createTestPalette()
	for i := range 256 {
		require.Equal(t, palette.RGBFromColor(expected[i]), spr.Palette[i])
	}

	// I forgot those were used in the binary and shifted them once, ensure I
	// won't do it again.
	require.Equal(t, int32(0), int32(spr.Type))
	require.Equal(t, int32(0), int32(spr.TextureFormat))
	require.Len(t, spr.Frames, 1)
}

func createTestPalette() color.Palette {
	palette := make([]color.Color, 256)
	for i := range 256 {
		palette[i] = color.NRGBA{
			byte(i%13) * 16,
			byte(i%7) * 16,
			byte(i%11) * 16,
			0xFF,
		}
	}

	return palette
}

func createTestImage(t *testing.T, width, height int) *image.Paletted {
	t.Helper()

	palette := createTestPalette()
	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	for y := range img.Rect.Max.Y {
		for x := range img.Rect.Max.X {
			offset := img.PixOffset(x, y)
			img.Pix[offset] = byte(offset % 256)
		}
	}

	return img
}
