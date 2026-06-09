package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/L-P/goldutil/goldsrc/qmap"
)

func GraphQMap(qm *qmap.QMap, w io.Writer) {
	fmt.Fprintln(w, "digraph TB {")
	fmt.Fprintln(w, "  overlap = false;")

	aliases := make(map[string]string)
	var entNumber int

	for v := range qm.Entities() {
		entNumber++
		class := v.KVs["classname"]
		name, ok := v.KVs["targetname"]
		if !ok {
			name = fmt.Sprintf("_%s_%d", class, entNumber)
		}

		if class == "multi_manager" {
			graphMultiManager(v, w, aliases)
			continue
		}

		target, ok := v.KVs["target"]
		if ok && target != "" {
			if class == "trigger_relay" {
				graphTriggerRelayTarget(v, target, w, aliases)
			} else {
				printRelation(w, aliases, name, target, "")
			}
		}

		message, ok := v.KVs["message"]
		if ok && (class == "path_track" || class == "path_corner") {
			printRelation(w, aliases, name, message, "")
		}

		triggerTarget, ok := v.KVs["TriggerTarget"]
		if ok && triggerTarget != "" {
			condition, ok := v.KVs["TriggerCondition"]
			if !ok {
				condition = "0"
			}

			printRelation(w, aliases, name, triggerTarget, "cond:"+condition)
		}

		killTarget, ok := v.KVs["killtarget"]
		if ok && killTarget != "" {
			printRelation(w, aliases, name, killTarget, "kill")
		}

		master, ok := v.KVs["master"]
		if ok && master != "" {
			printRelation(w, aliases, name, master, "master")
		}
	}

	fmt.Fprintln(w, "}")
}

var mmIgnored = map[string]struct{}{
	"targetname": {},
	"angles":     {},
	"classname":  {},
	"origin":     {},
	"spawnflags": {},
}

func graphMultiManager(mm qmap.AnonymousEntity, w io.Writer, aliases map[string]string) {
	for target := range mm.KVs {
		if _, ok := mmIgnored[target]; ok {
			continue
		}

		if strings.HasPrefix(target, "_tb_") {
			continue
		}

		// HACK: TB adds an angle property to multi_manager belonging to linked groups.
		if target == "angle" {
			continue
		}

		printRelation(w, aliases, mm.KVs["targetname"], target, "")
	}
}

func graphTriggerRelayTarget(
	relay qmap.AnonymousEntity,
	target string,
	w io.Writer,
	aliases map[string]string,
) {
	name := relay.KVs["targetname"]
	state, ok := relay.KVs["triggerstate"]
	if !ok {
		state = "0"
	}

	switch state {
	case "1":
		printRelation(w, aliases, name, target, "on")
	case "2":
		printRelation(w, aliases, name, target, "toggle")
	default:
		printRelation(w, aliases, name, target, "off")
	}
}

func printRelation(
	w io.Writer,
	aliases map[string]string,
	from, to, label string,
) {
	from = aliasTargetname(w, aliases, canonicalTargetName(from))
	to = aliasTargetname(w, aliases, canonicalTargetName(to))

	if label != "" {
		fmt.Fprintf(w, "  %s -> %s [label=\"%s\"];\n", from, to, label)
	} else {
		fmt.Fprintf(w, "  %s -> %s;\n", from, to)
	}
}

func aliasTargetname(w io.Writer, aliases map[string]string, name string) string {
	alias, ok := aliases[name]
	if !ok {
		alias = fmt.Sprintf("_goldutil_%d", len(aliases))
		aliases[name] = alias
		fmt.Fprintf(w, "  %s [label=\"%s\"];\n", alias, name)
	}

	return alias
}

func canonicalTargetName(str string) string {
	// The pound is used to repeat names in multi_manager, it's ignored
	// somewhere in the engine (not the dlls).
	ret, _, _ := strings.Cut(str, "#")
	return ret
}
