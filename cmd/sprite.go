package main

import (
	"context"
	"errors"
	"fmt"
	"goldutil/goldsrc/sprite"
	"goldutil/palette"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

func doSpriteExtract(ctx context.Context, cmd *cli.Command) error {
	path := cmd.Args().Get(0)
	if path == "" {
		return errors.New("expected one argument: the .spr to parse and extract")
	}

	spr, err := sprite.NewFromFile(path)
	if err != nil {
		return fmt.Errorf("unable to open sprite: %w", err)
	}

	return extractSprite(
		spr,
		cmd.String("dir"),
		filepath.Base(path),
		!cmd.Bool("no-alpha"),
	)
}

func doSpriteCreate(ctx context.Context, cmd *cli.Command) error {
	typ, err := sprite.ParseType(cmd.String("type"))
	if err != nil {
		return fmt.Errorf("unable to parse sprite type: %w", err)
	}

	format, err := sprite.ParseTextureFormat(cmd.String("format"))
	if err != nil {
		return fmt.Errorf("unable to parse texture format: %w", err)
	}

	images := make([]*image.Paletted, cmd.Args().Len())
	for i, path := range cmd.Args().Slice() {
		img, err := palette.OpenPalettedImage(path)
		if err != nil {
			return fmt.Errorf("unable to open image at '%s': %w", path, err)
		}
		images[i] = img
	}

	spr, err := sprite.NewFromImage(typ, format, images...)
	if err != nil {
		return fmt.Errorf("unable to create sprite: %w", err)
	}

	dest, err := os.OpenFile(cmd.String("out"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("unable to open dest SPR for writing: %w", err)
	}

	if err := spr.Write(dest); err != nil {
		return fmt.Errorf("unable to write to destination SPR: %w", err)
	}

	if err := dest.Close(); err != nil {
		return fmt.Errorf("unable to finalize writing to destination SPR: %w", err)
	}

	return nil
}

func doSpriteInfo(ctx context.Context, cmd *cli.Command) error {
	path := cmd.Args().Get(0)
	if path == "" {
		return errors.New("expected one argument: the .spr to parse and display")
	}

	spr, err := sprite.NewFromFile(path)
	if err != nil {
		return fmt.Errorf("unable to open sprite: %w", err)
	}

	fmt.Fprintln(cmd.Writer, spr.String())

	return nil
}

func extractSprite(spr sprite.Sprite, destDir, originalBaseName string, withAlpha bool) error {
	for i := range spr.Frames {
		var (
			destPath = filepath.Join(destDir, fmt.Sprintf(
				"%s.frame%03d.png",
				originalBaseName, i,
			))
			dest, err = os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		)
		if err != nil {
			return fmt.Errorf("unable to open '%s' for writing: %w", destPath, err)
		}

		img, err := spr.RenderFrame(i, withAlpha)
		if err != nil {
			dest.Close() //nolint:errcheck // in another error path already
			return fmt.Errorf("unable to encode frame %d: %w", i, err)
		}

		if err := png.Encode(dest, img); err != nil {
			dest.Close() //nolint:errcheck // in another error path already
			return fmt.Errorf("unable to encode png: %w", err)
		}

		if err := dest.Close(); err != nil {
			return fmt.Errorf("unable to finalize writing to '%s': %w", destPath, err)
		}
	}

	return nil
}
