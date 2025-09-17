package main

import (
	"fmt"
	"goldutil/qmap"
	"io"
	"strings"
)

func GraphQMap(qm qmap.QMap, w io.Writer) {
	fmt.Fprintln(w, "digraph TB {")
	fmt.Fprintln(w, "  overlap = false;")

	for _, v := range qm.RawEntities() {
		name := v.Name()
		class := v.Class()

		if class == "multi_manager" {
			graphMultiManager(v, w)
			continue
		}

		target, ok := v.GetProperty(qmap.KTarget)
		if ok && target != "" {
			if class == "trigger_relay" {
				graphTriggerRelayTarget(v, target, w)
			} else {
				fmt.Fprintf(w, "  %s -> %s;\n", name, target)
			}
		}

		message, ok := v.GetProperty(qmap.KMessage)
		if ok && class == "path_track" {
			fmt.Fprintf(w, "  %s -> %s;\n", name, message)
		}

		triggerTarget, ok := v.GetProperty(qmap.KTriggerTarget)
		if ok && triggerTarget != "" {
			condition, ok := v.GetProperty(qmap.KTriggerCondition)
			if !ok {
				condition = "0"
			}

			fmt.Fprintf(w, "  %s -> %s [label=\"cond:%s\"];\n", name, triggerTarget, condition)
		}

		killTarget, ok := v.GetProperty(qmap.KKillTarget)
		if ok && killTarget != "" {
			fmt.Fprintf(w, "  %s -> %s [label=\"kill\"];\n", name, killTarget)
		}

		master, ok := v.GetProperty(qmap.KMaster)
		if ok && master != "" {
			fmt.Fprintf(w, "  %s -> %s [label=\"master\"];\n", name, master)
		}
	}

	fmt.Fprintln(w, "}")
}

var mmIgnored = map[string]struct{}{
	qmap.KName:   {},
	qmap.KAngles: {},
	qmap.KClass:  {},
	qmap.KOrigin: {},
	qmap.KFlags:  {},
}

func graphMultiManager(mm qmap.Entity, w io.Writer) {
	for target := range mm.PropertyMap() {
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

		fmt.Fprintf(w, "  %s -> %s;\n", mm.Name(), target)
	}
}

func graphTriggerRelayTarget(relay qmap.Entity, target string, w io.Writer) {
	name := relay.Name()
	state, ok := relay.GetProperty("triggerstate")
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
