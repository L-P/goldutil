package qmap_test

import (
	"fmt"
	"github.com/L-P/goldutil/goldsrc/qmap"
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

	marshalled, _ := qmap.Marshal(Foo{
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

type AliasedStringType string

type Foo struct {
	String                               string
	AliasedString                        AliasedStringType
	NullableString                       *string
	NullableStringWithDefault            *string `qmap:",default"`
	OverriddenNullableStringWithDefault  *string `qmap:",default"`
	Int                                  int
	NullableInt                          *int
	NullableIntWithDefault               *int `qmap:",42"`
	OverriddenNullableIntWithDefault     *int `qmap:",42"`
	Byte                                 uint8
	NullableByte                         *uint8
	NullableByteWithDefault              *uint8 `qmap:",42"`
	OverriddenNullableByteWithDefault    *uint8 `qmap:",42"`
	Float32                              float32
	NullableFloat32                      *float32
	NullableFloat32WithDefault           *float32 `qmap:",4.2"`
	OverriddenNullableFloat32WithDefault *float32 `qmap:",4.2"`

	privateProperty string //nolint:unused
}

func TestUnmarshal(t *testing.T) {
	input := qmap.AnonymousEntity{KVs: map[string]string{
		"string":          "string",
		"aliased_string":  "aliased",
		"nullable_string": "string",
		"overridden_nullable_string_with_default": "string",
		"int":                                   "51966",
		"nullable_int":                          "51966",
		"overridden_nullable_int_with_default":  "51966",
		"byte":                                  "202",
		"nullable_byte":                         "202",
		"overridden_nullable_byte_with_default": "202",
		"float32":                               "-1",
		"nullable_float32":                      "-1",
		"overridden_nullable_float32_with_default": "-1",
	}}

	// Defaults are not loaded when unmarshaling, only when marshaling.
	// TODO: Challenge this. I feel like there should be no default values
	// anywhere to ensure we're not overriding any behavior.
	expected := Foo{
		String:                               "string",
		AliasedString:                        "aliased",
		NullableString:                       new("string"),
		OverriddenNullableStringWithDefault:  new("string"),
		Int:                                  0xCAFE,
		NullableInt:                          new(int(0xCAFE)),
		OverriddenNullableIntWithDefault:     new(int(0xCAFE)),
		Byte:                                 0xCA,
		NullableByte:                         new(uint8(0xCA)),
		OverriddenNullableByteWithDefault:    new(uint8(0xCA)),
		Float32:                              -1,
		NullableFloat32:                      new(float32(-1)),
		OverriddenNullableFloat32WithDefault: new(float32(-1)),

		// NullableStringWithDefault:            new("default"),
		// NullableIntWithDefault:               new([int](42),
		// NullableByteWithDefault:              new([uint8](42),
		// NullableFloat32WithDefault:           new([float32](4.2),
	}

	var actual Foo
	require.NoError(t, input.UnmarshalInto(&actual))
	require.Equal(t, expected, actual)
}

func TestMarshal(t *testing.T) {
	actual, err := qmap.Marshal(Foo{
		AliasedString:                       "aliased",
		NullableString:                      new("string"),
		NullableStringWithDefault:           new("default"),
		OverriddenNullableStringWithDefault: new("string"),
		String:                              "string",

		Int:                              0xCAFE,
		NullableInt:                      new(int(0xCAFE)),
		NullableIntWithDefault:           new(int(42)),
		OverriddenNullableIntWithDefault: new(int(0xCAFE)),

		Byte:                              0xCA,
		NullableByte:                      new(uint8(0xCA)),
		NullableByteWithDefault:           new(uint8(42)),
		OverriddenNullableByteWithDefault: new(uint8(0xCA)),

		Float32:                              -1,
		NullableFloat32:                      new(float32(-1)),
		NullableFloat32WithDefault:           new(float32(4.2)),
		OverriddenNullableFloat32WithDefault: new(float32(-1)),
	})
	require.NoError(t, err)

	expected := `{
"string" "string"
"aliased_string" "aliased"
"nullable_string" "string"
"nullable_string_with_default" "default"
"overridden_nullable_string_with_default" "string"
"int" "51966"
"nullable_int" "51966"
"nullable_int_with_default" "42"
"overridden_nullable_int_with_default" "51966"
"byte" "202"
"nullable_byte" "202"
"nullable_byte_with_default" "42"
"overridden_nullable_byte_with_default" "202"
"float32" "-1"
"nullable_float32" "-1"
"nullable_float32_with_default" "4.2"
"overridden_nullable_float32_with_default" "-1"
}
`

	require.Equal(t, expected, string(actual))
}

func TestRoundTrip(t *testing.T) {
	expected := Foo{
		NullableString:                      new("string"),
		NullableStringWithDefault:           new("default"),
		OverriddenNullableStringWithDefault: new("string"),
		String:                              "string",

		Int:                              0xCAFE,
		NullableInt:                      new(int(0xCAFE)),
		NullableIntWithDefault:           new(int(42)),
		OverriddenNullableIntWithDefault: new(int(0xCAFE)),

		Byte:                              0xCA,
		NullableByte:                      new(uint8(0xCA)),
		NullableByteWithDefault:           new(uint8(42)),
		OverriddenNullableByteWithDefault: new(uint8(0xCA)),

		Float32:                              -1,
		NullableFloat32:                      new(float32(-1)),
		NullableFloat32WithDefault:           new(float32(4.2)),
		OverriddenNullableFloat32WithDefault: new(float32(-1)),
	}

	anonymous, err := qmap.NewAnonymousEntityFromStruct(expected)
	require.NoError(t, err)

	var reparsed Foo
	require.NoError(t, anonymous.UnmarshalInto(&reparsed))

	require.Equal(t, expected, reparsed)
}
