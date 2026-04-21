package sprite

import (
	"errors"
	"fmt"
	"goldutil/palette"
	"image"
)

func NewFromImage(typ Type, format TextureFormat, frames ...*image.Paletted) (Sprite, error) {
	var zero Sprite

	if len(frames) < 1 {
		return zero, errors.New("at least one frame is required")
	}

	width, height := frames[0].Bounds().Max.X, frames[0].Bounds().Max.Y
	if (width%4 != 0) || (height%4 != 0) {
		return zero, errors.New("first frame dimensions not divisible by 4")
	}

	pal, remapIndex, shouldRemap, err := palette.FromImage(frames[0])
	if err != nil {
		return zero, fmt.Errorf("unable to process first frame palette: %w", err)
	}

	spr, err := New(width, height, typ, format, pal)
	if err != nil {
		return zero, fmt.Errorf("unable to create empty sprite: %w", err)
	}

	for i, frame := range frames {
		if err := spr.addFrameFromImage(frame, remapIndex, shouldRemap); err != nil {
			return zero, fmt.Errorf("unable to add frame #%d: %w", i, err)
		}
	}

	return spr, nil
}

func (spr *Sprite) addFrameFromImage(img *image.Paletted, remapIndex uint8, shouldRemap bool) error {
	if shouldRemap {
		palette.RemapLastColor(img, remapIndex)
	}

	rect := img.Bounds()
	if (rect.Max.X%4 != 0) || (rect.Max.Y%4 != 0) {
		return errors.New("frame dimensions not divisible by 4")
	}

	spr.AddFrame(NewFrame(
		int32(rect.Dx()), int32(rect.Dy()),
		int32(rect.Dx()/2), int32(rect.Dy()/2),
		img.Pix,
	))

	return nil
}
