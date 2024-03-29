package main

import (
	"fmt"
	"goldutil/qmap"
	"goldutil/set"
	"strings"
)

func getUnexportedLayerSet(qm qmap.QMap) (set.PresenceSet[string], error) {
	var skipIDs = set.NewPresenceSet[string](0)

	for _, layer := range qm.GetTBLayers() {
		locked, ok := layer.GetProperty("_tb_layer_omit_from_export")
		if ok && locked == "1" {
			id, ok := layer.GetProperty("_tb_id")
			if !ok {
				return nil, fmt.Errorf("found a layer with no _tb_id")
			}
			skipIDs.Set(id)
		}
	}

	return skipIDs, nil
}

func getUnexportedGroupSet(qm qmap.QMap, unexportedLayerIDs set.PresenceSet[string]) (set.PresenceSet[string], error) {
	var skipIDs = set.NewPresenceSet[string](0)

	for {
		var foundMatches bool

		for _, group := range qm.GetTBGroups() {
			groupID, ok := group.GetProperty("_tb_id")
			if !ok {
				return nil, fmt.Errorf("found a group with no _tb_id")
			}

			parentGroupID, ok := group.GetProperty("_tb_group")
			if ok && !skipIDs.Has(groupID) && skipIDs.Has(parentGroupID) {
				skipIDs.Set(groupID)
				foundMatches = true
			}

			layerID, ok := group.GetProperty("_tb_layer")
			if ok && !skipIDs.Has(groupID) && unexportedLayerIDs.Has(layerID) {
				skipIDs.Set(groupID)
				foundMatches = true
			}
		}

		// Nested groups may not have their layer ID set, meaning we need to
		// recurse from the topmost group into all subgroups to propagate their
		// unexportedness. Do this by iterating until we find nothing new.
		// Not optimal but you'll reach the limits of your target engine before
		// this process ever gets long enough to be noticeable.
		if !foundMatches {
			break
		}
	}

	return skipIDs, nil
}

func exportQMap(qm qmap.QMap, cleanupTB bool) (qmap.QMap, error) {
	skipLayerIDs, err := getUnexportedLayerSet(qm)
	if err != nil {
		return qmap.QMap{}, err
	}

	skipGroupIDs, err := getUnexportedGroupSet(qm, skipLayerIDs)
	if err != nil {
		return qmap.QMap{}, err
	}

	var clean qmap.QMap
	for _, v := range qm.RawEntities() {
		layerID, ok := v.GetProperty("_tb_layer")
		if ok && skipLayerIDs.Has(layerID) {
			continue
		}

		groupID, ok := v.GetProperty("_tb_group")
		if ok && skipGroupIDs.Has(groupID) {
			continue
		}

		id, ok := v.GetProperty("_tb_id")
		if ok && v.Class() == "func_group" {
			if typ, ok := v.GetProperty("_tb_type"); ok {
				if typ == "_tb_group" && skipGroupIDs.Has(id) {
					continue
				}
			}

			if typ, ok := v.GetProperty("_tb_type"); ok {
				if typ == "_tb_layer" && skipLayerIDs.Has(id) {
					continue
				}
			}
		}

		if cleanupTB {
			removeTBProps(&v) //nolint:gosec // We don't keep the pointer.
		}

		clean.AddEntity(v)
	}

	return clean, nil
}

func removeTBProps(ent *qmap.Entity) {
	for k := range ent.PropertyMap() {
		if strings.HasPrefix(k, "_tb_") {
			ent.RemoveProperty(k)
		}
	}
}
