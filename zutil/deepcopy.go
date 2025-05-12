package zutil

import (
	"fmt"
	"reflect"
)

type copier func(interface{}, map[uintptr]interface{}) (interface{}, error)

var copiers map[reflect.Kind]copier

func init() {
	copiers = map[reflect.Kind]copier{
		reflect.Bool:       _primitive,
		reflect.Int:        _primitive,
		reflect.Int8:       _primitive,
		reflect.Int16:      _primitive,
		reflect.Int32:      _primitive,
		reflect.Int64:      _primitive,
		reflect.Uint:       _primitive,
		reflect.Uint8:      _primitive,
		reflect.Uint16:     _primitive,
		reflect.Uint32:     _primitive,
		reflect.Uint64:     _primitive,
		reflect.Uintptr:    _primitive,
		reflect.Float32:    _primitive,
		reflect.Float64:    _primitive,
		reflect.Complex64:  _primitive,
		reflect.Complex128: _primitive,
		reflect.Array:      _array,
		reflect.Map:        _map,
		reflect.Ptr:        _pointer,
		reflect.Slice:      _slice,
		reflect.String:     _primitive,
		reflect.Struct:     _struct,
	}
}

// DeepCopy
// @Description: 深拷贝
// @param x
// @return interface{}
// @return error
func DeepCopy(x interface{}) (interface{}, error) {
	copd := make(map[uintptr]interface{})
	return _deepCopy(x, copd)
}

func _deepCopy(x interface{}, copd map[uintptr]interface{}) (interface{}, error) {
	v := reflect.ValueOf(x)
	if !v.IsValid() {
		return x, nil
	}
	if c, ok := copiers[v.Kind()]; ok {
		return c(x, copd)
	}
	t := reflect.TypeOf(x)
	return nil, fmt.Errorf("unable to make a deep copy of %v (type: %v) - kind %v is not supported", x, t, v.Kind())
}

func _primitive(x interface{}, copd map[uintptr]interface{}) (interface{}, error) {
	kind := reflect.ValueOf(x).Kind()
	if kind == reflect.Array || kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface || kind == reflect.Map || kind == reflect.Ptr || kind == reflect.Slice || kind == reflect.Struct || kind == reflect.UnsafePointer {
		return nil, fmt.Errorf("unable to copy %v (a %v) as a primitive", x, kind)
	}
	return x, nil
}

func _slice(x interface{}, copd map[uintptr]interface{}) (interface{}, error) {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("must pass a value with kind of Slice; got %v", v.Kind())
	}
	size := v.Len()
	t := reflect.TypeOf(x)
	dc := reflect.MakeSlice(t, size, size)
	for i := 0; i < size; i++ {
		item, err := _deepCopy(v.Index(i).Interface(), copd)
		if err != nil {
			return nil, fmt.Errorf("failed to clone slice item at index %v: %v", i, err)
		}
		iv := reflect.ValueOf(item)
		if iv.IsValid() {
			dc.Index(i).Set(iv)
		}
	}
	return dc.Interface(), nil
}

func _map(x interface{}, copd map[uintptr]interface{}) (interface{}, error) {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Map {
		return nil, fmt.Errorf("must pass a value with kind of Map; got %v", v.Kind())
	}
	t := reflect.TypeOf(x)
	dc := reflect.MakeMapWithSize(t, v.Len())
	iter := v.MapRange()
	for iter.Next() {
		item, err := _deepCopy(iter.Value().Interface(), copd)
		if err != nil {
			return nil, fmt.Errorf("failed to clone map item %v: %v", iter.Key().Interface(), err)
		}
		k, err := _deepCopy(iter.Key().Interface(), copd)
		if err != nil {
			return nil, fmt.Errorf("failed to clone the map key %v: %v", k, err)
		}
		dc.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(item))
	}
	return dc.Interface(), nil
}

func _pointer(x interface{}, copd map[uintptr]interface{}) (interface{}, error) {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("must pass a value with kind of Ptr; got %v", v.Kind())
	}

	if v.IsNil() {
		t := reflect.TypeOf(x)
		return reflect.Zero(t).Interface(), nil
	}

	addr := v.Pointer()
	if dc, ok := copd[addr]; ok {
		return dc, nil
	}
	t := reflect.TypeOf(x)
	dc := reflect.New(t.Elem())
	copd[addr] = dc.Interface()

	item, err := _deepCopy(v.Elem().Interface(), copd)
	if err != nil {
		return nil, fmt.Errorf("failed to copy the value under the pointer %v: %v", v, err)
	}
	iv := reflect.ValueOf(item)
	if iv.IsValid() {
		dc.Elem().Set(reflect.ValueOf(item))
	}

	return dc.Interface(), nil
}

func _struct(x interface{}, copd map[uintptr]interface{}) (interface{}, error) {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("must pass a value with kind of Struct; got %v", v.Kind())
	}
	t := reflect.TypeOf(x)
	dc := reflect.New(t)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		item, err := _deepCopy(v.Field(i).Interface(), copd)
		if err != nil {
			return nil, fmt.Errorf("failed to copy the field %v in the struct %#v: %v", t.Field(i).Name, x, err)
		}
		dc.Elem().Field(i).Set(reflect.ValueOf(item))
	}
	return dc.Elem().Interface(), nil
}

func _array(x interface{}, copd map[uintptr]interface{}) (interface{}, error) {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Array {
		return nil, fmt.Errorf("must pass a value with kind of Array; got %v", v.Kind())
	}
	t := reflect.TypeOf(x)
	size := t.Len()
	dc := reflect.New(reflect.ArrayOf(size, t.Elem())).Elem()
	for i := 0; i < size; i++ {
		item, err := _deepCopy(v.Index(i).Interface(), copd)
		if err != nil {
			return nil, fmt.Errorf("failed to clone array item at index %v: %v", i, err)
		}
		dc.Index(i).Set(reflect.ValueOf(item))
	}
	return dc.Interface(), nil
}
