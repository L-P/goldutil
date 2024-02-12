package set_test

import (
	"testing"

	"goldutil/set"

	"github.com/stretchr/testify/assert"
)

func TestPresenceSetString(t *testing.T) {
	var ps = set.NewPresenceSet[string](0)

	assert.False(t, ps.Has(""), "missing key doesn't exist")
	assert.False(t, ps.Has("foo"), "missing key doesn't exist")

	ps.Set("")
	assert.True(t, ps.Has(""), "empty key can be set")
	assert.False(t, ps.Has("foo"), "missing key still doesn't exist")

	ps.Set("foo")
	assert.True(t, ps.Has("foo"), "set key exists")

	assert.False(t, ps.Has("bar"), "missing key doesn't exist")
	settingBarOnPresenceSetAsValue(ps)
	assert.True(t, ps.Has("bar"), "key has been set on map given as value")
}

func settingBarOnPresenceSetAsValue(ps set.PresenceSet[string]) {
	ps.Set("bar")
}
