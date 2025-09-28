package main

import (
	"fmt"
	"goldutil/goldsrc/qmap"
	"goldutil/set"
	"strings"
)

func getUnexportedLayerSet(qm *qmap.QMap) (set.PresenceSet[string], error) {
	var skipIDs = set.NewPresenceSet[string](0)

	for _, layer := range qm.FindByClassNameAndKV("func_group", "_tb_type", "_tb_layer") {
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

func getUnexportedGroupSet(qm *qmap.QMap, unexportedLayerIDs set.PresenceSet[string]) (set.PresenceSet[string], error) {
	var skipIDs = set.NewPresenceSet[string](0)

	for {
		var foundMatches bool

		for _, group := range qm.FindByClassNameAndKV("func_group", "_tb_type", "_tb_group") {
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

func exportQMap(qm *qmap.QMap, cleanupTB bool) (*qmap.QMap, error) {
	skipLayerIDs, err := getUnexportedLayerSet(qm)
	if err != nil {
		return nil, err
	}

	skipGroupIDs, err := getUnexportedGroupSet(qm, skipLayerIDs)
	if err != nil {
		return nil, err
	}

	var clean []qmap.AnonymousEntity //nolint:prealloc
	for v := range qm.Entities() {
		if shouldSkipEntity(v, skipLayerIDs, skipGroupIDs) {
			continue
		}

		if cleanupTB {
			removeTBProps(v)
		}

		clean = append(clean, v)
	}

	out := qmap.New()
	if err := out.AddAnonymousEntities(clean...); err != nil {
		return nil, fmt.Errorf("unable to fill output qmap: %w", err)
	}

	return out, nil
}

func shouldSkipEntity(ent qmap.AnonymousEntity,
	skipLayerIDs set.PresenceSet[string],
	skipGroupIDs set.PresenceSet[string],
) bool {
	layerID, ok := ent.KVs["_tb_layer"]
	if ok && skipLayerIDs.Has(layerID) {
		return true
	}

	groupID, ok := ent.KVs["_tb_group"]
	if ok && skipGroupIDs.Has(groupID) {
		return true
	}

	id, ok := ent.KVs["_tb_id"]
	if ok && ent.KVs["classname"] == "func_group" {
		if typ, ok := ent.KVs["_tb_type"]; ok {
			if typ == "_tb_group" && skipGroupIDs.Has(id) {
				return true
			}
		}

		if typ, ok := ent.KVs["_tb_type"]; ok {
			if typ == "_tb_layer" && skipLayerIDs.Has(id) {
				return true
			}
		}
	}

	return false
}

func removeTBProps(ent qmap.AnonymousEntity) {
	for k := range ent.KVs {
		if strings.HasPrefix(k, "_tb_") {
			delete(ent.KVs, k)
		}
	}
}
