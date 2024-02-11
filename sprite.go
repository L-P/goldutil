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

func extractSprite(spr sprite.Sprite, destDir, originalBaseName string) error {
	for i := range spr.Frames {
		var (
			destPath = filepath.Join(destDir, fmt.Sprintf(
				"%s.frame%03d.png",
				originalBaseName, i,
			))
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
	img, err := openPalettedImage(path)
	if err != nil {
		return fmt.Errorf("unable to open image: %w", err)
	}

	rect := img.Bounds()
	if rect.Dx() != int(spr.Width) || rect.Dy() != int(spr.Height) {
		return fmt.Errorf("image size does not match sprite size")
	}

	if shouldRemap {
		remapLastColor(img, remapIndex)
	}

	spr.AddFrame(sprite.NewFrame(
		int32(rect.Dx()), int32(rect.Dy()),
		int32(rect.Dx()/2), int32(rect.Dy()/2),
		img.Pix,
	))

	return nil
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
