package query

import (
	"fmt"
	"reflect"
)

type ReflectTagged struct {
	rv   reflect.Value
	keys []string
	ks   map[string]struct{}
}

var _ Tagged = &ReflectTagged{}

func MustReflectTags(value interface{}, fieldNames ...string) *ReflectTagged {
	rt, err := ReflectTags(value, fieldNames...)
	if err != nil {
		panic(err)
	}
	return rt
}

// ReflectTags provides a query.Tagged on a structs exported fields using query.StringFromValue to derive the string
// values associated with each field. If passed explicit field names will only only provide those fields as tags,
// otherwise all exported fields are provided.
func ReflectTags(value interface{}, fieldNames ...string) (*ReflectTagged, error) {
	rv := reflect.ValueOf(value)
	if rv.IsNil() {
		return &ReflectTagged{}, nil
	}
	if rv.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("ReflectStructTags needs a pointer to a struct but %v is not a pointer",
			rv.Interface())
	}
	if rv.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("ReflectStructTags needs a pointer to a struct but %v does not point to a struct",
			rv.Interface())
	}
	ty := rv.Elem().Type()

	numField := ty.NumField()
	if len(fieldNames) > 0 {
		if len(fieldNames) > numField {
			return nil, fmt.Errorf("ReflectTags asked to tag %v fields but %v only has %v fields",
				len(fieldNames), rv.Interface(), numField)
		}
		numField = len(fieldNames)
	}
	rt := &ReflectTagged{
		rv:   rv,
		ks:   make(map[string]struct{}, numField),
		keys: make([]string, 0, numField),
	}
	if len(fieldNames) > 0 {
		for _, fieldName := range fieldNames {
			field, ok := ty.FieldByName(fieldName)
			if !ok {
				return nil, fmt.Errorf("ReflectTags asked to tag field named %s by no such field on %v",
					fieldName, rv.Interface())
			}
			ok = rt.registerField(field)
			if !ok {
				return nil, fmt.Errorf("field %s of %v is not exported so cannot act as tag", fieldName,
					rv.Interface())
			}
		}
	} else {
		for i := 0; i < numField; i++ {
			rt.registerField(ty.Field(i))
		}
	}
	return rt, nil
}

func (rt *ReflectTagged) registerField(field reflect.StructField) (ok bool) {
	// Empty iff struct field is exported
	if field.PkgPath == "" {
		rt.keys = append(rt.keys, field.Name)
		rt.ks[field.Name] = struct{}{}
		return true
	}
	return false
}

func (rt *ReflectTagged) Keys() []string {
	return rt.keys
}

func (rt *ReflectTagged) Get(key string) (value string, ok bool) {
	if _, ok := rt.ks[key]; ok {
		return StringFromValue(rt.rv.Elem().FieldByName(key).Interface()), true
	}
	return "", false
}

func (rt *ReflectTagged) Len() int {
	return len(rt.keys)
}
