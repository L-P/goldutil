package main

import (
	"errors"
	"fmt"
	"goldutil/sprite"
	"image"
	"os"
)

/* GoldSrc uses the last color of a palette to indicate what to use as the
* transparent color for masked transparency. If the palette is shorter than the
* maximum the last color of the palette needs to be moved to the last spot and
* all corresponding pixels also need to be updated to this new palette index. */
func remapLastColor(img *image.Paletted, remapIndex uint8) {
	for i, v := range img.Pix {
		if v == remapIndex {
			img.Pix[i] = 0xFF
		}
	}
}

func openPalettedImage(path string) (*image.Paletted, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open file at '%s': %w", path, err)
	}
	defer f.Close()

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

// Returns the final palette, the last index in the input palette, and true if
// this index must be remapped to 0xFF.
func imagePalette(path string) (sprite.Palette, uint8, bool, error) {
	var palette sprite.Palette

	img, err := openPalettedImage(path)
	if err != nil {
		return palette, 0, false, fmt.Errorf("unable to open image: %w", err)
	}

	if len(img.Palette) > 256 {
		return palette, 0, false, fmt.Errorf("expected at most 256 colors palette, got %d", len(palette))
	}

	for i, v := range img.Palette {
		r, g, b, _ := v.RGBA()
		palette[i] = sprite.RGB{
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
