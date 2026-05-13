package main

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v3"

	"github.com/L-P/goldutil/neat"
)

var Version = "unknown version"

func main() {
	var app = newApp()
	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Println(err.Error()) //nolint:forbidigo
		os.Exit(1)
	}
}

//nolint:funlen,maintidx,lll // descriptions
func newApp() *cli.Command {
	bold := color.New(color.Bold).Sprint
	catnl := func(str ...string) string {
		return strings.Join(str, "\n")
	}

	return &cli.Command{
		Version: Version,
		Usage:   "GoldSrc modding utilities.",
		Description: catnl(
			"goldutil can read, modify, and write multiple file formats used by the GoldSrc (Half-Life) engine.",
			"See more detailed help with `goldutil CMD -h` or `goldutil CMD SUBCMD -h`.",
		),

		Commands: []*cli.Command{
			{
				Name:  "bsp",
				Usage: "BSP (compiled maps) manipulation.",
				Commands: []*cli.Command{
					{
						Name:   "entities",
						Action: doBSPEntities,
						Usage:  "Print raw entity data from a BSP.",
					},
					{
						Name:   "info",
						Action: doBSPInfo,
						Usage:  "Print parsed data from a BSP.",
					},
					{
						Name:   "limits",
						Action: doBSPLimits,
						Usage:  "Show how much more details you can cram into your map.",
						Description: catnl(
							"Show how much more details you can cram into your map. These limits are sometimes hard limits of the BSP format, sometimes the engine, sometimes strong suggestions.",
							"They were taken from VHLT which is the de-facto standard.",
							"Exit with status code `1` if the BSP goes over a limit.",
						),
					},
					{
						Name: "remap-materials",
						Description: catnl(
							"On a BSP with embedded textures, change their names so they can match what is in the original game materials.txt.",
							"This allows setting proper material sounds to custom textures without having to distribute a materials.txt file.",
							bold("Warning")+": The BSP cannot use any of the textures listed in the original materials.txt",
						),

						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "original-materials",
								Value:    "valve/sound/materials.txt",
								Usage:    "Path to the materials.txt file of the original game, defaults to valve/sound/materials.txt.",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "out",
								Usage:    "Where to write the remapped BSP.",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "replacement-materials",
								Value:    "valve_addon/sound/materials.txt",
								Usage:    "Path to the replacement materials.txt file, defaults to valve_addon/sound/materials.txt.",
								Required: true,
							},
							&cli.BoolFlag{
								Name:  "verbose",
								Usage: "Output to STDOUT the details of what materials were remapped to.",
							},
						},
						Action: doBSPRemapMaterials,
					},
				},
			},

			{
				Name:  "fgd",
				Usage: "Output the FGD to use with goldutil map neat.",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Fprint(cmd.Writer, neat.FGD)
					return nil
				},
			},

			{
				Name:  "nod",
				Usage: "NPC pathfinding nodes manipulation.",
				Commands: []*cli.Command{
					{
						Name:  "export",
						Usage: "Extract nodes from a .nod file.",
						Description: catnl(
							"Extract node positions from a .nod graph into a .map populated with corresponding info_node entities.",
							"Links between nodes are represented using target/targetname, nodes are duplicated to allow showing all links, TrenchBroom layers are used to separate links by hull type. The resulting .map file is not for engine consumption, only for TrenchBroom-assisted archaeology.",
						),
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "input-format",
								Value:       "valve",
								HideDefault: true,
								Usage: catnl(
									"Parse the .nod file using a different node graph format instead of using the PC release format.\n`FORMAT` can be any one of:",
									"  - valve: Standard Half-Life node graph (default)",
									"  - decay: PlayStation 2 release of Half-Life: Decay",
								),
								Validator: func(str string) error {
									values := []string{"valve", "decay"}
									if slices.Index(values, str) < 0 {
										return fmt.Errorf("must be one of: %s", strings.Join(values, ", "))
									}

									return nil
								},
							},
							&cli.BoolFlag{
								Name:  "original-positions",
								Usage: "Use the node positions as they were set in the original .map instead of their position after being dropped to the ground during graph generation.",
							},
						},
						Action: doNodExport,
					},
				},
			},

			{
				Name:  "map",
				Usage: "Map pre-processing.",
				Commands: []*cli.Command{
					{
						Name:  "export",
						Usage: "Export a .map file the way TrenchBroom does.",
						Description: catnl(
							"Export a .map file the way TrenchBroom does, removing all layers marked as not exported.",
							"Output is written to STDOUT, if no FILE is provided the map will be read from standard input.",
						),
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "cleanup-tb",
								Usage: "Also remove properties added by TrenchBroom that are not understood by the engine and spam the console with errors.",
								Value: false,
							},
						},
						Action: doMapExport,
					},

					{
						Name:   "graph",
						Action: doMapGraph,
						Usage:  "Create a graphviz digraph of entity caller/callee relationships.",
						Description: catnl(
							"Create a graphviz digraph of entity caller/callee relationships from a .map file.",
							"ripent exports use the same format and can be read too. Output is written to STDOUT.",
						),
					},

					{
						Name:   "neat",
						Action: doNeat,
						Usage:  "Process neat_ entity macros and outputs the generated .map to standard output.",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "moddir",
								Value: ".",
								Usage: "root of the mod directory (eg. 'valve'), defaults to the current working directory.",
							},
						},
					},
				},
			},

			{
				Name:  "mod",
				Usage: "Misc modding utilities",
				Commands: []*cli.Command{
					{
						Name:  "filter-materials",
						Usage: "Filter unused materials out of materials.txt.",
						Description: catnl(
							"Takes a materials.txt file and only keep the texture names that are used in the given BSP files.",
							"This is useful to keep a final materials.txt under 512 entries when working with large texture collections.",
							"Filtered materials are written to STDOUT.",
						),
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "in",
								Value: "sound/materials.full.txt",
								Usage: "Path to the materials.txt file you want to filter.",
							},
						},
						Action: doModFilterMaterials,
					},
					{
						Name:  "filter-wads",
						Usage: "Filter unused textures out of WADs.",
						Description: catnl(
							"Reads all BSP files at the given directory and creates a WAD containing only the textures used by the BSPs.",
							"This allows using large texture collections during development but only distribute the smallest possible WAD at release time.",
							`Be aware that Half-Life requires the wads from the "wads" property to be present and will attempt to load them.`,
							`Before release you should generate the output WAD, remove the WADs that were used during filtering from your worldspawn "wads" property, add the output WAD, and then rebuild your maps.`,
						),
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "bspdir",
								Value: "valve_addon/maps",
								Usage: "Path of the directory containing the BSPs to use as a used texture list.",
							},
							&cli.StringFlag{
								Name:  "out",
								Value: "valve_addon/filtered.wad",
								Usage: "Path were the output WAD will be written.",
							},
						},
						Action: doModFilterWADs,
					},
				},
			},

			{
				Name:  "spr",
				Usage: "Sprite manipulation.",
				Commands: []*cli.Command{
					{
						Name:  "create",
						Usage: "Create a sprite.",
						Description: catnl(
							"Create a sprite from the given ordered list of PNG frames and write it to the given output path.\n",
							"Input images must be 256 colors paletted PNGs, the palette of the first frame will be used, the other palettes are discarded and all frames will be interpreted using the first frame's palette.",
							"If the palette has under 256 colors it will be extended to 256, putting the last color of the palette in the 256th spot and remapping the image to match this updated palette. This matters for transparent formats.",
							"If you use pngquant(1) to create your palletized input files, you can use its --pngbug option to ensure the transparent color will always be last.",
						),
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "format",
								Value: "normal",
								Usage: catnl(
									"Texture format, determines how the palette is interpreted and the texture is rendered by the engine. `FORMAT` can be any one of:",
									"  - normal: 256 colors sprite (default).",
									"  - additive: Additive 256 colors sprite, dark values are rendered as transparent, the darker the less opacity.",
									"  - index-alpha: Monochromatic sprite with 255 alpha levels, the base color is determined by the last color on the palette.",
									"  - alpha-test: Transparent 255 colors sprite. The last color on the palette will be rendered as fully transparent.",
								),
								HideDefault: true,
							},
							&cli.StringFlag{
								Name:     "out",
								Required: true,
								Usage:    "Path to the output .spr file.",
							},
							&cli.StringFlag{
								Name:  "type",
								Value: "parallel",
								Usage: catnl(
									"Sprite type, TYPE can be any one of:",
									"  - parallel: Always face camera (default).",
									"  - parallel-upright: Always face camera except for the locked Z axis.",
									"  - oriented: Orientation set by the level.",
									"  - parallel-oriented: Faces camera but can be rotated by the level.",
									"  - facing-upright: Like parallel-upright but faces the player origin instead of the camera.",
								),
								HideDefault: true,
							},
						},
						Action: doSpriteCreate,
					},

					{
						Name:        "extract",
						Usage:       "Output all frames of a sprite to the current directory.",
						Description: "The output files will be named after the original sprite file name plus a frame number suffix and an extension.",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "dir",
								Usage: "Output frames to the specified directory instead of the current one.",
							},
							&cli.BoolFlag{
								Name: "no-alpha",
								Usage: catnl(
									"Don't add an alpha channel and keep the original sprite palette verbatim.",
									"By default goldutil rewrites the palette to include both color and alpha on index-alpha and alpha-test sprites to make them appear as they do in the engine, if you want to leave the original palette untouched use this flag.",
								),
							},
						},
						Action: doSpriteExtract,
					},

					{
						Name:   "info",
						Action: doSpriteInfo,
						Usage:  "Print parsed frame data from the given FILE.",
					},
				},
			},

			{
				Name:  "wad",
				Usage: "Texture files manipulation.",
				Commands: []*cli.Command{
					{
						Name:  "create",
						Usage: "Create a WAD file.",
						Description: catnl(
							"Create a WAD file from a list of PNG files and directories. Directories are not scanned recursively and only PNG files are used.",
							"File base names (without extensions) are uppercased and used as texture names.",
							bold("Warning")+": Names exceeding 15 chars will trigger an error as this is the maximum length supported by the WAD format.",
							"decals.wad use a different texture format and are not handled by goldutil.",
						),
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "out",
								Required: true,
								Usage:    "Path to the output .wad file.",
							},
						},
						Action: doWADCreate,
					},

					{
						Name:  "extract",
						Usage: "Extract a WAD file in the given DIR as a bunch of PNG files.",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "dir",
								Required: true,
								Usage:    "Path to the directory where to write PNG files.",
							},
							&cli.BoolFlag{
								Name: "no-alpha",
								Usage: catnl(
									"Don't add an alpha channel and keep the original textures palette verbatim.",
									"By default goldutil rewrites the palette to add an alpha channel on transparent texture (those that start with a '{').",
								),
							},
						},
						Action: doWADExtract,
					},

					{
						Name:   "info",
						Action: doWADInfo,
						Usage:  "Print parsed data from a WAD file.",
					},
				},
			},

			{
				Name:  "wav",
				Usage: "Audio manipulation.",
				Commands: []*cli.Command{
					{
						Name:   "loop",
						Usage:  "Set CUE points to make a WAV loop.",
						Action: doWAVLoop,
						Description: catnl(
							"Make a WAV loop by setting CUE points.",
							"In GoldSrc only the presence of these CUE points is checked, not their position.",
							"This commands adds a CUE point at the end so ambient_generic can do its work.",
						),
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "out",
								Usage:    "Where to write the looped WAV.",
								Required: true,
							},
						},
					},
				},
			},
		},
	}
}
