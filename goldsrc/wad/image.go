package wad

import (
	"fmt"
	"github.com/L-P/goldutil/palette"
	"image"
	"path/filepath"
	"strings"
)

// Because a MIPTexture cannot exist without a name we need to wrap the input
// type and stick a string to it.
type NamedImage struct {
	Image *image.Paletted
	Name  string
}

func OpenNamedImage(path string) (NamedImage, error) {
	img, err := palette.OpenPalettedImage(path)
	if err != nil {
		return NamedImage{}, fmt.Errorf("unable to open image: %w", err)
	}

	return NamedImage{
		Image: img,
		Name:  strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
	}, nil
}

func NewFromImages(images ...NamedImage) (WAD, error) {
	wad := New()

	for _, img := range images {
		tex, err := NewMIPTextureFromImage(img)
		if err != nil {
			return WAD{}, fmt.Errorf("unable to create texture: %w", err)
		}

		if err := wad.AddTexture(tex); err != nil {
			return WAD{}, fmt.Errorf("unable to add texture: %w", err)
		}
	}

	return wad, nil
}

func NewMIPTextureFromImage(input NamedImage) (MIPTexture, error) {
	var zero MIPTexture
	width := input.Image.Rect.Max.X
	height := input.Image.Rect.Max.Y

	ret, err := NewMIPTexture(input.Name, width, height)
	if err != nil {
		return zero, fmt.Errorf("unable to create MIPTexture: %w", err)
	}

	pal, remapIndex, shouldRemap, err := palette.FromImage(input.Image)
	if err != nil {
		return zero, fmt.Errorf("unable to process image palette: %w", err)
	}
	if shouldRemap {
		palette.RemapLastColor(input.Image, remapIndex)
	}

	copy(ret.Palette[:], pal[:])

	if err := ret.SetData(input.Image.Pix); err != nil {
		return zero, fmt.Errorf("unable to write pix data to texture: %w", err)
	}

	return ret, nil
}
