package qmap

import (
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
)

type SearchResult[T any] struct {
	Index      uuid.UUID
	Entity     T
	MatchedKey string
}

func (qm *QMap) FindByKV(key, value string) []SearchResult[AnonymousEntity] {
	var out []SearchResult[AnonymousEntity]

	for index, ent := range qm.entities {
		propValue, ok := ent.KVs[key]
		if ok && propValue == value {
			out = append(out, SearchResult[AnonymousEntity]{
				Index:      index,
				Entity:     ent,
				MatchedKey: key,
			})
		}
	}

	return out
}

func (qm *QMap) FindByClassNameAndKV(className, key, value string) []SearchResult[AnonymousEntity] {
	var out []SearchResult[AnonymousEntity]

	for index, ent := range qm.entities {
		if ent.KVs["classname"] != className {
			continue
		}

		propValue, ok := ent.KVs[key]
		if ok && propValue == value {
			out = append(out, SearchResult[AnonymousEntity]{
				Index:      index,
				Entity:     ent,
				MatchedKey: key,
			})
		}
	}

	return out
}

func FindByKV[T any](qm *QMap, key, value string) ([]SearchResult[T], error) {
	var out []SearchResult[T] //nolint:prealloc // unknowable

	for index, ent := range qm.entities {
		propValue, ok := ent.KVs[key]
		if !ok || propValue != value {
			continue
		}

		var dst T
		if err := ent.UnmarshalInto(&dst); err != nil {
			return nil, fmt.Errorf("unable to unmarshal: %w", err)
		}

		out = append(out, SearchResult[T]{
			Index:      index,
			Entity:     dst,
			MatchedKey: key,
		})
	}

	return out, nil
}

// Returns all entities targeting a given targetname.
// A single entity can appear multiple times using different targeting means.
func (qm *QMap) FindCallers(callee string) []SearchResult[AnonymousEntity] {
	out := slices.Concat(
		qm.FindByKV("target", callee),
		qm.FindByKV("killtarget", callee),
		qm.FindByKV("TriggerTarget", callee), // monster_*
		qm.FindByClassNameAndKV("trigger_changetarget", "m_iszNewTarget", callee),
		qm.FindByClassNameAndKV("path_track", "message", callee),
		qm.FindByClassNameAndKV("path_corner", "message", callee),
	)

	for _, mm := range qm.FindByKV("classname", "multi_manager") {
		for key := range mm.Entity.KVs {
			if key == callee || strings.HasPrefix(key, callee+"#") {
				mm.MatchedKey = key
				out = append(out, mm)
			}
		}
	}

	return out
}
