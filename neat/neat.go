package neat

import (
	"fmt"
	"goldutil/goldsrc"
	"goldutil/goldsrc/typedmap"
	"goldutil/goldsrc/typedmap/valve"
	"os"
	"strings"

	"github.com/google/uuid"
)

func Neatify(tmap typedmap.TypedMap, mod *os.Root) error {
	if err := handleNeatMasters(tmap); err != nil {
		return fmt.Errorf("unable to handle neat_master: %w", err)
	}

	if err := handleNeatMessages(tmap, mod); err != nil {
		return fmt.Errorf("unable to handle neat_message: %w", err)
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
			Origin:     master.Origin,
			TargetName: master.TargetName,
			Target:     master.Target,
		},
		valve.ButtonTarget{
			Origin:     master.Origin,
			TargetName: master.TargetName + "_proxy",
			Target:     master.TargetName,
		},
		valve.TriggerRelay{
			Origin:       master.Origin,
			TargetName:   master.TargetName + "_on",
			Target:       master.TargetName + "_proxy",
			TriggerState: valve.TriggerStateOn,
		},
		valve.TriggerRelay{
			Origin:       master.Origin,
			TargetName:   master.TargetName + "_off",
			Target:       master.TargetName + "_proxy",
			TriggerState: valve.TriggerStateOff,
		},
		valve.TriggerRelay{
			Origin:       master.Origin,
			TargetName:   master.TargetName + "_toggle",
			Target:       master.TargetName + "_proxy",
			TriggerState: valve.TriggerStateToggle,
		},
	}
}

func handleNeatMessages(tmap typedmap.TypedMap, mod *os.Root) error {
	messages, err := typedmap.FindByKV[NeatMessage](tmap, "classname", "neat_message")
	if err != nil {
		return fmt.Errorf("unable to obtain neat_message entitites: %w", err)
	}

	titles, err := goldsrc.NewTitlesFromModRoot(mod)
	if err != nil {
		return fmt.Errorf("unable to parse titles.txt: %w", err)
	}
	for _, v := range messages {
		if err := handleNeatMessage(tmap, v.Index, v.Entity, titles); err != nil {
			return err
		}
	}

	return nil
}

func handleNeatMessage(
	tmap typedmap.TypedMap,
	index uuid.UUID,
	msg NeatMessage,
	titles map[string]goldsrc.Title,
) error {
	if err := msg.Validate(titles); err != nil {
		return err
	}
	tmap.Delete(index)

	return tmap.AddEntities([]any{
		valve.EnvMessage{
			Origin:      msg.Origin,
			TargetName:  msg.TargetName,
			Message:     msg.Message,
			Flags:       msg.Flags,
			Sound:       msg.Sound,
			Volume:      msg.Volume,
			Attenuation: msg.Attenuation,
		},
		valve.TriggerRelay{
			Origin:       msg.Origin,
			TargetName:   msg.TargetName,
			Delay:        titles[msg.Message].HoldTime + msg.Delay,
			Target:       msg.Target,
			TriggerState: msg.TriggerState,
		},
	})
}
