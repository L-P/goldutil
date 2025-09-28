package main

import (
	"fmt"
	"goldutil/goldsrc/qmap"
	"io"
	"strings"
)

func GraphQMap(qm qmap.QMap, w io.Writer) {
	fmt.Fprintln(w, "digraph TB {")
	fmt.Fprintln(w, "  overlap = false;")

	for _, v := range qm {
		name := v.KVs["targetname"]
		class := v.KVs["classname"]

		if class == "multi_manager" {
			graphMultiManager(v, w)
			continue
		}

		target, ok := v.KVs["target"]
		if ok && target != "" {
			if class == "trigger_relay" {
				graphTriggerRelayTarget(v, target, w)
			} else {
				fmt.Fprintf(w, "  %s -> %s;\n", name, target)
			}
		}

		message, ok := v.KVs["message"]
		if ok && (class == "path_track" || class == "path_corner") {
			fmt.Fprintf(w, "  %s -> %s;\n", name, message)
		}

		triggerTarget, ok := v.KVs["TriggerTarget"]
		if ok && triggerTarget != "" {
			condition, ok := v.KVs["TriggerCondition"]
			if !ok {
				condition = "0"
			}

			fmt.Fprintf(w, "  %s -> %s [label=\"cond:%s\"];\n", name, triggerTarget, condition)
		}

		killTarget, ok := v.KVs["killtarget"]
		if ok && killTarget != "" {
			fmt.Fprintf(w, "  %s -> %s [label=\"kill\"];\n", name, killTarget)
		}

		master, ok := v.KVs["master"]
		if ok && master != "" {
			fmt.Fprintf(w, "  %s -> %s [label=\"master\"];\n", name, master)
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

func graphMultiManager(mm qmap.AnonymousEntity, w io.Writer) {
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

		fmt.Fprintf(w, "  %s -> %s;\n", mm.KVs["targetname"], target)
	}
}

func graphTriggerRelayTarget(relay qmap.AnonymousEntity, target string, w io.Writer) {
	name := relay.KVs["targetname"]
	state, ok := relay.KVs["triggerstate"]
	if !ok {
		state = "0"
	}

	switch state {
	case "1":
		fmt.Fprintf(w, "  %s -> %s [label=\"on\"];\n", name, target)
	case "2":
		fmt.Fprintf(w, "  %s -> %s [label=\"toggle\"];\n", name, target)
	default:
		fmt.Fprintf(w, "  %s -> %s [label=\"off\"];\n", name, target)
	}
}
