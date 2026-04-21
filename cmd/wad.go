package main

import (
	"context"
	"errors"
	"fmt"
	"goldutil/goldsrc/wad"
	"goldutil/set"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/urfave/cli/v3"
)

func doWADInfo(ctx context.Context, cmd *cli.Command) error {
	wad3, err := wad.NewFromFile(cmd.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to open and parse WAD file: %w", err)
	}

	fmt.Fprintln(cmd.Writer, wad3.String())

	return nil
}

func doWADCreate(ctx context.Context, cmd *cli.Command) error {
	input, err := collectPaths(cmd.Args().Slice(), "*.png")
	if err != nil {
		return fmt.Errorf("unable to collect paths: %w", err)
	}

	return createWAD(cmd.String("out"), input)
}

func doWADExtract(ctx context.Context, cmd *cli.Command) error {
	var dir = cmd.String("dir")
	stat, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("unable to use destination directory: %w", err)
	}
	if err == nil && !stat.IsDir() {
		return errors.New("output directory paths exists but is not a directory")
	}

	wad3, err := wad.NewFromFile(cmd.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to open and parse WAD file: %w", err)
	}

	return extractWAD(wad3, dir)
}

func extractWAD(wad wad.WAD, dir string) error {
	for _, name := range wad.Names() {
		tex, ok := wad.GetTexture(name)
		if !ok {
			panic("has name but no texture, programming error")
		}

		if strings.ContainsRune(name, os.PathSeparator) {
			return fmt.Errorf("texture name contains a separator: %s", name)
		}
		destPath := filepath.Join(dir, name+".png")

		if err := writeTexture(tex, destPath); err != nil {
			return fmt.Errorf("unable to write texture: %w", err)
		}
	}

	return nil
}

func writeTexture(tex wad.MIPTexture, destPath string) error {
	img, err := tex.Render()
	if err != nil {
		return fmt.Errorf("unable to render texture: %w", err)
	}

	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("unable to open '%s' for writing: %w", destPath, err)
	}

	if err := png.Encode(dest, img); err != nil {
		dest.Close() //nolint:errcheck // in another error path already
		return fmt.Errorf("unable to encode png: %w", err)
	}

	if err := dest.Close(); err != nil {
		return fmt.Errorf("unable to finalize writing to '%s': %w", destPath, err)
	}

	return nil
}

func createWAD(destPath string, inputFiles []string) error {
	wad := wad.New()

	for _, path := range inputFiles {
		tex, err := createTexture(path)
		if err != nil {
			return fmt.Errorf("unable to create texture: %w", err)
		}

		if err := wad.AddTexture(tex); err != nil {
			return fmt.Errorf("unable to add texture: %w", err)
		}
	}

	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("unable to open '%s' for writing: %w", destPath, err)
	}

	if err := wad.Write(dest); err != nil {
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

// Returns the paths when they're files, and the pattern-matching files inside
// them if they're directories.
func collectPaths(input []string, pattern string) ([]string, error) {
	ret := make([]string, 0, len(input))

	for _, path := range input {
		stat, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("could not stat '%s': %w", path, err)
		}

		if !stat.IsDir() {
			ret = append(ret, path)
			continue
		}

		matches, err := filepath.Glob(filepath.Join(path, pattern))
		if err != nil {
			return nil, fmt.Errorf("unable to glob dir '%s': %w", path, err)
		}

		ret = append(ret, matches...)
	}

	ret = dedupeStrs(ret)
	sort.Strings(ret)

	return ret, nil
}

func dedupeStrs(in []string) []string {
	var (
		ret  = make([]string, 0, len(in))
		seen = set.NewPresenceSet[string](len(in))
	)

	for _, v := range in {
		if seen.Has(v) {
			continue
		}

		ret = append(ret, v)
		seen.Set(v)
	}

	return ret
}
