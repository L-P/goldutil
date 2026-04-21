package main

import (
	"context"
	"errors"
	"fmt"
	"goldutil/goldsrc/sprite"
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

	spr, err := createSprite(typ, format, cmd.Args().Slice())
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

func createSprite(
	typ sprite.Type,
	format sprite.TextureFormat,
	framePaths []string,
) (sprite.Sprite, error) {
	var zero sprite.Sprite

	if len(framePaths) < 1 {
		return zero, errors.New("at least one frame is required")
	}
	width, height, err := imageSize(framePaths[0])
	if err != nil {
		return zero, fmt.Errorf("unable to read first frame dimensions: %w", err)
	}
	if (width%4 != 0) || (height%4 != 0) {
		return zero, errors.New("dimensions not divisible by 4")
	}

	palette, remapIndex, shouldRemap, err := imagePalette(framePaths[0])
	if err != nil {
		return zero, fmt.Errorf("unable to process first frame palette: %w", err)
	}

	spr, err := sprite.New(width, height, typ, format, palette)
	if err != nil {
		return zero, fmt.Errorf("unable to create empty sprite: %w", err)
	}

	for i, inPath := range framePaths {
		if err := addFrameToSprite(&spr, inPath, remapIndex, shouldRemap); err != nil {
			return zero, fmt.Errorf("unable to add frame #%d: %w", i, err)
		}
	}

	return spr, nil
}

func addFrameToSprite(spr *sprite.Sprite, path string, remapIndex uint8, shouldRemap bool) error {
	img, err := openPalettedImage(path)
	if err != nil {
		return fmt.Errorf("unable to open image: %w", err)
	}

	if shouldRemap {
		remapLastColor(img, remapIndex)
	}

	rect := img.Bounds()
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
	defer f.Close() //nolint:errcheck // readonly

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to decode image config: %w", err)
	}

	return cfg.Width, cfg.Height, nil
}
