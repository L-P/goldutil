package set

type PresenceSet[T comparable] map[T]struct{}

func NewPresenceSet[T comparable](capacity int) PresenceSet[T] {
	return make(map[T]struct{}, capacity)
}

func (ps PresenceSet[T]) Has(key T) bool {
	_, ok := ps[key]
	return ok
}

func (ps PresenceSet[T]) Set(key T) {
	ps[key] = struct{}{}
}
