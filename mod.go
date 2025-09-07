package main

import (
	"fmt"
	"goldutil/goldsrc"
	"goldutil/set"
	"goldutil/wad"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

func doModFilterMaterials(cCtx *cli.Context) error {
	materials, err := goldsrc.LoadMaterialsFromFile(cCtx.String("in"))
	if err != nil {
		return fmt.Errorf("unable to load in: %w", err)
	}

	seen, err := getUsedTextureNames(cCtx.Args().Slice())
	if err != nil {
		return fmt.Errorf("unable to obtain textures list: %w", err)
	}

	for name, typ := range materials {
		if seen.Has(name) {
			fmt.Printf("%c %s\n", typ, name)
		}
	}

	return nil
}

func doModFilterWADs(cCtx *cli.Context) error {
	bspPaths, err := filepath.Glob(cCtx.String("bspdir") + "/*.bsp")
	if err != nil {
		return fmt.Errorf("unable to glob for BSP files: %w", err)
	}

	seen, err := getUsedTextureNames(bspPaths)
	if err != nil {
		return fmt.Errorf("unable to obtain textures list: %w", err)
	}

	output := wad.New()
	stored := set.NewPresenceSet[string](0)
	destPath, err := filepath.Abs(cCtx.String("out"))
	if err != nil {
		return fmt.Errorf("unable to obtain absolute output path: %w", err)
	}

	for i, rawPath := range cCtx.Args().Slice() {
		path, err := filepath.Abs(rawPath)
		if err != nil {
			return fmt.Errorf("unable to obtain absolute path for WAD at '%s' : %w", rawPath, err)
		}

		if path == destPath {
			fmt.Printf("Skipping WAD %d/%d at '%s' (output wad)\n", i+1, cCtx.Args().Len(), path)
			continue
		}

		fmt.Printf("Parsing WAD %d/%d at '%s'\n", i+1, cCtx.Args().Len(), path)
		wad3, err := wad.NewFromFile(path)
		if err != nil {
			return fmt.Errorf("unable to open WAD at '%s': %w", path, err)
		}

		for _, name := range wad3.Names() {
			if !seen.Has(name) {
				continue
			}

			if stored.Has(name) {
				continue
			}

			stored.Set(name)
			tex, ok := wad3.GetTexture(name)
			if !ok {
				return fmt.Errorf("unable to read texture '%s' from WAD at '%s'", name, path)
			}

			if err := output.AddTexture(tex); err != nil {
				return fmt.Errorf("unable to store texture '%s' to output WAD: %w", name, err)
			}
		}
	}

	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open '%s' for writing: %w", destPath, err)
	}

	if err := output.Write(dest); err != nil {
		return fmt.Errorf("unable to write to WAD file: %w", err)
	}

	if err := dest.Close(); err != nil {
		return fmt.Errorf("unable to finalize writing to '%s': %w", destPath, err)
	}

	fmt.Printf("Wrote %d textures to '%s'.\n", len(seen), destPath)

	return nil
}

func getUsedTextureNames(paths []string) (set.PresenceSet[string], error) {
	seen := set.NewPresenceSet[string](0)
	for _, path := range paths {
		bsp, err := goldsrc.LoadBSPFromFile(path)
		if err != nil {
			return nil, fmt.Errorf("unable to load BSP at '%s': %w", path, err)
		}

		for _, tex := range bsp.Textures.Textures {
			seen.Set(strings.ToUpper(tex.Name.String()))
		}
	}

	return seen, nil
}
