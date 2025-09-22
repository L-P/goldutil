package qmap_test

import (
	"fmt"
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

func ExampleMarshal() {
	type Foo struct {
		RegularProperty string
		HardcodedValue  string `qmap:",hardcoded"`
		HardcodedKey    string `qmap:"some_other_key"`
		UntouchedValue  string

		privateProperty string
	}

	marshalled, _ := qmap.Marshal(Foo{
		RegularProperty: "regular",
		HardcodedValue:  "ignored",
		HardcodedKey:    "hardcoded key",
		privateProperty: "also ignored",
	})

	fmt.Print(marshalled)

	// Output:
	// {
	// "regular_property" "regular"
	// "hardcoded_value" "hardcoded"
	// "some_other_key" "hardcoded key"
	// "untouched_value" ""
	// }
}
