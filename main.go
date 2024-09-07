package main

import (
	_ "embed"
	"errors"
	"fmt"
	"goldutil/goldsrc"
	"goldutil/qmap"
	"goldutil/set"
	"goldutil/sprite"
	"goldutil/wad"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
)

var Version = "unknown version"

func main() {
	var app = newApp()
	if err := app.Run(os.Args); err != nil {
		// HACK: -h will panic.
		if err.Error() != "flag: help requested" {
			panic(err)
		}
	}
}

//nolint:funlen // descriptions
func newApp() *cli.App {
	cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		_ = doHelp(nil)
	}

	return &cli.App{
		Version: Version,
		Commands: []*cli.Command{
			{
				Name:   "help",
				Action: doHelp,
			},
			{
				Name: "map",
				Subcommands: []*cli.Command{
					{
						Name: "export",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "cleanup-tb",
								Value: false,
							},
						},
						Action: doMapExport,
					},

					{
						Name:   "graph",
						Action: doMapGraph,
					},
				},
			},

			{
				Name: "spr",
				Subcommands: []*cli.Command{
					{
						Name:   "info",
						Action: doSpriteInfo,
					},

					{
						Name: "extract",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "dir"},
						},
						Action: doSpriteExtract,
					},

					{
						Name: "create",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "out",
								Required: true,
							},
							&cli.StringFlag{
								Name:  "type",
								Value: "parallel",
							},
							&cli.StringFlag{
								Name:  "format",
								Value: "normal",
							},
						},
						Action: doSpriteCreate,
					},
				},
			},

			{
				Name: "wad",
				Subcommands: []*cli.Command{
					{
						Name: "create",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "out",
								Required: true,
							},
						},
						Action: doWADCreate,
					},

					{
						Name: "extract",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "dir",
								Required: true,
							},
						},
						Action: doWADExtract,
					},

					{
						Name:   "info",
						Action: doWADInfo,
					},
				},
			},

			{
				Name: "bsp",
				Subcommands: []*cli.Command{
					{
						Name: "remap-materials",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name: "verbose",
							},
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
							},
						},
						Action: doBSPRemapMaterials,
					},

					{
						Name:   "info",
						Action: doBSPInfo,
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

	if source.IsEmpty() || replacement.IsEmpty() {
		return errors.New("no materials in source or replacement list")
	}

	bsp, err := goldsrc.LoadBSPFromFile(cCtx.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to load BSP: %w", err)
	}

	var (
		verbose  = cCtx.Bool("verbose")
		remapper = goldsrc.NewMaterialsRemapper(source)
	)
	mapping, err := remapper.ReMap(bsp.Textures.Textures, replacement)
	if err != nil {
		return fmt.Errorf("unable to remap materials: %w", err)
	}

	for i, tex := range bsp.Textures.Textures {
		mapTo, ok := mapping[tex.Name]
		if !ok {
			continue
		}

		if verbose {
			fmt.Printf(
				"Remapping %-15s to %s\n",
				strings.ToUpper(tex.Name.String()),
				strings.ToUpper(mapTo.String()),
			)
		}

		bsp.Textures.Textures[i].Name = mapTo
	}

	if err := bsp.WriteToFile(cCtx.String("out")); err != nil {
		return fmt.Errorf("unable to write BSP: %w", err)
	}

	if cCtx.Bool("verbose") {
		remapper.PrintAvailable()
	}

	return nil
}

func doBSPInfo(cCtx *cli.Context) error {
	bsp, err := goldsrc.LoadBSPFromFile(cCtx.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to load BSP: %w", err)
	}

	fmt.Print(bsp.String())

	return nil
}

//go:embed goldutil.1
var manPage string

func doHelp(cCtx *cli.Context) error {
	if runtime.GOOS == "windows" {
		return errors.New("man page is only available on *NIX operating systems, see https://l-p.github.io/goldutil/ instead")
	}

	var cmd = exec.Command("man", "-l", "-")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("unable to obtain man stdin: %w", err)
	}
	go func() {
		defer stdin.Close()
		if _, err := io.WriteString(stdin, manPage); err != nil {
			panic("unable to write to man stdin")
		}
	}()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unable to run man: %w", err)
	}

	return nil
}
