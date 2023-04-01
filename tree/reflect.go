// Copyright (c) 2019-2023 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tree

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func getTags(f reflect.StructField) (name string, options map[string]interface{}) {
	name = f.Name
	options = map[string]interface{}{}

	// Check for the newer conf tag.
	tags, ok := f.Tag.Lookup("conf")

	// If there is no conf tag, check for the old cdb tag.
	if !ok {
		tags, ok = f.Tag.Lookup("cdb")
	}

	if !ok {
		return
	}

	split := strings.Split(tags, ",")
	name = split[0]

	for _, v := range split[1:] {
		switch v {
		case "omit":
			options[v] = true
		}
	}

	return
}

// marshal recursively converts any values to a valid tree.
// Everything is copied, it will not contain references to the original values.
func marshal(v reflect.Value) (interface{}, error) {

	if !v.IsValid() || v.Kind() == reflect.Ptr && v.IsNil() {
		return nil, nil
	}

	switch i := v.Interface().(type) {
	case Number, json.Number:
		return NumberCreate(i)
	case encoding.TextMarshaler:
		text, err := i.MarshalText()
		if err != nil {
			return nil, err
		}
		return string(text), nil
	case nil:
		return nil, nil
	}

	t := v.Type()

	switch v.Kind() {
	case reflect.Ptr:
		return marshal(v.Elem())

	case reflect.Interface:
		return marshal(v.Elem())

	case reflect.Struct:
		node := Node{}
		for i := 0; i < t.NumField(); i++ {
			ft, fv := t.Field(i), v.Field(i)
			name, options := getTags(ft)
			var err error
			if ft.PkgPath == "" && !(options["omit"] == true) { // Ignore unexported fields, or fields with "omit" set.
				node[name], err = marshal(fv)
				if err != nil {
					return nil, err
				}
			}
		}
		return node, nil

	case reflect.Map:
		node := Node{}
		for _, e := range v.MapKeys() {
			// Only allow strings as keys, because JSON and some other formats wont allow anything else.
			for e.Kind() == reflect.Interface || e.Kind() == reflect.Ptr {
				e = e.Elem()
			}
			if e.Kind() != reflect.String {
				return nil, &ErrKeyIsNotString{e.String(), e.Kind().String()}
			}
			key := e.String()
			var err error
			node[key], err = marshal(v.MapIndex(e))
			if err != nil {
				return nil, err
			}
		}
		return node, nil

	case reflect.Array, reflect.Slice:
		slice := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			index := v.Index(i)
			var err error
			slice[i], err = marshal(index)
			if err != nil {
				return nil, err
			}
		}
		return slice, nil

	case reflect.Bool:
		return v.Bool(), nil

	case reflect.String:
		return v.String(), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64:
		return NumberCreate(v.Interface())

	}

	return nil, &ErrUnexpectedType{"", fmt.Sprintf("%v", t), ""}
}

// unmarshal recursively converts any tree into a given structure/value.
//
// Everything is copied, it will not contain references to the tree values.
// In case of an error, nothing will be written.
func unmarshal(tree interface{}, v reflect.Value) error {
	if !v.IsValid() {
		return nil
	}

	t := v.Type()

	if v.Kind() != reflect.Ptr && !v.CanSet() {
		return &ErrCannotModify{v.String(), v.Kind().String()}
	}

	// TODO: Reduce duplicate code

	if v.CanAddr() && v.Addr().Elem().CanSet() {
		switch i := v.Addr().Interface().(type) {
		case encoding.TextUnmarshaler:
			text, ok := tree.(string)
			if !ok {
				return &ErrUnexpectedType{"", fmt.Sprintf("%T", tree), "string"}
			}
			err := i.UnmarshalText([]byte(text))
			if err != nil {
				return err
			}
			return nil
		}
	}

	switch i := v.Interface().(type) {
	case encoding.TextUnmarshaler:
		text, ok := tree.(string)
		if !ok {
			return &ErrUnexpectedType{"", fmt.Sprintf("%T", tree), "string"}
		}
		err := i.UnmarshalText([]byte(text))
		if err != nil {
			return err
		}
		return nil
	}

	switch v.Kind() {
	case reflect.Interface:
		if v.NumMethod() == 0 { // TODO: Fix interface handling
			copy := recursiveCopy(tree)
			v.Set(reflect.ValueOf(copy))
			return nil
		}
		return unmarshal(tree, v.Elem())

	case reflect.Ptr:
		if tree == nil && v.CanSet() { // If element in tree is nil, write nil pointer.
			v.Set(reflect.Zero(t))
			return nil
		}
		if v.IsNil() {
			if !v.CanSet() {
				return &ErrCannotModify{v.String(), v.Kind().String()}
			}
			new := reflect.New(t.Elem())
			if err := unmarshal(tree, new.Elem()); err != nil {
				return err
			}
			v.Set(new)
			return nil
		}
		return unmarshal(tree, v.Elem())

	case reflect.Struct:
		if node, ok := tree.(Node); ok {
			rStruct := reflect.New(t).Elem()
			for i := 0; i < t.NumField(); i++ {
				ft, fv := t.Field(i), rStruct.Field(i)
				name, options := getTags(ft)
				if ft.PkgPath == "" && !(options["omit"] == true) { // Ignore unexported fields, or fields with "omit" set.
					if subTree, ok := node[name]; ok {
						err := unmarshal(subTree, fv)
						if err != nil {
							return err
						}
					}
				}
			}
			v.Set(rStruct)
			return nil
		}

	case reflect.Map:
		if node, ok := tree.(Node); ok {
			if t.Key() != reflect.TypeOf(tree).Key() {
				return nil
			}
			rMap := reflect.MakeMap(t)
			for k, tv := range node {
				rv := reflect.New(t.Elem()).Elem()
				err := unmarshal(tv, rv)
				if err != nil {
					return err
				}
				rMap.SetMapIndex(reflect.ValueOf(k), rv)
			}
			v.Set(rMap)
			return nil
		}

	case reflect.Slice:
		if slice, ok := tree.([]interface{}); ok {
			rSlice := reflect.MakeSlice(t, len(slice), cap(slice))
			for i, tv := range slice {
				err := unmarshal(tv, rSlice.Index(i))
				if err != nil {
					return err
				}
			}
			v.Set(rSlice)
			return nil
		}

	case reflect.Array:
		if array, ok := tree.([]interface{}); ok {
			rArray := reflect.New(t).Elem()
			for i, tv := range array {
				if i >= rArray.Len() {
					break
				}
				err := unmarshal(tv, rArray.Index(i))
				if err != nil {
					return err
				}
			}
			v.Set(rArray)
			return nil
		}

	case reflect.Bool:
		if tv, ok := tree.(bool); ok {
			v.SetBool(tv)
			return nil
		}

	case reflect.String:
		if tv, ok := tree.(string); ok {
			v.SetString(tv)
			return nil
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if tv, ok := tree.(Number); ok {
			integer, err := tv.Int64()
			if err != nil {
				return err
			}
			v.SetInt(integer)
			return nil
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if tv, ok := tree.(Number); ok {
			integer, err := tv.Uint64()
			if err != nil {
				return err
			}
			v.SetUint(integer)
			return nil
		}

	case reflect.Float32, reflect.Float64:
		if tv, ok := tree.(Number); ok {
			float, err := tv.Float64()
			if err != nil {
				return err
			}
			v.SetFloat(float)
			return nil
		}
	}

	return &ErrUnexpectedType{"", fmt.Sprintf("%T", tree), t.Kind().String()}
}
