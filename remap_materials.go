package main

import (
	"errors"
	"fmt"
	"goldutil/goldsrc"
	"goldutil/wad"
	"strings"
)

func remapBSPMaterials(
	bsp *goldsrc.BSP,
	source, replacement goldsrc.Materials,
) error {
	if source.IsEmpty() || replacement.IsEmpty() {
		return errors.New("no materials in source or replacement list")
	}

	// Pool of assignable texture names, each map texture will expend one entry.
	var (
		pools = source.Invert()
		err   error
	)

	for i, tex := range bsp.Textures.Textures {
		// Uppercase in materials, lowercase in BSP. Case is all over the place.
		var name = strings.ToUpper(tex.Name.String())

		if !tex.IsEmbedded() {
			fmt.Printf("texture #%d (%s) is not embedded, cannot safely remap", i, name)
			continue
		}

		mapToMat, ok := replacement[name]
		if !ok {
			continue
		}

		if len(pools[mapToMat]) == 0 {
			return fmt.Errorf("exhausted material pool for %s", mapToMat.String())
		}

		end := len(pools[mapToMat]) - 1
		mapToName := pools[mapToMat][end]
		pools[mapToMat] = pools[mapToMat][:end]

		fmt.Printf("Remapping %-15s to %s.\n", name, mapToName)
		bsp.Textures.Textures[i].Name, err = wad.NewTextureName(strings.ToLower(mapToName))
		if err != nil {
			return err
		}
	}

	fmt.Println("\nTexture names still usable in source:")
	for mat, v := range pools {
		fmt.Printf("  - %s: %d\n", mat.String(), len(v))
	}

	return nil
}
