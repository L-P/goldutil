// Package palette contains image and palette manipulation utilities.
package palette

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"
)

type RGB struct {
	R, G, B uint8
}

func RGBFromColor(c color.Color) RGB {
	r, g, b, _ := c.RGBA()
	return RGB{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
	}
}

const Size = 256

type Palette [Size]RGB

func (p Palette) AsColorPalette() color.Palette {
	ret := make([]color.Color, len(p))
	for i := range p {
		ret[i] = color.NRGBA{
			p[i].R,
			p[i].G,
			p[i].B,
			0xFF,
		}
	}

	return ret
}

func OpenPalettedImage(path string) (*image.Paletted, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open file at '%s': %w", path, err)
	}
	defer f.Close() //nolint:errcheck // readonly

	mysteryImg, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("unable to decode image: %w", err)
	}

	img, ok := mysteryImg.(*image.Paletted)
	if !ok {
		return nil, errors.New("image is not paletted")
	}

	return img, nil
}

/* GoldSrc uses the last color of a palette to indicate what to use as the
* transparent color for masked transparency. If the palette is shorter than the
* maximum the last color of the palette needs to be moved to the last spot and
* all corresponding pixels also need to be updated to this new palette index. */
func RemapLastColor(img *image.Paletted, remapIndex uint8) {
	for i, v := range img.Pix {
		if v == remapIndex {
			img.Pix[i] = 0xFF
		}
	}
}

// Returns the final palette, the last index in the input palette, and true if
// this index must be remapped to 0xFF.
func FromImage(img *image.Paletted) (Palette, uint8, bool, error) {
	var palette Palette
	if len(img.Palette) > 256 {
		return palette, 0, false, fmt.Errorf("expected at most 256 colors palette, got %d", len(palette))
	}

	for i, v := range img.Palette {
		r, g, b, _ := v.RGBA()
		palette[i] = RGB{
			R: uint8(r),
			G: uint8(g),
			B: uint8(b),
		}
	}

	lastIndex := len(img.Palette) - 1
	if lastIndex != 255 {
		palette[255] = palette[lastIndex]
		return palette, uint8(lastIndex), true, nil
	}

	return palette, uint8(lastIndex), false, nil
}
