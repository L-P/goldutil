package goldsrc

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

type MaterialType rune

const (
	MaterialTypeInvalid  MaterialType = 0
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

type Materials map[string]MaterialType // texture name => material type

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

	return ret
}

func LoadMaterials(r io.Reader) (Materials, error) {
	var (
		mats    = Materials(make(map[string]MaterialType))
		scanner = bufio.NewScanner(r)
	)

	var number int
	for scanner.Scan() {
		var line = scanner.Text()
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		typ, name, err := parseMaterialsLine(line)
		if err != nil {
			return nil, fmt.Errorf("unable to parse materials: line #%d: %w", number, err)
		}

		mats[name] = typ
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

		if !unicode.IsUpper(char) && !unicode.IsNumber(char) && char != '_' {
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
