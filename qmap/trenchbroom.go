package qmap

// Returns TrenchBroom layer entities.
func (qm QMap) GetTBLayers() []Entity {
	var ret []Entity
	for _, v := range qm.entities {
		if v.Class() != "func_group" {
			continue
		}

		if typ, ok := v.GetProperty("_tb_type"); ok && typ == "_tb_layer" {
			ret = append(ret, v)
		}
	}

	return ret
}

func (qm QMap) GetTBGroups() []Entity {
	var ret []Entity
	for _, v := range qm.entities {
		if v.Class() != "func_group" {
			continue
		}

		if typ, ok := v.GetProperty("_tb_type"); ok && typ == "_tb_group" {
			ret = append(ret, v)
		}
	}

	return ret
}
