package wad

import (
	"image"
	"image/color"
)

func (mip MIPTexture) Render(mipIndex int) (image.Image, error) {
	image := image.NewPaletted(mip.Rect(mipIndex), mip.PaletteNRGBA())
	image.Pix = mip.MIPData[mipIndex]

	return image, nil
}

// TODO merge with Sprite render code.
func (mip MIPTexture) PaletteNRGBA() color.Palette {
	palette := make([]color.Color, len(mip.Palette))
	for i := 0; i < len(mip.Palette); i++ {
		palette[i] = color.NRGBA{
			mip.Palette[i].R,
			mip.Palette[i].G,
			mip.Palette[i].B,
			0xFF,
		}
	}

	return palette
}

// Returns the mipmap size divisor.
func mipIndexToScale(i int) int {
	return (2 << i) / 2
}

func (mip MIPTexture) Rect(mipIndex int) image.Rectangle {
	div := mipIndexToScale(mipIndex)
	return image.Rect(0, 0, int(mip.Width)/div, int(mip.Height)/div)
}
