package goldsrc_test

import (
	"goldutil/goldsrc"
	"goldutil/wad"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaterialsRemapper(t *testing.T) {
	var (
		source = goldsrc.Materials(map[string]goldsrc.MaterialType{
			"SRC_DIRT":     'D',
			"SRC_GRATE":    'G',
			"SRC_METAL":    'M',
			"SRC_COMP":     'P',
			"SRC_LIQUID":   'S',
			"SRC_TILE":     'T',
			"SRC_VENTS":    'V',
			"SRC_WOOD":     'W',
			"SRC_GLASS":    'Y',
			"SRC_DIRTXXXX": 'D',
			"SRC_GRATEXXX": 'G',
			"SRC_METALXXX": 'M',
			"SRC_COMPXXXX": 'P',
			"SRC_LIQUIDXX": 'S',
			"SRC_TILEXXXX": 'T',
			"SRC_VENTSXXX": 'V',
			// /* skip to force verbatim reuse */ "SRC_WOODXXXX": 'W',
			// /* skip to force verbatim reuse */ "SRC_GLASSXXX": 'Y',
		})

		replacement = goldsrc.Materials(map[string]goldsrc.MaterialType{
			"REPL_METAL1": 'M',
			"REPL_METAL2": 'M',
			"REPL_GRATE1": 'G',
			"REPL_GRATE2": 'G',
			"REPL_GRATE3": 'G',
			"+0REPL_COMP": 'P',
			"REPL_WOOD":   'W',
			"REPL_GRASS":  'Y',
		})

		textures = mustTextureSlice(t, []string{
			"repl_metal1",
			"repl_metal2",
			"{repl_grate1",
			"{repl_grate2",
			"repl_grate3",
			"repl_wood",
		})

		remapper = goldsrc.NewMaterialsRemapper(source)

		expected = map[wad.TextureName]wad.TextureName{
			mustTextureName(t, "repl_metal1"):  mustTextureName(t, "src_metalxxx000"),
			mustTextureName(t, "repl_metal2"):  mustTextureName(t, "src_metalxxx001"),
			mustTextureName(t, "{repl_grate1"): mustTextureName(t, "{src_gratexxx00"),
			mustTextureName(t, "{repl_grate2"): mustTextureName(t, "{src_gratexxx01"),
			mustTextureName(t, "repl_grate3"):  mustTextureName(t, "src_gratexxx000"),
			mustTextureName(t, "repl_wood"):    mustTextureName(t, "src_wood"),
		}
	)

	actual, err := remapper.ReMap(textures, replacement)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func mustTextureSlice(t *testing.T, names []string) []wad.MIPTexture {
	var ret = make([]wad.MIPTexture, len(names))
	for i, name := range names {
		ret[i] = wad.MIPTexture{
			MIPTextureHeader: wad.MIPTextureHeader{
				Name:   mustTextureName(t, name),
				Width:  512,
				Height: 512,

				// Needs to be not zero to avoid being flagged as not embedded.
				MIPOffsets: [wad.NumMIPMaps]int32{1, 1, 1, 1},
			},
		}
	}

	return ret
}

func mustTextureName(t *testing.T, str string) wad.TextureName {
	ret, err := wad.NewTextureName(str)
	require.NoError(t, err)
	return ret
}

func TestMaterialsRemapperExhaust(t *testing.T) {
	var (
		source = goldsrc.Materials(map[string]goldsrc.MaterialType{
			"SRC_DIRT": 'D',
		})

		replacement = goldsrc.Materials(map[string]goldsrc.MaterialType{
			"REPL_DIRT1": 'D',
			"REPL_DIRT2": 'D',
		})

		textures = mustTextureSlice(t, []string{
			"repl_dirt1",
			"repl_dirt2",
		})

		remapper = goldsrc.NewMaterialsRemapper(source)
	)

	_, err := remapper.ReMap(textures, replacement)
	require.Error(t, err, "exhausted material pool for MaterialTypeDirt")
}
