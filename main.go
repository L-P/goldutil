package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"goldutil/goldsrc"
	"goldutil/goldsrc/qmap"
	"goldutil/neat"
	"goldutil/set"
	"goldutil/sprite"
	"goldutil/wad"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"
)

var Version = "unknown version"

//go:embed goldutil.fgd
var fgd string

func main() {
	var app = newApp()
	if err := app.Run(context.Background(), os.Args); err != nil {
		// HACK: -h will panic.
		if err.Error() != "flag: help requested" {
			panic(err)
		}
	}
}

//nolint:funlen // descriptions
func newApp() *cli.Command {
	cli.HelpPrinter = func(w io.Writer, templ string, data any) {
		_ = doHelp(context.Background(), nil)
	}

	return &cli.Command{
		Version: Version,
		Commands: []*cli.Command{
			{
				Name:   "help",
				Action: doHelp,
			},
			{
				Name: "fgd",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Fprint(cmd.Writer, fgd)
					return nil
				},
			},
			{
				Name: "nod",
				Commands: []*cli.Command{
					{
						Name: "export",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name: "original-positions",
							},
							&cli.StringFlag{
								Name:  "input-format",
								Value: "valve",
							},
						},
						Action: doNodExport,
					},
				},
			},
			{
				Name: "mod",
				Commands: []*cli.Command{
					{
						Name: "filter-materials",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "in",
								Value: "sound/materials.full.txt",
							},
						},
						Action: doModFilterMaterials,
					},
					{
						Name: "filter-wads",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "bspdir",
								Value: "valve_addon/maps",
							},
							&cli.StringFlag{
								Name:  "out",
								Value: "valve_addon/filtered.wad",
							},
						},
						Action: doModFilterWADs,
					},
				},
			},
			{
				Name: "map",
				Commands: []*cli.Command{
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

					{
						Name:   "neat",
						Action: doNeat,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "moddir",
								Value: ".",
							},
						},
					},
				},
			},

			{
				Name: "spr",
				Commands: []*cli.Command{
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
				Commands: []*cli.Command{
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
				Commands: []*cli.Command{
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

func doSpriteExtract(ctx context.Context, cmd *cli.Command) error {
	path := cmd.Args().Get(0)
	if path == "" {
		return errors.New("expected one argument: the .spr to parse and extract")
	}

	spr, err := sprite.NewFromFile(path)
	if err != nil {
		return fmt.Errorf("unable to open sprite: %w", err)
	}

	return extractSprite(spr, cmd.String("dir"), filepath.Base(path))
}

func doSpriteCreate(ctx context.Context, cmd *cli.Command) error {
	typ, ok := map[string]sprite.Type{
		"parallel-upright":  sprite.ParallelUpright,
		"facing-upright":    sprite.FacingUpright,
		"parallel":          sprite.Parallel,
		"oriented":          sprite.Oriented,
		"parallel-oriented": sprite.ParallelOriented,
	}[cmd.String("type")]
	if !ok {
		return errors.New("unrecognize sprite type")
	}

	format, ok := map[string]sprite.TextureFormat{
		"normal":      sprite.Normal,
		"additive":    sprite.Additive,
		"index-alpha": sprite.IndexAlpha,
		"alpha-test":  sprite.AlphaTest,
	}[cmd.String("format")]
	if !ok {
		return errors.New("unrecognize texture format")
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

func doMapGraph(ctx context.Context, cmd *cli.Command) error {
	path := cmd.Args().Get(0)
	if path == "" {
		return errors.New("expected one argument: the .map to parse and graph")
	}

	qm, err := loadQMap(cmd.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to read from map: %w", err)
	}

	GraphQMap(qm, os.Stdout)

	return nil
}

func doNeat(ctx context.Context, cmd *cli.Command) error {
	qm, err := loadQMap(cmd.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to read from map: %w", err)
	}

	mod, err := os.OpenRoot(cmd.String("moddir"))
	if err != nil {
		return fmt.Errorf("unable to open current working directory: %w", err)
	}

	if err := neat.Neatify(qm, mod); err != nil {
		return fmt.Errorf("unable to neatify map: %w", err)
	}

	fmt.Fprint(cmd.Writer, qm.String())

	return nil
}

func doMapExport(ctx context.Context, cmd *cli.Command) error {
	qm, err := loadQMap(cmd.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to read from map: %w", err)
	}

	clean, err := exportQMap(qm, cmd.Bool("cleanup-tb"))
	if err != nil {
		return fmt.Errorf("unable to export map: %w", err)
	}

	fmt.Fprint(cmd.Writer, clean.String())

	return nil
}

func loadQMap(path string) (*qmap.QMap, error) {
	if path == "" {
		return qmap.LoadFromReader(os.Stdin)
	}

	return qmap.LoadFromFile(path)
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

func doBSPRemapMaterials(ctx context.Context, cmd *cli.Command) error {
	source, err := goldsrc.LoadMaterialsFromFile(cmd.String("original-materials"))
	if err != nil {
		return fmt.Errorf("unable to load original-materials: %w", err)
	}

	replacement, err := goldsrc.LoadMaterialsFromFile(cmd.String("replacement-materials"))
	if err != nil {
		return fmt.Errorf("unable to load replacement-materials: %w", err)
	}

	if source.IsEmpty() || replacement.IsEmpty() {
		return errors.New("no materials in source or replacement list")
	}

	bsp, err := goldsrc.LoadBSPFromFile(cmd.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to load BSP: %w", err)
	}

	var (
		verbose  = cmd.Bool("verbose")
		remapper = goldsrc.NewMaterialsRemapper(source)
	)
	mapping, err := remapper.ReMap(cmd.ErrWriter, bsp.Textures.Textures, replacement)
	if err != nil {
		return fmt.Errorf("unable to remap materials: %w", err)
	}

	for i, tex := range bsp.Textures.Textures {
		mapTo, ok := mapping[tex.Name]
		if !ok {
			continue
		}

		if verbose {
			fmt.Fprintf(
				cmd.Writer,
				"Remapping %-15s to %s\n",
				strings.ToUpper(tex.Name.String()),
				strings.ToUpper(mapTo.String()),
			)
		}

		bsp.Textures.Textures[i].Name = mapTo
	}

	if err := bsp.WriteToFile(cmd.String("out")); err != nil {
		return fmt.Errorf("unable to write BSP: %w", err)
	}

	if cmd.Bool("verbose") {
		remapper.PrintAvailable(cmd.Writer)
	}

	return nil
}

func doBSPInfo(ctx context.Context, cmd *cli.Command) error {
	bsp, err := goldsrc.LoadBSPFromFile(cmd.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to load BSP: %w", err)
	}

	fmt.Fprint(cmd.Writer, bsp.String())

	return nil
}

//go:embed goldutil.1
var manPage string

func doHelp(ctx context.Context, cmd *cli.Command) error {
	if runtime.GOOS == "windows" {
		return errors.New("man page is only available on *NIX operating systems, see https://l-p.github.io/goldutil/ instead")
	}

	var man = exec.CommandContext(ctx, "man", "-l", "-")

	stdin, err := man.StdinPipe()
	if err != nil {
		return fmt.Errorf("unable to obtain man stdin: %w", err)
	}
	go func() {
		defer stdin.Close() //nolint:errcheck // readonly
		if _, err := io.WriteString(stdin, manPage); err != nil {
			panic("unable to write to man stdin")
		}
	}()

	man.Stdout = os.Stdout
	man.Stderr = os.Stderr

	if err := man.Run(); err != nil {
		return fmt.Errorf("unable to run man: %w", err)
	}

	return nil
}

func doNodExport(ctx context.Context, cmd *cli.Command) error {
	format, ok := map[string]goldsrc.NodeFormat{
		"valve": goldsrc.NodeFormatValve,
		"decay": goldsrc.NodeFormatDecay,
	}[cmd.String("input-format")]
	if !ok {
		return errors.New("unrecognize .nod format")
	}

	f, err := os.Open(cmd.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to open file for reading: %w", err)
	}
	defer f.Close() //nolint:errcheck // readonly

	nodes, links, err := goldsrc.ReadNodes(f, format)
	if err != nil {
		return fmt.Errorf("unable to read nodes: %w", err)
	}

	original := cmd.Bool("original-positions")
	entities := make([]qmap.AnonymousEntity, 0, len(nodes)+len(links))
	for i, v := range nodes {
		entities = append(entities, qmap.AnonymousEntity{KVs: map[string]string{
			"classname":  v.ClassName(),
			"origin":     v.Position(original).String(),
			"targetname": fmt.Sprintf("node#%d", i),
		}})
	}

	for linkTypeBitID := range goldsrc.LinkTypeBitMax {
		entities = append(entities, qmap.AnonymousEntity{KVs: map[string]string{
			"classname":            "func_group",
			"_tb_type":             "_tb_layer",
			"_tb_name":             fmt.Sprintf("hull#%d links (%s)", linkTypeBitID, goldsrc.LinkTypeName(linkTypeBitID)),
			"_tb_id":               strconv.Itoa(linkTypeBitID + 1),
			"_tb_layer_sort_index": strconv.Itoa(linkTypeBitID + 1),
		}})

		for _, v := range links {
			if (v.LinkInfo & (1 << linkTypeBitID)) == 0 {
				continue
			}

			src := entities[v.SrcNode]
			src.KVs["target"] = fmt.Sprintf("node#%d", v.DstNode)
			src.KVs["_tb_layer"] = strconv.Itoa(linkTypeBitID + 1)
			entities = append(entities, src)
		}
	}

	out := qmap.New()
	if err := out.AddAnonymousEntities(entities...); err != nil {
		return fmt.Errorf("unable to append entities to output map: %w", err)
	}

	fmt.Fprintln(cmd.Writer, out.String())

	return nil
}
