//nolint:gofmt // BUG
package typedmap_test

import (
	"goldutil/goldsrc/typedmap"
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTypedMap(t *testing.T) { //nolint:funlen
	tmap, err := typedmap.LoadFromFile("test.map")
	if err != nil {
		t.Fatal(err)
	}

	expected := typedmap.New()
	require.NoError(t, expected.AddAnonymousEntities([]typedmap.AnonymousEntity{
		typedmap.AnonymousEntity{KVs: map[string]string{
			"mapversion":  "220",
			"wad":         "",
			"classname":   "worldspawn",
			"sounds":      "1",
			"MaxRange":    "4096",
			"startdark":   "0",
			"gametitle":   "0",
			"newunit":     "0",
			"defaultteam": "0",
		}},
		typedmap.AnonymousEntity{KVs: map[string]string{
			"classname":   "multisource",
			"origin":      "0 0 0",
			"target":      "mstarget",
			"globalstate": "msglobalstate",
			"targetname":  "mstargetname",
		}},
		typedmap.AnonymousEntity{KVs: map[string]string{
			"classname":    "trigger_relay",
			"spawnflags":   "0",
			"triggerstate": "1",
			"delay":        "0",
			"origin":       "16 0 0",
			"killtarget":   "relaaykilltarget",
			"target":       "relaytarget",
			"targetname":   "relaytargetname",
		}},
		typedmap.AnonymousEntity{
			KVs: map[string]string{
				"classname":   "button_target",
				"spawnflags":  "1",
				"renderfx":    "0",
				"rendermode":  "0",
				"rendercolor": "0 0 0",
				"renderamt":   "0",
				"master":      "btnmaster",
				"target":      "btntarget",
			},
			BrushEntity: typedmap.BrushEntity{
				Brushes: []typedmap.Brush{
					typedmap.Brush{
						"( 16 -16 0 ) ( 16 -15 0 ) ( 16 -16 1 ) __TB_empty [ 0 1 0 0 ] [ 0 0 -1 0 ] 0 1 1",
						"( 16 -16 0 ) ( 16 -16 1 ) ( 17 -16 0 ) __TB_empty [ 1 0 0 0 ] [ 0 0 -1 0 ] 0 1 1",
						"( 16 -16 0 ) ( 17 -16 0 ) ( 16 -15 0 ) __TB_empty [ 1 0 0 0 ] [ 0 -1 0 0 ] 0 1 1",
						"( 64 32 16 ) ( 64 33 16 ) ( 65 32 16 ) __TB_empty [ 1 0 0 0 ] [ 0 -1 0 0 ] 0 1 1",
						"( 64 0 16 ) ( 65 0 16 ) ( 64 0 17 ) __TB_empty [ 1 0 0 0 ] [ 0 0 -1 0 ] 0 1 1",
						"( 32 32 16 ) ( 32 32 17 ) ( 32 33 16 ) __TB_empty [ 0 1 0 0 ] [ 0 0 -1 0 ] 0 1 1",
					},
				},
			},
		},
	}))

	require.ElementsMatch(
		t,
		slices.Collect(maps.Values(expected)),
		slices.Collect(maps.Values(tmap)),
	)
}
