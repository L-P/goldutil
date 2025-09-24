package typedmap_test

import (
	"fmt"
	"goldutil/goldsrc/typedmap"
	"goldutil/nullable"
	"testing"

	"github.com/stretchr/testify/require"
)

func ExampleMarshal() {
	type Foo struct {
		RegularProperty   string
		DefaultedValue    *string `qmap:",defaulted"`
		OverriddenDefault *string `qmap:",defaulted"`
		HardcodedKey      string  `qmap:"some_other_key"`
		EmptyString       string

		privateProperty string
	}

	overridden := "overridden"

	marshalled, _ := typedmap.Marshal(Foo{
		RegularProperty:   "regular",
		OverriddenDefault: &overridden,
		HardcodedKey:      "hardcoded key",
		privateProperty:   "also ignored",
	})

	fmt.Printf("%s", marshalled)

	// Output:
	// {
	// "regular_property" "regular"
	// "defaulted_value" "defaulted"
	// "overridden_default" "overridden"
	// "some_other_key" "hardcoded key"
	// }
}

func TestUnmarshal(t *testing.T) {
	type Foo struct {
		RegularProperty            string
		DefaultedValue             *string `qmap:",defaulted"`
		OverriddenDefault          *string `qmap:",defaulted"`
		HardcodedKey               string  `qmap:"some_other_key"`
		SurpriseInteger            int
		SurpriseIntegerButUnsigned uint8
		EmptyString                string
		FloatVal                   float32

		privateProperty string //nolint:unused
	}

	expected := Foo{
		RegularProperty:            "regular",
		OverriddenDefault:          nullable.New("overridden"),
		DefaultedValue:             nullable.New("non default"),
		HardcodedKey:               "hardcoded key",
		SurpriseInteger:            42,
		SurpriseIntegerButUnsigned: 255,
		FloatVal:                   0.1234,
	}

	input := typedmap.NewAnonymousEntity(map[string]string{
		"regular_property":              "regular",
		"defaulted_value":               "non default",
		"overridden_default":            "overridden",
		"some_other_key":                "hardcoded key",
		"surprise_integer":              "42",
		"surprise_integer_but_unsigned": "255",
		"float_val":                     "0.1234",
	})

	var dest Foo
	require.NoError(t, input.UnmarshalInto(&dest))
	require.Equal(t, expected, dest)
}
