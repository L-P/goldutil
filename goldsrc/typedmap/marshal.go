package typedmap

import (
	"bytes"
	"fmt"
	"goldutil/goldsrc/typedmap/valve"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

const (
	TagName = "qmap"
)

// Converts a qmap tagged struct into an anonymous entity.
// Shortcuts were taken here.
func NewAnonymousEntityFromStruct(in any) (AnonymousEntity, error) {
	var zero AnonymousEntity

	marshalled, err := Marshal(in)
	if err != nil {
		return zero, fmt.Errorf("unable to marshall struct: %w", err)
	}

	tmap, err := LoadFromReader(bytes.NewReader(marshalled))
	if err != nil {
		return zero, fmt.Errorf("unable to parse entity back: %w", err)
	}

	return slices.Collect(maps.Values(tmap))[0], nil
}

// Marshals structs and pointer to structs into TypedMap entities.
// The qmap: field tags is of the form: property_name[,default_value].
func Marshal(in any) ([]byte, error) {
	typ := reflect.TypeOf(in)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	numFields := typ.NumField()
	var out bytes.Buffer

	fmt.Fprint(&out, "{\n")

	for i := 0; i < numFields; i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		propName, propDefault, hasDefault := strings.Cut(field.Tag.Get(TagName), ",")
		if propName == "" {
			propName = toSnakeCase(typ.Field(i).Name)
		}

		var propValue string
		var value reflect.Value
		if reflect.TypeOf(in).Kind() == reflect.Pointer {
			value = reflect.ValueOf(in).Elem().Field(i)
		} else {
			value = reflect.ValueOf(in).Field(i)
		}

		if value.Kind() == reflect.Pointer && value.IsNil() {
			if hasDefault {
				propValue = propDefault
			} else {
				continue
			}
		} else {
			propValue = toStringValue(reflect.Indirect(value).Interface())
		}

		if propValue == "" {
			continue
		}

		if strings.Contains(propName, `"`) {
			return nil, fmt.Errorf("property name cannot contain double-quotes: %s", propName)
		}

		if strings.Contains(propValue, `"`) {
			return nil, fmt.Errorf("property value cannot contain double-quotes: %s", propValue)
		}

		fmt.Fprintf(&out, `"%s" "%s"`, propName, propValue)
		out.WriteRune('\n')
	}

	fmt.Fprint(&out, "}\n")

	return out.Bytes(), nil
}

func toSnakeCase(in string) string {
	var b strings.Builder
	b.Grow(len(in))

	for i, c := range in {
		if unicode.IsUpper(c) && i != 0 {
			b.WriteRune('_')
		}

		b.WriteRune(unicode.ToLower(c))
	}

	return b.String()
}

func toStringValue(in any) string {
	switch v := in.(type) {
	case
		// HACK: Should I use reflection instead of bringing these types here?
		valve.RenderMode, valve.TriggerState,
		int, uint8:
		return fmt.Sprintf("%d", v)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return v
	default:
		return fmt.Sprintf("%s", v)
	}
}

func (ent *AnonymousEntity) UnmarshalInto(v any) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return fmt.Errorf("can only unmarshal into a pointer type, got %T", v)
	}

	destTyp := reflect.TypeOf(v).Elem()
	numFields := destTyp.NumField()

	for i := 0; i < numFields; i++ {
		field := destTyp.Field(i)
		if !field.IsExported() {
			continue
		}

		propName, _, _ := strings.Cut(field.Tag.Get(TagName), ",")
		if propName == "" {
			propName = toSnakeCase(destTyp.Field(i).Name)
		}

		propValue, ok := ent.KVs[propName]
		if !ok {
			continue
		}

		var dstValue reflect.Value
		if reflect.TypeOf(v).Kind() == reflect.Pointer {
			dstValue = reflect.ValueOf(v).Elem().Field(i)
		} else {
			dstValue = reflect.ValueOf(v).Field(i)
		}

		if dstValue.Kind() == reflect.Pointer {
			dstValue.Set(reflect.New(field.Type.Elem()))
			dstValue = dstValue.Elem()
		}

		if err := setReflectedValue(dstValue, propValue); err != nil {
			return fmt.Errorf("unable to set value to property %s on type %T: %w", field.Name, v, err)
		}
	}

	return nil
}

func setReflectedValue(dst reflect.Value, srcStr string) error {
	switch dst.Kind() { //nolint:exhaustive // that's why there's a default
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		srcInt, err := strconv.ParseInt(srcStr, 10, 64)
		if err != nil {
			return fmt.Errorf("unable to cast value to int: %w", err)
		}
		if dst.OverflowInt(srcInt) {
			return fmt.Errorf("value overflows destination type %s", dst.Kind().String())
		}
		dst.SetInt(srcInt)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		srcUint, err := strconv.ParseUint(srcStr, 10, 64)
		if err != nil {
			return fmt.Errorf("unable to cast value to int: %w", err)
		}
		if dst.OverflowUint(srcUint) {
			return fmt.Errorf("value overflows destination type %s", dst.Kind().String())
		}
		dst.SetUint(srcUint)
	case reflect.Float32, reflect.Float64:
		srcFloat, err := strconv.ParseFloat(srcStr, 64)
		if err != nil {
			return fmt.Errorf("unable to cast value to float: %w", err)
		}
		if dst.OverflowFloat(srcFloat) {
			return fmt.Errorf("value overflows destination type %s", dst.Kind().String())
		}
		dst.SetFloat(srcFloat)
	case reflect.String:
		dst.SetString(srcStr)
	default:
		return fmt.Errorf("unhandled type: %s", dst.Kind().String())
	}

	return nil
}
