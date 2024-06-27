package main

import (
	"fmt"
	"goldutil/goldsrc"
	"goldutil/wad"
	"sort"
	"strings"
)

// TODO:
// 1. Decouple from BSP:
//   - Input texture names
//   - Output mappings
//   - Write to BSP outside of remapper
//
// 2. Tests
type remapper struct {
	// Available textures names we can remap to. Last entry will pop when we
	// use one.
	pools map[goldsrc.MaterialType][]string

	// Templated texture names for the 12-chars hack. See getTemplatedName().
	templates templateListSet

	// Next available index for templated texture names.
	tplIndices map[goldsrc.MaterialType][]int
}

// One list per texture prefix length.
type templateListSet [3]map[goldsrc.MaterialType][]materialTemplate

func generatePools(mats goldsrc.Materials) map[goldsrc.MaterialType][]string {
	var ret = make(map[goldsrc.MaterialType][]string, 10)

	for texture, material := range mats {
		ret[material] = append(ret[material], texture)
	}

	// Ensure determinism.
	for k := range ret {
		sort.Strings(ret[k])
	}

	return ret
}

type materialTemplate struct {
	tpl  string // contains a single
	uses int    // used as hex in tpl
}

func generateTemplates(mats goldsrc.Materials) templateListSet {
	var (
		ret       templateListSet
		formatTpl = "%%0%dx"
	)

	for prefixLen := 0; prefixLen <= 2; prefixLen++ {
		ret[prefixLen] = make(map[goldsrc.MaterialType][]materialTemplate, 10)

		for texture, material := range mats {
			// Only use exact matches to ensure we don't conflict with source
			// texture names unless with reuse one in getReusableTexture.
			if len(texture) != 12 {
				continue
			}

			ret[prefixLen][material] = append(
				ret[prefixLen][material],
				materialTemplate{
					tpl: texture[:12] + fmt.Sprintf(formatTpl, 3-prefixLen),
				},
			)
		}
	}

	return ret
}

func newRemapper(source goldsrc.Materials) remapper {
	return remapper{
		pools:      generatePools(source),
		templates:  generateTemplates(source),
		tplIndices: make(map[goldsrc.MaterialType][]int),
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

		var prefixLen = getTexturePrefixLength(name)
		mapToMat, ok := replacements[name[prefixLen:]]
		if !ok { // No remapping requested.
			continue
		}

		var reuse bool
		mapToName, ok := r.getTemplatedName(name, mapToMat)
		if !ok { // Material as no template-able texture name.
			mapToName, err = r.getReusableTexture(mapToMat)
			if err != nil {
				return err
			}
			reuse = true
		}

		// Re-apply prefix.
		mapToName = name[:prefixLen] + mapToName

		fmt.Printf("Remapping %-15s to %c %s", name, mapToMat, mapToName)
		if reuse {
			fmt.Println(" (reused existing name).")
		} else {
			fmt.Println(".")
		}
		bsp.Textures.Textures[i].Name, err = wad.NewTextureName(strings.ToLower(mapToName))
		if err != nil {
			return err
		}
	}

	r.printAvailable()

	return nil
}

func (r *remapper) printAvailable() {
	fmt.Println("\nTexture names still usable in source:")
	for mat, v := range r.pools {
		fmt.Printf("  - %-20s: %d\n", mat.String(), len(v))
	}

	var totals = map[goldsrc.MaterialType][3]int{}
	for prefixLen, set := range r.templates {
		for mat, list := range set {
			perPrefix := totals[mat]
			for _, v := range list {
				perPrefix[prefixLen] += maxUses(prefixLen) - v.uses
			}
			totals[mat] = perPrefix
		}
	}
	fmt.Println("\nTemplated texture entries still usable in source (unprefixed, one char, two chars):")
	for mat, v := range totals {
		fmt.Printf("  - %-20s: % 6d, % 6d, % 6d\n", mat.String(), v[0], v[1], v[2])
	}
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

func (r *remapper) getFirstAvailableTemplate(
	prefixLen int,
	mapToMat goldsrc.MaterialType,
) (int, materialTemplate, bool) {
	var max = maxUses(prefixLen)
	for i, v := range r.templates[prefixLen][mapToMat] {
		if v.uses < max {
			return i, v, true
		}
	}

	return -1, materialTemplate{}, false
}

func maxUses(prefixLen int) int {
	switch prefixLen {
	case 0:
		return 16 * 16 * 16
	case 1:
		return 16 * 16
	case 2:
		return 16
	default:
		panic("(prefixLen > 2)")
	}
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
func (r *remapper) getTemplatedName(name string, mapToMat goldsrc.MaterialType) (string, bool) {
	prefixLen := getTexturePrefixLength(name)
	index, tpl, ok := r.getFirstAvailableTemplate(prefixLen, mapToMat)
	if !ok {
		return "", false
	}

	mapToName := strings.ToUpper(fmt.Sprintf(tpl.tpl, tpl.uses))
	tpl.uses++
	r.templates[prefixLen][mapToMat][index] = tpl

	return mapToName, true
}

func getTexturePrefixLength(name string) int {
	if len(name) < 1 {
		return 0
	}
	switch name[0] {
	case '{', '!', '~', ' ':
		return 1
	}

	if name[0] == '-' || name[0] == '+' {
		return 2
	}

	return 0
}
