package goldsrc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

type MaterialType rune

const MaterialTypeInvalid MaterialType = 0

func (t MaterialType) String() string {
	if t == MaterialTypeInvalid {
		return "MaterialTypeInvalid"
	}

	return fmt.Sprintf("%c", t)
}

const MaterialTypeCount = 9 + 1

// materials.txt parser.
type Materials map[string]MaterialType // texture name => material type
func (m Materials) IsEmpty() bool {
	return len(m) == 0
}

func LoadMaterialsFromFile(path string) (Materials, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open materials file: %w", err)
	}
	defer f.Close() //nolint:errcheck // readonly

	return LoadMaterials(f)
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
		return MaterialTypeInvalid, "", errors.New("wrong number of fields")
	}

	if len(parts[0]) != 1 {
		return MaterialTypeInvalid, "", errors.New("material type field is invalid, expected a single uppercase letter")
	}

	mat := parseMaterialType(parts[0][0])
	if mat == MaterialTypeInvalid {
		return MaterialTypeInvalid, "", fmt.Errorf("invalid material type: %s", parts[0])
	}

	if !IsValidTextureName(parts[1]) {
		return MaterialTypeInvalid, "", fmt.Errorf("invalid texture name: %s", parts[1])
	}

	return mat, parts[1], nil
}

func IsValidTextureName(str string) bool {
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
	if c < 'A' || c > 'Z' {
		return MaterialTypeInvalid
	}

	return MaterialType(c)
}
