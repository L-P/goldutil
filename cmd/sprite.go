package main

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/urfave/cli/v3"

	"github.com/L-P/goldutil/goldsrc/sprite"
	"github.com/L-P/goldutil/palette"
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

	fmt.Fprintf(cmd.Writer, "Palette: \n  ")

	for i, rgb := range spr.Palette {
		luma := 0.299*float32(rgb.R) + 0.587*float32(rgb.G) + 0.114*float32(rgb.B)
		c := color.New(color.FgBlack)
		if luma <= 128 {
			c = color.New(color.FgWhite)
		}
		c.AddBgRGB(int(rgb.R), int(rgb.G), int(rgb.B))
		var pad rune
		if i%16 != 15 {
			pad = ' '
		}
		c.Fprintf(cmd.Writer, "%02X%02X%02X%c", rgb.R, rgb.G, rgb.B, pad) //nolint:errcheck // don't care.

		if (i+1)%16 == 0 && i != 255 {
			fmt.Fprint(cmd.Writer, "\n  ")
		}
	}

	fmt.Fprintln(cmd.Writer)

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
