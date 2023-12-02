package qmap

// Returns TrenchBroom layer entities.
func (qm QMap) GetTBLayers() []Entity {
	var ret []Entity
	for _, v := range qm.entities {
		if typ, ok := v.GetProperty("_tb_type"); ok && typ == "_tb_layer" {
			ret = append(ret, v)
		}
	}

	return ret
}
