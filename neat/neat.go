package neat

import (
	"fmt"
	"goldutil/goldsrc/typedmap"
	"goldutil/goldsrc/typedmap/valve"
	"strings"

	"github.com/google/uuid"
)

func Neatify(tmap typedmap.TypedMap) error {
	if err := handleNeatMasters(tmap); err != nil {
		return fmt.Errorf("unable to handle neat_master: %w", err)
	}

	return nil
}

func handleNeatMasters(tmap typedmap.TypedMap) error {
	masters, err := typedmap.FindByKV[NeatMaster](tmap, "classname", "neat_master")
	if err != nil {
		return fmt.Errorf("unable to obtain neat_master entitites: %w", err)
	}

	for _, v := range masters {
		if err := handleNeatMaster(tmap, v.Index, v.Entity); err != nil {
			return err
		}
	}

	return nil
}

func handleNeatMaster(tmap typedmap.TypedMap, index uuid.UUID, master NeatMaster) error {
	if err := master.Validate(); err != nil {
		return err
	}
	tmap.Delete(index)

	for _, caller := range tmap.FindCallers(master.TargetName) {
		if caller.Entity.KVs["classname"] == "trigger_relay" {
			caller.Entity.KVs[caller.MatchedKey] = master.TargetName + "_proxy"
			continue
		}

		if caller.Entity.KVs["classname"] == "multi_manager" {
			_, suffix, hasSuffix := strings.Cut(caller.MatchedKey, "#")
			if hasSuffix {
				caller.Entity.KVs[master.TargetName+"_toggle#"+suffix] = caller.Entity.KVs[caller.MatchedKey]
			} else {
				caller.Entity.KVs[master.TargetName+"_toggle"] = caller.Entity.KVs[caller.MatchedKey]
			}
			delete(caller.Entity.KVs, caller.MatchedKey)
			continue
		}

		caller.Entity.KVs[caller.MatchedKey] = master.TargetName + "_toggle"
	}

	additions := getNeatMasterAdditions(master)
	if err := tmap.AddEntities(additions); err != nil {
		return fmt.Errorf("unable to append entities: %w", err)
	}

	return nil
}

func getNeatMasterAdditions(master NeatMaster) []any {
	return []any{
		valve.MultiSource{
			TargetName: master.TargetName,
			Target:     master.Target,
		},
		valve.ButtonTarget{
			TargetName: master.TargetName + "_proxy",
			Target:     master.TargetName,
		},
		valve.TriggerRelay{
			TargetName:   master.TargetName + "_on",
			Target:       master.TargetName + "_proxy",
			TriggerState: valve.TriggerStateOn,
		},
		valve.TriggerRelay{
			TargetName:   master.TargetName + "_off",
			Target:       master.TargetName + "_proxy",
			TriggerState: valve.TriggerStateOff,
		},
		valve.TriggerRelay{
			TargetName:   master.TargetName + "_toggle",
			Target:       master.TargetName + "_proxy",
			TriggerState: valve.TriggerStateToggle,
		},
	}
}
