package main

import (
	"context"
	"errors"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/L-P/goldutil/goldsrc/wad"
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

	return extractWAD(wad3, dir, !cmd.Bool("no-alpha"))
}

func extractWAD(wad wad.WAD, dir string, withAlpha bool) error {
	for _, name := range wad.Names() {
		tex, ok := wad.GetTexture(name)
		if !ok {
			panic("has name but no texture, programming error")
		}

		if strings.ContainsRune(name, os.PathSeparator) {
			return fmt.Errorf("texture name contains a separator: %s", name)
		}
		destPath := filepath.Join(dir, name+".png")

		if err := writeTexture(tex, destPath, withAlpha); err != nil {
			return fmt.Errorf("unable to write texture: %w", err)
		}
	}

	return nil
}

func writeTexture(tex wad.MIPTexture, destPath string, withAlpha bool) error {
	img, err := tex.Render(withAlpha)
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
	var err error
	images := make([]wad.NamedImage, len(inputFiles))
	for i, path := range inputFiles {
		images[i], err = wad.OpenNamedImage(path)
		if err != nil {
			return fmt.Errorf("unable to open image at path %s: %w", path, err)
		}
	}

	wad, err := wad.NewFromImages(images...)
	if err != nil {
		return fmt.Errorf("unable to create wad from images: %w", err)
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

	sort.Strings(ret)
	ret = slices.Compact(ret)

	return ret, nil
}
