package main

import (
	"fmt"
	"goldutil/wad"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func extractWAD(wad3 wad.WAD, dir string) error {
	for _, name := range wad3.Names() {
		if strings.ContainsRune(name, os.PathSeparator) {
			return fmt.Errorf("texture name contains a separator: %s", name)
		}

		var (
			destPath  = filepath.Join(dir, fmt.Sprintf("%s.png", name))
			dest, err = os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		)
		if err != nil {
			return fmt.Errorf("unable to open '%s' for writing: %w", destPath, err)
		}

		tex, ok := wad3.GetTexture(name)
		if !ok {
			panic("has name but no texture, programming error")
		}

		if err := writeTexture(tex, dest); err != nil {
			dest.Close()
			return fmt.Errorf(": %w", err)
		}

		if err := dest.Close(); err != nil {
			return fmt.Errorf("unable to finalize writing to '%s': %w", destPath, err)
		}
	}

	return nil
}

func writeTexture(tex wad.MIPTexture, w io.Writer) error {
	img, err := tex.Render(0)
	if err != nil {
		return fmt.Errorf("unable to render texture: %w", err)
	}

	if err := png.Encode(w, img); err != nil {
		return fmt.Errorf("unable to encode png: %w", err)
	}

	return nil
}
