package typedmap_test

import (
	"goldutil/goldsrc/typedmap"
	"goldutil/goldsrc/typedmap/valve"
	"goldutil/nullable"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQMap(t *testing.T) {
	tmap, err := typedmap.LoadFromFile(
		"test.map",
		valve.GetEntityCollection(),
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := typedmap.TypedMap{
		typedmap.NewAnonymousEntity(
			map[string]string{
				"mapversion":  "220",
				"wad":         "",
				"classname":   "worldspawn",
				"sounds":      "1",
				"MaxRange":    "4096",
				"startdark":   "0",
				"gametitle":   "0",
				"newunit":     "0",
				"defaultteam": "0",
			},
		),
		&valve.MultiSource{
			ClassName:   nullable.New("multisource"),
			Origin:      "0 0 0",
			Target:      "mstarget",
			GlobalState: "msglobalstate",
			TargetName:  "mstargetname",
		},
		&valve.TriggerRelay{
			ClassName:    nullable.New("trigger_relay"),
			TriggerState: valve.TriggerStateOn,
			Origin:       "16 0 0",
			KillTarget:   "relaaykilltarget",
			Target:       "relaytarget",
			TargetName:   "relaytargetname",
		},
		&valve.ButtonTarget{
			ClassName:    nullable.New("button_target"),
			Flags:        1,
			RenderFX:     nullable.New[uint8](0),
			RenderMode:   nullable.New[valve.RenderMode](0),
			RenderAmount: nullable.New[uint8](0),
			RenderColor:  nullable.New[valve.Color]("0 0 0"),
			Master:       "btnmaster",
			Target:       "btntarget",
		},
	}

	require.ElementsMatch(t, expected, tmap)
}
