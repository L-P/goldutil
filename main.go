package main

import (
	"errors"
	"flag"
	"fmt"
	"goldutil/qmap"
	"goldutil/sprite"
	"os"
)

var help = `Usage: %s COMMAND [ARGS…]

Commands:
    entity-graph MAP
        Outputs a graphviz digraph of entity caller/callee relationships from a
        .map file. ripent exports use the same format and can be read too.

    sprite-info SPR
        Prints parsed frame data from a sprite.

    sprite-extract [-dir DIR] SPR
        Outputs all frames of a sprite to the current directory.

        Options:
            -dir DIR    Outputs frames to the specified directory instead of
                        the current one.

    sprite-create [-type TYPE] [-format FORMAT] FRAME0 [FRAMEX…]
        Creates a sprite from the given ordered list of PNG frames and writes
        it to the given SPR path.
        Input images must be 256 colors paletted PNGs. The palette of
        the first frame will be used, the other palettes are discarded and all
        frames will be interpreted using the first frame's palette.
        If the palette has under 256 colors it will be extended to 256,
        putting the last color of the palette in the 256th spot and remapping
        the image to match this updated palette. This matters for some texture
        formats.

        Options:
            -out SPR
                Path to the output .spr file.

            -type TYPE
                Sprite type, TYPE can be any one of:

                parallel           Always face camera. (Default)
                parallel-upright   Always face camera except for the locked Z axis.
                oriented           Orientation set by the level.
                parallel-oriented  Faces camera but can be rotated by the level.
                facing-upright     Like parallel upright but faces the player
                                   origin instead of the camera.

            -format FORMAT
                Texture format, determines how the palette is interpreted and the
                texture is rendered by the engine. FORMAT can be any one of:

                normal      256 colors sprite. (Default)
                additive    Additive 256 colors sprite.
                index-alpha Monochromatic sprite with 256 alpha levels, the base
                            color is determined by the last color on the palette.
                alpha-test  Transparent 255 colors sprite. The 256th color on the
                            palette will be rendered as fully transparent.
`

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), help, os.Args[0])
}

func main() {
	flag.Usage = usage

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, help, os.Args[0])
		os.Exit(1)
	}

	if err := dispatch(os.Args[1], os.Args[2:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			flag.Usage()
			return
		}

		fmt.Fprintln(os.Stderr, err.Error())
	}
}

func dispatch(command string, args []string) error {
	switch command {
	case "entity-graph":
		return doEntGraph(args)
	case "sprite-info":
		return doSpriteInfo(args)
	case "sprite-extract":
		return doSpriteExtract(args)
	case "sprite-create":
		return doSpriteCreate(args)
	default:
		return fmt.Errorf("unrecognized command: %s", command)
	}
}

func doSpriteExtract(args []string) error {
	fset := flag.NewFlagSet("sprite-extract", flag.ExitOnError)
	fset.Usage = usage
	dir := fset.String("dir", "", "destination directory")

	if err := fset.Parse(args); err != nil {
		return err
	}

	path := fset.Arg(0)
	if path == "" {
		return errors.New("expected one argument: the .spr to parse and extract")
	}

	spr, err := sprite.NewFromFile(path)
	if err != nil {
		return fmt.Errorf("unable to open sprite: %w", err)
	}

	return extractSprite(spr, *dir)
}

func doSpriteCreate(args []string) error {
	fset := flag.NewFlagSet("sprite-extract", flag.ExitOnError)
	fset.Usage = usage
	formatStr := fset.String("format", "", "texture format (normal, additive, index-alpha, alpha-test)")
	typeStr := fset.String("type", "", "sprite type (parallel-upright, facing-upright, parallel, oriented, parallel-oriented)") //nolint
	out := fset.String("out", "", "destination .spr file")                                                                      //nolint
	if err := fset.Parse(args); err != nil {
		return err
	}

	typ, ok := map[string]sprite.Type{
		"parallel-upright":  sprite.ParallelUpright,
		"facing-upright":    sprite.FacingUpright,
		"parallel":          sprite.Parallel,
		"oriented":          sprite.Oriented,
		"parallel-oriented": sprite.ParallelOriented,
	}[*typeStr]
	if !ok {
		return errors.New("unrecognize sprite type")
	}

	format, ok := map[string]sprite.TextureFormat{
		"normal":      sprite.Normal,
		"additive":    sprite.Additive,
		"index-alpha": sprite.IndexAlpha,
		"alpha-test":  sprite.AlphaTest,
	}[*formatStr]
	if !ok {
		return errors.New("unrecognize texture format")
	}

	spr, err := createSprite(typ, format, fset.Args())
	if err != nil {
		return fmt.Errorf("unable to create sprite: %w", err)
	}

	dest, err := os.OpenFile(*out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
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

func doSpriteInfo(args []string) error {
	fset := flag.NewFlagSet("sprite-info", flag.ExitOnError)
	fset.Usage = usage
	if err := fset.Parse(args); err != nil {
		return err
	}

	path := fset.Arg(0)
	if path == "" {
		return errors.New("expected one argument: the .spr to parse and display")
	}

	spr, err := sprite.NewFromFile(path)
	if err != nil {
		return fmt.Errorf("unable to open sprite: %w", err)
	}

	fmt.Println(spr.String())

	return nil
}

func doEntGraph(args []string) error {
	fset := flag.NewFlagSet("entity-graph", flag.ExitOnError)
	fset.Usage = usage
	if err := fset.Parse(args); err != nil {
		return err
	}

	path := fset.Arg(0)
	if path == "" {
		return errors.New("expected one argument: the .map to parse and graph")
	}

	qm, err := qmap.LoadFromFile(path)
	if err != nil {
		return fmt.Errorf("unable to read from map: %w", err)
	}

	GraphQMap(qm, os.Stdout)

	return nil
}
