package main

import (
	"fmt"
	"goldutil/goldsrc"
	"goldutil/set"
	"strings"

	"github.com/urfave/cli/v2"
)

func doModFilterMaterials(cCtx *cli.Context) error {
	materials, err := goldsrc.LoadMaterialsFromFile(cCtx.String("in"))
	if err != nil {
		return fmt.Errorf("unable to load in: %w", err)
	}

	seen := set.NewPresenceSet[string](len(materials))

	for _, bspPath := range cCtx.Args().Slice() {
		bsp, err := goldsrc.LoadBSPFromFile(bspPath)
		if err != nil {
			return fmt.Errorf("unable to load BSP at '%s': %w", bspPath, err)
		}

		for _, tex := range bsp.Textures.Textures {
			seen.Set(strings.ToUpper(tex.Name.String()))
		}
	}

	for name, typ := range materials {
		if seen.Has(name) {
			fmt.Printf("%c %s\n", typ, name)
		}
	}

	return nil
}
