package main

import (
	"errors"
	"fmt"
	"goldutil/sprite"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

func extractSprite(spr sprite.Sprite, destDir string) error {
	for i := range spr.Frames {
		var (
			destPath  = filepath.Join(destDir, fmt.Sprintf("frame%03d.png", i))
			dest, err = os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		)
		if err != nil {
			return fmt.Errorf("unable to open '%s' for writing: %w", destPath, err)
		}

		img, err := spr.RenderFrame(i)
		if err != nil {
			dest.Close()
			return fmt.Errorf("unable to encode frame %d: %w", i, err)
		}

		if err := png.Encode(dest, img); err != nil {
			dest.Close()
			return fmt.Errorf("unable to encode png: %w", err)
		}

		if err := dest.Close(); err != nil {
			return fmt.Errorf("unable to finalize writing to '%s': %w", destPath, err)
		}
	}

	return nil
}

func createSprite(typ sprite.Type, format sprite.TextureFormat, framePaths []string) (sprite.Sprite, error) {
	if len(framePaths) < 1 {
		return sprite.Sprite{}, errors.New("at least one frame is required")
	}
	width, height, err := imageSize(framePaths[0])
	if err != nil {
		return sprite.Sprite{}, fmt.Errorf("unable to read first frame dimensions: %w", err)
	}
	if (width%16 != 0) || (height%16 != 0) {
		return sprite.Sprite{}, fmt.Errorf("dimensions not divisible by 16: %w", err)
	}

	palette, remapIndex, shouldRemap, err := imagePalette(framePaths[0])
	if err != nil {
		return sprite.Sprite{}, fmt.Errorf("unable to process first frame palette: %w", err)
	}

	spr, err := sprite.New(width, height, typ, format, palette)
	if err != nil {
		return sprite.Sprite{}, fmt.Errorf("unable to create empty sprite: %w", err)
	}

	for i, inPath := range framePaths {
		if err := addFrameToSprite(&spr, inPath, remapIndex, shouldRemap); err != nil {
			return spr, fmt.Errorf("unable to add frame #%d: %w", i, err)
		}
	}

	return spr, nil
}

func addFrameToSprite(spr *sprite.Sprite, path string, remapIndex uint8, shouldRemap bool) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("unable to open file at '%s': %w", path, err)
	}
	defer f.Close()

	mysteryImg, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("unable to decode image: %w", err)
	}

	img, ok := mysteryImg.(*image.Paletted)
	if !ok {
		return errors.New("image is not paletted")
	}

	rect := img.Bounds()
	if rect.Dx() != int(spr.Width) || rect.Dy() != int(spr.Height) {
		return fmt.Errorf("image size does not match sprite size")
	}

	if shouldRemap {
		for i, v := range img.Pix {
			if v == remapIndex {
				img.Pix[i] = 0xFF
			}
		}
	}

	spr.AddFrame(sprite.NewFrame(
		int32(rect.Dx()), int32(rect.Dy()),
		// TODO understand why I'm seeing so much negative origins in valve sprites.
		-int32(rect.Dx()/2), -int32(rect.Dy()/2),
		img.Pix,
	))

	return nil
}

// Returns the final palette, the last index in the input palette, and true if
// this index must be remapped to 0xFF.
func imagePalette(path string) (sprite.Palette, uint8, bool, error) {
	var palette sprite.Palette

	f, err := os.Open(path)
	if err != nil {
		return palette, 0, false, fmt.Errorf("unable to open file at '%s': %w", path, err)
	}
	defer f.Close()

	mysteryImg, _, err := image.Decode(f)
	if err != nil {
		return palette, 0, false, fmt.Errorf("could not decode image: %w", err)
	}

	img, ok := mysteryImg.(*image.Paletted)
	if !ok {
		return palette, 0, false, errors.New("image is not paletted")
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

func imageSize(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to open image for reading: %w", err)
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to decode image config: %w", err)
	}

	return cfg.Width, cfg.Height, nil
}
