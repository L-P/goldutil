package main

import (
	"fmt"
	"goldutil/sprite"
	"image/png"
	"os"
	"path/filepath"
)

func extractSprite(spr sprite.Sprite, destDir string) error {
	fmt.Println(spr.String())

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
