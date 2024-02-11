package main

import (
	"fmt"
	"goldutil/wad"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

func extractWAD(wad3 wad.WAD, dir string) error {
	for _, name := range wad3.Names() {
		tex, ok := wad3.GetTexture(name)
		if !ok {
			panic("has name but no texture, programming error")
		}

		if strings.ContainsRune(name, os.PathSeparator) {
			return fmt.Errorf("texture name contains a separator: %s", name)
		}
		destPath := filepath.Join(dir, fmt.Sprintf("%s.png", name))

		if err := writeTexture(tex, destPath); err != nil {
			return fmt.Errorf("unable to write texture: %w", err)
		}
	}

	return nil
}

func writeTexture(tex wad.MIPTexture, destPath string) error {
	img, err := tex.Render(0)
	if err != nil {
		return fmt.Errorf("unable to render texture: %w", err)
	}

	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open '%s' for writing: %w", destPath, err)
	}

	if err := png.Encode(dest, img); err != nil {
		dest.Close()
		return fmt.Errorf("unable to encode png: %w", err)
	}

	if err := dest.Close(); err != nil {
		return fmt.Errorf("unable to finalize writing to '%s': %w", destPath, err)
	}

	return nil
}

func createWAD(destPath string, inputFiles []string) error {
	wad3 := wad.New()

	for _, path := range inputFiles {
		tex, err := createTexture(path)
		if err != nil {
			return fmt.Errorf("unable to create texture: %w", err)
		}

		if err := wad3.AddTexture(tex); err != nil {
			return fmt.Errorf("unable to add texture: %w", err)
		}
	}

	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open '%s' for writing: %w", destPath, err)
	}

	if err := wad3.Write(dest); err != nil {
		return fmt.Errorf("unable to write to WAD file: %w", err)
	}

	if err := dest.Close(); err != nil {
		return fmt.Errorf("unable to finalize writing to '%s': %w", destPath, err)
	}

	return nil
}

func createTexture(path string) (wad.MIPTexture, error) {
	var (
		empty              = wad.MIPTexture{}
		name               = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		width, height, err = imageSize(path)
	)
	if err != nil {
		return empty, fmt.Errorf("unable to read image dimensions: %w", err)
	}

	ret, err := wad.NewMIPTexture(name, width, height)
	if err != nil {
		return empty, fmt.Errorf("unable to create MIPTexture: %w", err)
	}

	palette, remapIndex, shouldRemap, err := imagePalette(path)
	if err != nil {
		return empty, fmt.Errorf("unable to process image palette: %w", err)
	}

	img, err := openPalettedImage(path)
	if err != nil {
		return empty, fmt.Errorf("unable to open image: %w", err)
	}
	if shouldRemap {
		remapLastColor(img, remapIndex)
	}

	copy(ret.Palette[:], palette[:])

	if err := ret.SetData(img.Pix); err != nil {
		return empty, fmt.Errorf("unable to write pix data to texture: %w", err)
	}

	return ret, nil
}
