/*
Copyright IBM Corp. 2017 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Field is a field of an arbitrary struct
type Field struct {
	Name  string
	Path  string
	Type  reflect.Type
	Kind  reflect.Kind
	Leaf  bool
	Depth int
	Tag   reflect.StructTag
	Value interface{}
	Addr  interface{}
}

// ParseObj parses an object structure, calling back with field info
// for each field
func ParseObj(obj interface{}, cb func(*Field) error) error {
	if cb == nil {
		return errors.New("nil callback")
	}
	return parse(obj, cb, nil)
}

func parse(ptr interface{}, cb func(*Field) error, parent *Field) error {
	var path string
	var depth int
	v := reflect.ValueOf(ptr).Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		vf := v.Field(i)
		tf := t.Field(i)
		name := strings.ToLower(tf.Name)
		if tf.Name[0] == name[0] {
			continue // skip unexported fields
		}
		if parent != nil {
			path = fmt.Sprintf("%s.%s", parent.Path, name)
			depth = parent.Depth + 1
		} else {
			path = name
		}
		kind := vf.Kind()
		leaf := kind != reflect.Struct && kind != reflect.Ptr
		field := &Field{
			Name:  name,
			Path:  path,
			Type:  tf.Type,
			Kind:  kind,
			Leaf:  leaf,
			Depth: depth,
			Tag:   tf.Tag,
			Value: vf.Interface(),
			Addr:  vf.Addr().Interface(),
		}
		err := cb(field)
		if err != nil {
			return err
		}
		if kind == reflect.Struct {
			// Skip parsing the entire struct if "skip" tag is present on a struct field
			if tf.Tag.Get(TagSkip) == "true" {
				continue
			}
			err := parse(field.Addr, cb, field)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CopyMissingValues checks the dst interface for missing values and
// replaces them with value from src config struct.
// This does a deep copy of pointers.
func CopyMissingValues(src, dst interface{}) {
	s := reflect.ValueOf(src).Elem()
	d := reflect.ValueOf(dst).Elem()
	copyMissingValues(s, d)
}

func copyMissingValues(src, dst reflect.Value) {
	if !src.IsValid() {
		return
	}
	switch src.Kind() {
	case reflect.Ptr:
		src = src.Elem()
		if !src.IsValid() {
			return
		}
		if dst.IsNil() {
			dst.Set(reflect.New(src.Type()))
		}
		copyMissingValues(src, dst.Elem())
	case reflect.Interface:
		if src.IsNil() {
			return
		}
		src = src.Elem()
		if dst.IsNil() {
			newVal := reflect.New(src.Type()).Elem()
			copyMissingValues(src, newVal)
			dst.Set(newVal)
		} else {
			copyMissingValues(src, dst.Elem())
		}
	case reflect.Struct:
		if !src.IsValid() {
			return
		}
		t, ok := src.Interface().(time.Time)
		if ok {
			dst.Set(reflect.ValueOf(t))
		}
		for i := 0; i < src.NumField(); i++ {
			copyMissingValues(src.Field(i), dst.Field(i))
		}
	case reflect.Slice:
		if !dst.IsNil() {
			return
		}
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			copyMissingValues(src.Index(i), dst.Index(i))
		}
	case reflect.Map:
		if dst.IsNil() {
			dst.Set(reflect.MakeMap(src.Type()))
		}
		for _, key := range src.MapKeys() {
			sval := src.MapIndex(key)
			dval := dst.MapIndex(key)
			copy := !dval.IsValid()
			if copy {
				dval = reflect.New(sval.Type()).Elem()
			}
			copyMissingValues(sval, dval)
			if copy {
				dst.SetMapIndex(key, dval)
			}
		}
	default:
		if !dst.CanInterface() {
			return
		}
		dval := dst.Interface()
		zval := reflect.Zero(dst.Type()).Interface()
		if reflect.DeepEqual(dval, zval) {
			dst.Set(src)
		}
	}
}
