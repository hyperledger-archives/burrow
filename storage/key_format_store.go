// Copyright 2019 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
