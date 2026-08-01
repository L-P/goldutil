// Package neat implements an map pre-processor for GoldSrc.
package neat

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/L-P/goldutil/goldsrc"
	"github.com/L-P/goldutil/goldsrc/qmap"
	"github.com/L-P/goldutil/goldsrc/qmap/valve"
)

//go:embed goldutil.fgd
var FGD string

func Neatify(qm *qmap.QMap, mod *os.Root) error {
	if err := handleMasters(qm); err != nil {
		return fmt.Errorf("unable to handle neat_master: %w", err)
	}

	if err := handleMessages(qm, mod); err != nil {
		return fmt.Errorf("unable to handle neat_message: %w", err)
	}

	if err := handleSentences(qm, mod); err != nil {
		return fmt.Errorf("unable to handle neat_sentence: %w", err)
	}

	return nil
}

func handleMasters(qm *qmap.QMap) error {
	masters, err := qmap.FindByKV[Master](qm, "classname", "neat_master")
	if err != nil {
		return fmt.Errorf("unable to obtain neat_master entitites: %w", err)
	}

	for _, v := range masters {
		if err := handleMaster(qm, v.Index, v.Entity); err != nil {
			return err
		}
	}

	return nil
}

func handleMaster(qm *qmap.QMap, index uuid.UUID, master Master) error {
	if err := master.Validate(); err != nil {
		return err
	}
	qm.Delete(index)

	for _, caller := range qm.FindCallers(master.TargetName) {
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

	additions := getMasterAdditions(master)
	if err := qm.AddEntities(additions); err != nil {
		return fmt.Errorf("unable to append entities: %w", err)
	}

	return nil
}

func getMasterAdditions(master Master) []any {
	return []any{
		valve.MultiSource{
			Origin:      master.Origin,
			TargetName:  master.TargetName,
			Target:      master.Target,
			GlobalState: master.GlobalState,
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

func handleMessages(qm *qmap.QMap, mod *os.Root) error {
	messages, err := qmap.FindByKV[Message](qm, "classname", "neat_message")
	if err != nil {
		return fmt.Errorf("unable to obtain neat_message entitites: %w", err)
	}

	titles, err := goldsrc.NewTitlesFromModRoot(mod)
	if err != nil {
		return fmt.Errorf("unable to parse titles.txt: %w", err)
	}
	for _, v := range messages {
		if err := handleMessage(qm, v.Index, v.Entity, titles); err != nil {
			return err
		}
	}

	return nil
}

func handleMessage(
	qm *qmap.QMap,
	index uuid.UUID,
	msg Message,
	titles map[string]goldsrc.Title,
) error {
	if err := msg.Validate(titles); err != nil {
		return err
	}
	qm.Delete(index)

	return qm.AddEntities([]any{
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
			KillTarget:   msg.KillTarget,
			TriggerState: msg.TriggerState,
		},
	})
}

func handleSentences(qm *qmap.QMap, mod *os.Root) error {
	sentences, err := qmap.FindByKV[Sentence](qm, "classname", "neat_sentence")
	if err != nil {
		return fmt.Errorf("unable to obtain neat_sentence entitites: %w", err)
	}

	titles, err := goldsrc.NewTitlesFromModRoot(mod)
	if err != nil {
		return fmt.Errorf("unable to parse titles.txt: %w", err)
	}
	for _, v := range sentences {
		if err := handleSentence(qm, v.Index, v.Entity, titles); err != nil {
			return err
		}
	}

	return nil
}

func handleSentence(
	qm *qmap.QMap,
	index uuid.UUID,
	sentence Sentence,
	titles map[string]goldsrc.Title,
) error {
	if err := sentence.Validate(titles); err != nil {
		return err
	}
	qm.Delete(index)

	return qm.AddEntities([]any{
		valve.EnvMessage{
			Origin:     sentence.Origin,
			TargetName: sentence.TargetName,
			Message:    sentence.Sentence,
			Flags:      sentence.MessageFlags,
		},
		valve.ScriptedSentence{
			Origin:     sentence.Origin,
			TargetName: sentence.TargetName,

			Flags:       sentence.SentenceFlags,
			Sentence:    sentence.Sentence,
			Entity:      sentence.Entity,
			Listener:    sentence.Listener,
			Duration:    titles[sentence.Sentence].HoldTime + sentence.Duration,
			Radius:      sentence.Radius,
			Refire:      sentence.Refire,
			Volume:      sentence.Volume,
			Attenuation: sentence.Attenuation,
		},
		valve.TriggerRelay{
			Origin:       sentence.Origin,
			TargetName:   sentence.TargetName,
			Delay:        titles[sentence.Sentence].HoldTime + sentence.Delay,
			Target:       sentence.Target,
			KillTarget:   sentence.KillTarget,
			TriggerState: sentence.TriggerState,
		},
	})
}
