package typedmap

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

func (tmap *TypedMap) FindByKV(key, value string) []SearchResult[AnonymousEntity] {
	var out []SearchResult[AnonymousEntity]

	for index, ent := range *tmap {
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

func (tmap *TypedMap) FindByClassNameAndKV(className, key, value string) []SearchResult[AnonymousEntity] {
	var out []SearchResult[AnonymousEntity]

	for index, ent := range *tmap {
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

func FindByKV[T any](tmap TypedMap, key, value string) ([]SearchResult[T], error) {
	var out []SearchResult[T] //nolint:prealloc // unknowable

	for index, ent := range tmap {
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
func (tmap *TypedMap) FindCallers(callee string) []SearchResult[AnonymousEntity] {
	out := slices.Concat(
		tmap.FindByKV("target", callee),
		tmap.FindByKV("killtarget", callee),
		tmap.FindByKV("TriggerTarget", callee), // monster_*
		tmap.FindByClassNameAndKV("trigger_changetarget", "m_iszNewTarget", callee),
		tmap.FindByClassNameAndKV("path_track", "message", callee),
		tmap.FindByClassNameAndKV("path_corner", "message", callee),
	)

	for _, mm := range tmap.FindByKV("classname", "multi_manager") {
		for key := range mm.Entity.KVs {
			if key == callee || strings.HasPrefix(key, callee+"#") {
				mm.MatchedKey = key
				out = append(out, mm)
			}
		}
	}

	return out
}
