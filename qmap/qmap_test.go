package qmap_test

import (
	"goldutil/qmap"
	"reflect"
	"testing"
)

func TestQMap(t *testing.T) {
	qm, err := qmap.LoadFromFile("test.map")
	if err != nil {
		t.Fatal(err)
	}

	expected := qmap.Stats{
		NumEntities: 397,
		NumProps:    3310, // 📱
		NumBrushes:  0,
		NumPlanes:   0,
	}
	actual := qm.ComputeStats()

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("mismatched map stats\nexpected: %#v\ngot: %#v\n", expected, actual)
	}
}
