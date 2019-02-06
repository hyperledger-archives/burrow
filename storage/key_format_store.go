package storage

import (
	"fmt"
	"reflect"
)

var expectedKeyFormatType = reflect.TypeOf(MustKeyFormat{})

func EnsureKeyFormatStore(ks interface{}) error {
	rv := reflect.ValueOf(ks)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	rt := rv.Type()

	keyFormats := make(map[string]MustKeyFormat)
	for i := 0; i < rt.NumField(); i++ {
		fv := rv.Field(i)
		if fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				return fmt.Errorf("key format field '%s' is nil", rt.Field(i).Name)
			}
			fv = fv.Elem()
		}
		ft := fv.Type()
		if ft == expectedKeyFormatType {
			kf := fv.Interface().(MustKeyFormat)
			prefix := kf.Prefix().String()
			if kfDuplicate, ok := keyFormats[prefix]; ok {
				return fmt.Errorf("duplicate prefix %q between key format %v and %v",
					prefix, kfDuplicate, kf)
			}
			keyFormats[prefix] = kf
		}
	}
	return nil
}
