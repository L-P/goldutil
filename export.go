package main

import (
	"fmt"
	"goldutil/qmap"
	"strings"
)

type set map[string]struct{}

func getUnexportedLayerSet(qm qmap.QMap) (set, error) {
	var skipIDs = make(set)
	for _, layer := range qm.GetTBLayers() {
		locked, ok := layer.GetProperty("_tb_layer_omit_from_export")
		if ok && locked == "1" {
			id, ok := layer.GetProperty("_tb_id")
			if !ok {
				return nil, fmt.Errorf("found a layer with no _tb_id")
			}
			skipIDs[id] = struct{}{}
		}
	}

	return skipIDs, nil
}

func getUnexportedGroupSet(qm qmap.QMap, unexportedLayerIDs set) (set, error) {
	var skipIDs = make(set)
	for _, group := range qm.GetTBGroups() {
		groupID, ok := group.GetProperty("_tb_id")
		if !ok {
			return nil, fmt.Errorf("found a group with no _tb_id")
		}

		layerID, ok := group.GetProperty("_tb_layer")
		if _, skip := unexportedLayerIDs[layerID]; ok && skip {
			skipIDs[groupID] = struct{}{}
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
		if _, skip := skipLayerIDs[layerID]; ok && skip {
			continue
		}

		groupID, ok := v.GetProperty("_tb_group")
		if _, skip := skipGroupIDs[groupID]; ok && skip {
			continue
		}

		id, ok := v.GetProperty("_tb_id")
		if ok && v.Class() == "func_group" {
			if typ, ok := v.GetProperty("_tb_type"); ok {
				if _, ok := skipGroupIDs[id]; ok && typ == "_tb_group" {
					continue
				}
			}

			if typ, ok := v.GetProperty("_tb_type"); ok {
				if _, ok := skipLayerIDs[id]; ok && typ == "_tb_layer" {
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
