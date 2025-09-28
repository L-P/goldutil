package main

import (
	"fmt"
	"goldutil/goldsrc/typedmap"
	"goldutil/set"
	"strings"
)

func getUnexportedLayerSet(tmap typedmap.TypedMap) (set.PresenceSet[string], error) {
	var skipIDs = set.NewPresenceSet[string](0)

	for _, layer := range tmap.FindByClassNameAndKV("func_group", "_tb_type", "_tb_layer") {
		locked, ok := layer.Entity.KVs["_tb_layer_omit_from_export"]
		if ok && locked == "1" {
			id, ok := layer.Entity.KVs["_tb_id"]
			if !ok {
				return nil, fmt.Errorf("found a layer with no _tb_id")
			}
			skipIDs.Set(id)
		}
	}

	return skipIDs, nil
}

func getUnexportedGroupSet(tmap typedmap.TypedMap, unexportedLayerIDs set.PresenceSet[string]) (set.PresenceSet[string], error) {
	var skipIDs = set.NewPresenceSet[string](0)

	for {
		var foundMatches bool

		for _, group := range tmap.FindByClassNameAndKV("func_group", "_tb_type", "_tb_group") {
			groupID, ok := group.Entity.KVs["_tb_id"]
			if !ok {
				return nil, fmt.Errorf("found a group with no _tb_id")
			}

			parentGroupID, ok := group.Entity.KVs["_tb_group"]
			if ok && !skipIDs.Has(groupID) && skipIDs.Has(parentGroupID) {
				skipIDs.Set(groupID)
				foundMatches = true
			}

			layerID, ok := group.Entity.KVs["_tb_layer"]
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

func exportTypedMap(tmap typedmap.TypedMap, cleanupTB bool) (typedmap.TypedMap, error) {
	skipLayerIDs, err := getUnexportedLayerSet(tmap)
	if err != nil {
		return typedmap.TypedMap{}, err
	}

	skipGroupIDs, err := getUnexportedGroupSet(tmap, skipLayerIDs)
	if err != nil {
		return typedmap.TypedMap{}, err
	}

	clean := typedmap.New()
	for _, v := range tmap {
		layerID, ok := v.KVs["_tb_layer"]
		if ok && skipLayerIDs.Has(layerID) {
			continue
		}

		groupID, ok := v.KVs["_tb_group"]
		if ok && skipGroupIDs.Has(groupID) {
			continue
		}

		id, ok := v.KVs["_tb_id"]
		if ok && v.KVs["classname"] == "func_group" {
			if typ, ok := v.KVs["_tb_type"]; ok {
				if typ == "_tb_group" && skipGroupIDs.Has(id) {
					continue
				}
			}

			if typ, ok := v.KVs["_tb_type"]; ok {
				if typ == "_tb_layer" && skipLayerIDs.Has(id) {
					continue
				}
			}
		}

		if cleanupTB {
			removeTBProps(&v)
		}

		clean.AddAnonymousEntities(v)
	}

	return clean, nil
}

func removeTBProps(ent *typedmap.AnonymousEntity) {
	for k := range ent.KVs {
		if strings.HasPrefix(k, "_tb_") {
			delete(ent.KVs, k)
		}
	}
}
