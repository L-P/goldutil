package main

import (
	"errors"
	"fmt"
	"goldutil/goldsrc"
	"goldutil/qmap"
	"goldutil/set"
	"goldutil/sprite"
	"goldutil/wad"
	"os"
	"path/filepath"
	"sort"

	"github.com/urfave/cli/v2"
)

var Version = "unknown version"

func main() {
	var app = newApp()
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

//nolint:lll,funlen // descriptions
func newApp() *cli.App {
	return &cli.App{
		Version: Version,
		Usage:   "GoldSrc modding utility.",
		Commands: []*cli.Command{
			{
				Name:  "map",
				Usage: "Read and write MAP files.",
				Subcommands: []*cli.Command{
					{
						Name:      "export",
						Usage:     "Exports a .map file the way TrenchBroom does, removing all layers marked as not exported. Output is written to stdout.",
						ArgsUsage: " MAP",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "cleanup-tb",
								Value: false,
								Usage: "Removes properties added by TrenchBroom that are not understood by the engine and spam the console with errors.",
							},
						},
						Action: doMapExport,
					},

					{
						Name:      "graph",
						Usage:     "Creates a graphviz digraph of entity caller/callee relationships from a .map file. ripent exports use the same format and can be read too. Output is written to stdout.",
						ArgsUsage: " MAP",
						Action:    doMapGraph,
					},
				},
			},

			{
				Name:  "spr",
				Usage: "Read and write SPR files (sprites).",
				Subcommands: []*cli.Command{
					{
						Name:      "info",
						Usage:     "Prints parsed frame data from a sprite.",
						ArgsUsage: " SPR",
						Action:    doSpriteInfo,
					},

					{
						Name:      "extract",
						Usage:     "Outputs all frames of a sprite to the current directory. The output files will be named after the original sprite file name plus a frame number suffix and an extension.",
						ArgsUsage: " SPR",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "dir",
								Usage: "Outputs frames to the specified directory instead of the current one.",
							},
						},
						Action: doSpriteExtract,
					},

					{
						Name:      "create",
						Usage:     "Creates a sprite from the given ordered list of PNG frames and writes it to the given SPR path.\nInput images must be 256 colors paletted PNGs. The palette of the first frame will be used, the other palettes are discarded and all frames will be interpreted using the first frame's palette.  If the palette has under 256 colors it will be extended to 256, putting the last color of the palette in the 256th spot and remapping the image to match this updated palette. This matters for some texture formats.",
						ArgsUsage: " FRAME0 [FRAMEX…]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "out",
								Usage:    "Path to the output .spr file.",
								Required: true,
							},
							&cli.StringFlag{
								Name:  "type",
								Value: "parallel",
								Usage: `Sprite type, TYPE can be any one of:
parallel           Always face camera. (Default)
parallel-upright   Always face camera except for the locked Z axis.
oriented           Orientation set by the level.
parallel-oriented  Faces camera but can be rotated by the level.
facing-upright     Like parallel upright but faces the player origin instead of the camera.`,
							},
							&cli.StringFlag{
								Name:  "format",
								Value: "normal",
								Usage: `Texture format, determines how the palette is interpreted and the texture is rendered by the engine. FORMAT can be any one of:
normal      256 colors sprite. (Default)
additive    Additive 256 colors sprite.
index-alpha Monochromatic sprite with 256 alpha levels, the base color is determined by the last color on the palette.
alpha-test  Transparent 255 colors sprite. The 256th color on the palette will be rendered as fully transparent.`,
							},
						},
						Action: doSpriteCreate,
					},
				},
			},

			{
				Name:  "wad",
				Usage: "Read and write WAD files.",
				Subcommands: []*cli.Command{
					{
						Name:      "create",
						Usage:     "Creates a WAD file from a list of PNG files and directories. Directories are not scanned recursively and only PNG files are used.\nFile base names (without extensions) are uppercased and used as texture names. This means that names exceeding 15 chars will trigger an error.",
						ArgsUsage: " PATH [PATH…]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "out",
								Usage:    "Path to the output .wad file.",
								Required: true,
							},
						},
						Action: doWADCreate,
					},

					{
						Name:      "extract",
						Usage:     "Extracts a WAD file in the given directory as a bunch of PNG files.",
						ArgsUsage: " WAD",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "dir",
								Usage:    "Path to the output directory.",
								Required: true,
							},
						},
						Action: doWADExtract,
					},

					{
						Name:      "info",
						Usage:     "Prints parsed data from a WAD file.",
						ArgsUsage: " WAD",
						Action:    doWADInfo,
					},
				},
			},

			{
				Name:  "bsp",
				Usage: "Read and write BSP files.",
				Subcommands: []*cli.Command{
					{
						Name:      "remap-materials",
						Usage:     "On a BSP with embedded textures, change their names so they can match what's in materials.txt.",
						ArgsUsage: " BSP",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "original-materials",
								Value:    "valve/sound/materials.txt",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "replacement-materials",
								Value:    "valve_addon/sound/materials.txt",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "out",
								Required: true,
								Usage:    "Where to write the remapped BSP.",
							},
						},
						Action: doBSPRemapMaterials,
					},
				},
			},
		},
	}
}

func doSpriteExtract(cCtx *cli.Context) error {
	path := cCtx.Args().Get(0)
	if path == "" {
		return errors.New("expected one argument: the .spr to parse and extract")
	}

	spr, err := sprite.NewFromFile(path)
	if err != nil {
		return fmt.Errorf("unable to open sprite: %w", err)
	}

	return extractSprite(spr, cCtx.String("dir"), filepath.Base(path))
}

func doSpriteCreate(cCtx *cli.Context) error {
	typ, ok := map[string]sprite.Type{
		"parallel-upright":  sprite.ParallelUpright,
		"facing-upright":    sprite.FacingUpright,
		"parallel":          sprite.Parallel,
		"oriented":          sprite.Oriented,
		"parallel-oriented": sprite.ParallelOriented,
	}[cCtx.String("type")]
	if !ok {
		return errors.New("unrecognize sprite type")
	}

	format, ok := map[string]sprite.TextureFormat{
		"normal":      sprite.Normal,
		"additive":    sprite.Additive,
		"index-alpha": sprite.IndexAlpha,
		"alpha-test":  sprite.AlphaTest,
	}[cCtx.String("format")]
	if !ok {
		return errors.New("unrecognize texture format")
	}

	spr, err := createSprite(typ, format, cCtx.Args().Slice())
	if err != nil {
		return fmt.Errorf("unable to create sprite: %w", err)
	}

	dest, err := os.OpenFile(cCtx.String("out"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
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

func doSpriteInfo(cCtx *cli.Context) error {
	path := cCtx.Args().Get(0)
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

func doMapGraph(cCtx *cli.Context) error {
	path := cCtx.Args().Get(0)
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

func doMapExport(cCtx *cli.Context) error {
	path := cCtx.Args().Get(0)
	if path == "" {
		return errors.New("expected one argument: the .map to export")
	}

	qm, err := qmap.LoadFromFile(path)
	if err != nil {
		return fmt.Errorf("unable to read from map: %w", err)
	}

	clean, err := exportQMap(qm, cCtx.Bool("cleanup-tb"))
	if err != nil {
		return fmt.Errorf("unable to export map: %w", err)
	}

	fmt.Print(clean.String())

	return nil
}

func doWADExtract(cCtx *cli.Context) error {
	var dir = cCtx.String("dir")
	stat, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("unable to use destination directory: %w", err)
	}
	if err == nil && !stat.IsDir() {
		return errors.New("output directory paths exists but is not a directory")
	}

	wad3, err := wad.NewFromFile(cCtx.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to open and parse WAD file: %w", err)
	}

	return extractWAD(wad3, dir)
}

func doWADInfo(cCtx *cli.Context) error {
	wad3, err := wad.NewFromFile(cCtx.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to open and parse WAD file: %w", err)
	}

	fmt.Println(wad3.String())

	return nil
}

func doWADCreate(cCtx *cli.Context) error {
	input, err := collectPaths(cCtx.Args().Slice(), "*.png")
	if err != nil {
		return fmt.Errorf("unable to collect paths: %w", err)
	}

	return createWAD(cCtx.String("out"), input)
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

func doBSPRemapMaterials(cCtx *cli.Context) error {
	source, err := goldsrc.LoadMaterialsFromFile(cCtx.String("original-materials"))
	if err != nil {
		return fmt.Errorf("unable to load original-materials: %w", err)
	}

	replacement, err := goldsrc.LoadMaterialsFromFile(cCtx.String("replacement-materials"))
	if err != nil {
		return fmt.Errorf("unable to load replacement-materials: %w", err)
	}

	bsp, err := goldsrc.LoadBSPFromFile(cCtx.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to load BSP: %w", err)
	}

	_, _, _ = source, replacement, bsp

	return nil
}
