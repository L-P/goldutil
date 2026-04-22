package wad

import (
	"image"
	"image/color"
	"strings"
)

func (mip *MIPTexture) Render(withAlpha bool) (image.Image, error) {
	return mip.RenderMipmap(0, withAlpha)
}

func (mip *MIPTexture) RenderMipmap(mipIndex int, withAlpha bool) (image.Image, error) {
	image := image.NewPaletted(mip.Rect(mipIndex), mip.PaletteNRGBA())
	image.Pix = mip.MIPData[mipIndex]

	if withAlpha && strings.HasPrefix(mip.Name.String(), "{") {
		r, g, b, _ := image.Palette[0xFF].RGBA()
		image.Palette[0xFF] = color.NRGBA{
			uint8(r),
			uint8(g),
			uint8(b),
			0x00,
		}
	}

	return image, nil
}

func (mip *MIPTexture) PaletteNRGBA() color.Palette {
	palette := make(color.Palette, len(mip.Palette))
	for i := range len(mip.Palette) {
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

func (mip *MIPTexture) Rect(mipIndex int) image.Rectangle {
	div := mipIndexToScale(mipIndex)
	return image.Rect(0, 0, int(mip.Width)/div, int(mip.Height)/div)
}
