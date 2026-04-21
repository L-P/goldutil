package sprite

import (
	"fmt"
	"image"
	"image/color"
)

func (spr *Sprite) RenderFrame(i int, withAlpha bool) (image.Image, error) {
	if i < 0 || i > len(spr.Frames) {
		return nil, fmt.Errorf("frame index %d is out of bounds [0-%d]", i, len(spr.Frames))
	}

	frame := spr.Frames[i]
	var palette color.Palette
	if withAlpha {
		palette = spr.PaletteNRGBA()
	} else {
		palette = spr.Palette.AsColorPalette()
	}

	image := image.NewPaletted(frame.Rect(), palette)
	image.Pix = frame.Data

	return image, nil
}

func (spr *Sprite) PaletteNRGBA() color.Palette {
	if spr.TextureFormat == TextureFormatIndexAlpha {
		return spr.indexAlphaPaletteNRGBA()
	}

	palette := make([]color.Color, spr.PaletteSize)
	for i := range spr.PaletteSize {
		palette[i] = color.NRGBA{
			spr.Palette[i].R,
			spr.Palette[i].G,
			spr.Palette[i].B,
			0xFF,
		}
	}
	if spr.TextureFormat == TextureFormatAlphaTest {
		palette[len(palette)-1] = color.NRGBA{0, 0, 0, 0}
	}

	return palette
}

// Last index is used as the color, the rest is actually the color defined at
// the last index + the current index as the alpha value.
func (spr *Sprite) indexAlphaPaletteNRGBA() color.Palette {
	palette := make([]color.Color, spr.PaletteSize)

	for i := range spr.PaletteSize {
		palette[i] = color.NRGBA{
			spr.Palette[0xFF].R,
			spr.Palette[0xFF].G,
			spr.Palette[0xFF].B,
			uint8(i),
		}
	}

	return palette
}
