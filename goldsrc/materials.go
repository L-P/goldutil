package goldsrc

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode"
)

type MaterialType rune

const (
	MaterialTypeInvalid MaterialType = 0

	MaterialTypeComputer MaterialType = 'P'
	MaterialTypeDirt     MaterialType = 'D'
	MaterialTypeGlass    MaterialType = 'Y'
	MaterialTypeGrate    MaterialType = 'G'
	MaterialTypeLiquid   MaterialType = 'S' // "slosh"
	MaterialTypeMetal    MaterialType = 'M'
	MaterialTypeTile     MaterialType = 'T'
	MaterialTypeVents    MaterialType = 'V'
	MaterialTypeWood     MaterialType = 'W'
)

func (t MaterialType) String() string {
	switch t {
	case MaterialTypeComputer:
		return "MaterialTypeComputer"
	case MaterialTypeDirt:
		return "MaterialTypeDirt"
	case MaterialTypeGlass:
		return "MaterialTypeGlass"
	case MaterialTypeGrate:
		return "MaterialTypeGrate"
	case MaterialTypeLiquid:
		return "MaterialTypeLiquid"
	case MaterialTypeMetal:
		return "MaterialTypeMetal"
	case MaterialTypeTile:
		return "MaterialTypeTile"
	case MaterialTypeVents:
		return "MaterialTypeVents"
	case MaterialTypeWood:
		return "MaterialTypeWood"
	case MaterialTypeInvalid:
		return "MaterialTypeInvalid"
	}

	return fmt.Sprintf("<invalid: %d>", t)
}

const MaterialTypeCount = 9 + 1

type Materials map[string]MaterialType // texture name => material type
func (m Materials) IsEmpty() bool {
	return len(m) == 0
}

func LoadMaterialsFromFile(path string) (Materials, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open materials file: %w", err)
	}
	defer f.Close()

	return LoadMaterials(f)
}

func (mats Materials) Invert() map[MaterialType][]string {
	var ret = make(map[MaterialType][]string, 10)

	for texture, material := range mats {
		ret[material] = append(ret[material], texture)
	}

	// Ensure reproducibility in remap-materials.
	for k := range ret {
		sort.Strings(ret[k])
	}

	return ret
}

func (mats Materials) Templates() map[MaterialType]string {
	var ret = make(map[MaterialType]string, 10)

	for texture, material := range mats {
		if len(texture) != 12 {
			continue
		}

		ret[material] = texture[:12] + "%03d"
	}

	return ret
}

func LoadMaterials(r io.Reader) (Materials, error) {
	var (
		mats    = Materials(make(map[string]MaterialType))
		scanner = bufio.NewScanner(r)
	)

	var (
		lineNumber = -1
		entries    = 0
	)

	for scanner.Scan() {
		lineNumber++
		var line = scanner.Text()
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		typ, name, err := parseMaterialsLine(line)
		if err != nil {
			return nil, fmt.Errorf("unable to parse materials: line #%d: %w", lineNumber, err)
		}

		mats[name] = typ
		entries++
	}

	return mats, scanner.Err()
}

func parseMaterialsLine(line string) (MaterialType, string, error) {
	var parts = strings.SplitN(line, " ", 3)
	if len(parts) != 2 {
		return MaterialTypeInvalid, "", fmt.Errorf("wrong number of fields")
	}

	if len(parts[0]) != 1 {
		return MaterialTypeInvalid, "", fmt.Errorf("material type field is invalid, expected a single uppercase letter")
	}

	mat := parseMaterialType(parts[0][0])
	if mat == MaterialTypeInvalid {
		return MaterialTypeInvalid, "", fmt.Errorf("invalid material type: %s", parts[0])
	}

	if !isValidTextureName(parts[1]) {
		return MaterialTypeInvalid, "", fmt.Errorf("invalid texture name: %s", parts[1])
	}

	return mat, parts[1], nil
}

func isValidTextureName(str string) bool {
	if len(str) < 1 || len(str) > 15 {
		return false
	}

	for _, char := range str {
		if char > unicode.MaxASCII {
			return false
		}

		if !unicode.IsPrint(char) || unicode.IsSpace(char) {
			return false
		}
	}

	return true
}

func parseMaterialType(c byte) MaterialType {
	switch MaterialType(c) {
	case MaterialTypeComputer,
		MaterialTypeDirt,
		MaterialTypeGlass,
		MaterialTypeGrate,
		MaterialTypeLiquid,
		MaterialTypeMetal,
		MaterialTypeTile,
		MaterialTypeVents,
		MaterialTypeWood:
		return MaterialType(c)
	case MaterialTypeInvalid:
		return MaterialTypeInvalid
	}

	return MaterialTypeInvalid
}
