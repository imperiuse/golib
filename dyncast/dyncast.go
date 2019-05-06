package dyncast

import (
	"fmt"
	"reflect"
)

func ReflectCast(vI interface{}, kind reflect.Kind) (reflect.Value, error) {
	var r reflect.Value

	switch kind {
	// NUMERIC
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

		f64, ok := vI.(float64)
		if !ok {
			return r, fmt.Errorf("can't type assertion vI (interface{}) to float64")
		}

		i64 := int64(f64)
		u64 := uint64(f64)

		switch kind {

		case reflect.Int:
			var temp int
			r = reflect.ValueOf(temp)
			if !r.OverflowInt(i64) {
				r.SetInt(i64)
			}

		case reflect.Int8:
			var temp int8
			r = reflect.ValueOf(temp)
			if !r.OverflowInt(i64) {
				r.SetInt(i64)
			}

		case reflect.Int16:
			var temp int16
			r = reflect.ValueOf(temp)
			if !r.OverflowInt(i64) {
				r.SetInt(i64)
			}

		case reflect.Int32:
			var temp int32
			r = reflect.ValueOf(temp)
			if !r.OverflowInt(i64) {
				r.SetInt(i64)
			}

		case reflect.Int64:
			var temp int64
			r = reflect.ValueOf(temp)
			if !r.OverflowInt(i64) {
				r.SetInt(i64)
			}

		case reflect.Uint:
			var temp uint
			r = reflect.ValueOf(temp)
			if !r.OverflowUint(u64) {
				r.SetUint(u64)
			}

		case reflect.Uint8:
			var temp int
			r = reflect.ValueOf(temp)
			if !r.OverflowUint(u64) {
				r.SetUint(u64)
			}

		case reflect.Uint16:
			var temp uint16
			r = reflect.ValueOf(temp)
			if !r.OverflowUint(u64) {
				r.SetUint(u64)
			}

		case reflect.Uint32:
			var temp uint32
			r = reflect.ValueOf(temp)
			if !r.OverflowUint(u64) {
				r.SetUint(u64)
			}

		case reflect.Uint64:
			var temp uint64
			r = reflect.ValueOf(temp)
			if !r.OverflowUint(u64) {
				r.SetUint(u64)
			}

		case reflect.Float32:
			var temp float32
			r = reflect.ValueOf(temp)
			if !r.OverflowFloat(f64) {
				r.SetFloat(f64)
			}

		}
	case reflect.Bool:
		b, ok := vI.(bool)
		if !ok {
			return r, fmt.Errorf("can't type assertion vI (interface{}) to bool")
		}
		var temp bool
		r = reflect.ValueOf(temp)
		r.SetBool(b)

	case reflect.String:
		s, ok := vI.(string)
		if !ok {
			return r, fmt.Errorf("can't type assertion vI (interface{}) to string")
		}
		var temp string
		r = reflect.ValueOf(temp)
		r.SetString(s)

	default:
		r = reflect.ValueOf(reflect.ValueOf(vI)) // перем значение, а не указатель
	}
	return r, nil
}
