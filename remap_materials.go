package main

import (
	"fmt"
	"goldutil/goldsrc"
	"goldutil/wad"
	"strings"
)

type remapper struct {
	// Available textures names we can remap to. Last entry will pop when we
	// use one.
	pools map[goldsrc.MaterialType][]string

	// Templated texture names for the 12-chars hack. See getTemplate().
	templates map[goldsrc.MaterialType]string

	// Next available index for templated texture names.
	tplIndices map[goldsrc.MaterialType]int
}

func newRemapper(source goldsrc.Materials) remapper {
	return remapper{
		pools:      source.Invert(),
		templates:  source.Templates(),
		tplIndices: make(map[goldsrc.MaterialType]int),
	}
}

func (r *remapper) remap(bsp *goldsrc.BSP, replacements goldsrc.Materials) (err error) {
	for i, tex := range bsp.Textures.Textures {
		// Uppercase in materials, lowercase in BSP. Case is all over the place.
		var name = strings.ToUpper(tex.Name.String())

		if !tex.IsEmbedded() {
			fmt.Printf("texture #%d (%s) is not embedded, cannot safely remap", i, name)
			continue
		}

		mapToMat, ok := replacements[name]
		if !ok { // No remapping requested.
			continue
		}

		mapToName, ok := r.getTemplate(mapToMat)
		if !ok {
			mapToName, err = r.getReusableTexture(mapToMat)
			if err != nil {
				return err
			}
		}

		fmt.Printf("Remapping %-15s to %c %s.\n", name, mapToMat, mapToName)
		bsp.Textures.Textures[i].Name, err = wad.NewTextureName(strings.ToLower(mapToName))
		if err != nil {
			return err
		}
	}

	fmt.Println("\nTexture names still usable in source:")
	for mat, v := range r.pools {
		fmt.Printf("  - %s: %d\n", mat.String(), len(v))
	}

	return nil
}

// Returns a texture name we can reuse to give a specific material type to
// another texture.
func (r *remapper) getReusableTexture(mapToMat goldsrc.MaterialType) (string, error) {
	if len(r.pools[mapToMat]) == 0 {
		return "", fmt.Errorf("exhausted material pool for %s", mapToMat.String())
	}

	end := len(r.pools[mapToMat]) - 1
	mapToName := r.pools[mapToMat][end]
	r.pools[mapToMat] = r.pools[mapToMat][:end]

	return mapToName, nil
}

/* Returns a fmt template for a %d that generates a usable texture name
 * for the given material type.
 * This is a dirty hack to honor Hyrum's Law, the Half-Life materials.txt
 * system only uses the first 12 chars of a texture name to check for its
 * material type, this means _any_ string matching those first 12 chars will be
 * given the corresponding material.
 * In Half-Life most materials have textures with such names giving effectively
 * 576 slots for a given material _per texture name_. Greatly increasing the
 * 512 entry limit of the materials.txt and allowing to remap to more textures
 * than actually exists within this file.
 *
 * As a side-benefit, it limits risks of texture name collision which can
 * result in wrong texture reuse across level changes if the texture cached is
 * full. Cf.:
 *   - https://github.com/ValveSoftware/halflife/issues/102
 *   - https://github.com/ValveSoftware/halflife/issues/3102
 */
func (r *remapper) getTemplate(mapToMat goldsrc.MaterialType) (string, bool) {
	tpl, ok := r.templates[mapToMat]
	if !ok {
		return "", false
	}

	mapToName := strings.ToUpper(fmt.Sprintf(tpl, r.tplIndices[mapToMat]))
	r.tplIndices[mapToMat]++
	return mapToName, true
}
