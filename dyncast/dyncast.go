package dyncast

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

func ReflectCast(vI interface{}, origin reflect.Value) (reflect.Value, error) {
	var r reflect.Value

	switch origin.Kind() {
	// NUMERIC
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

		f64, ok := vI.(float64)
		if !ok {
			return r, errors.New("can't type assertion vI (interface{}) to float64")
		}

		i64 := int64(f64)
		u64 := uint64(f64)

		switch origin.Kind() {

		case reflect.Int:
			var temp int
			r = reflect.ValueOf(&temp).Elem()
			if !r.OverflowInt(i64) {
				r.SetInt(i64)
			}

		case reflect.Int8:
			var temp int8
			r = reflect.ValueOf(&temp).Elem()
			if !r.OverflowInt(i64) {
				r.SetInt(i64)
			}

		case reflect.Int16:
			var temp int16
			r = reflect.ValueOf(&temp).Elem()
			if !r.OverflowInt(i64) {
				r.SetInt(i64)
			}

		case reflect.Int32:
			var temp int32
			r = reflect.ValueOf(&temp).Elem()
			if !r.OverflowInt(i64) {
				r.SetInt(i64)
			}

		case reflect.Int64:
			var temp int64
			r = reflect.ValueOf(&temp).Elem()
			if !r.OverflowInt(i64) {
				r.SetInt(i64)
			}

		case reflect.Uint:
			var temp uint
			r = reflect.ValueOf(&temp).Elem()
			if !r.OverflowUint(u64) {
				r.SetUint(u64)
			}

		case reflect.Uint8:
			var temp int
			r = reflect.ValueOf(&temp).Elem()
			if !r.OverflowUint(u64) {
				r.SetUint(u64)
			}

		case reflect.Uint16:
			var temp uint16
			r = reflect.ValueOf(&temp).Elem()
			if !r.OverflowUint(u64) {
				r.SetUint(u64)
			}

		case reflect.Uint32:
			var temp uint32
			r = reflect.ValueOf(&temp).Elem()
			if !r.OverflowUint(u64) {
				r.SetUint(u64)
			}

		case reflect.Uint64:
			var temp uint64
			r = reflect.ValueOf(&temp).Elem()
			if !r.OverflowUint(u64) {
				r.SetUint(u64)
			}

		case reflect.Float32:
			var temp float32
			r = reflect.ValueOf(&temp).Elem()
			if !r.OverflowFloat(f64) {
				r.SetFloat(f64)
			}

		case reflect.Float64:
			var temp float64
			r = reflect.ValueOf(&temp).Elem()
			if !r.OverflowFloat(f64) {
				r.SetFloat(f64)
			}

		}

	case reflect.Bool:
		b, ok := vI.(bool)
		if !ok {
			return r, errors.New("can't type assertion vI (interface{}) to bool")
		}
		var temp bool
		r = reflect.ValueOf(&temp).Elem()
		r.SetBool(b)

	case reflect.String:
		s, ok := vI.(string)
		if !ok {
			return r, errors.New("can't type assertion vI (interface{}) to string")
		}
		var temp string
		r = reflect.ValueOf(&temp).Elem()
		r.SetString(s)

	case reflect.Slice:
		//pSI := p.Value.(origin.Type()) - это решило бы все проблемы!!!!

		// ТАК ТОЖЕ НЕЛЬЗЯ ><
		//switch x.Interface().(type) {
		//case []float32:
		//	pSF32 := p.Value.([]float32)
		//	for i, v := range pSF32 {
		//		x.Index(i).Set(reflect.ValueOf(v))
		//	}
		//	.....
		//}

		// Остается так! =(

		sI := vI.([]interface{})
		lsI := len(sI)
		r = reflect.MakeSlice(origin.Type(), lsI, lsI)

		if lsI == 0 {
			break
		}

		for i, v := range sI {
			cv, err := ReflectCast(v, r.Index(0))
			if err != nil {
				return r, errors.WithMessagef(err, "can't dynamic cast element of slice [%d]", i)
			}
			r.Index(i).Set(cv)
		}

	case reflect.Map: // ОГРАНИЧЕНИЯ для MAP - только: map[string]T
		r = reflect.MakeMap(origin.Type())
		typeMapElement := reflect.TypeOf(r.Interface()).Elem()

		m := vI.(map[string]interface{})

		if len(m) == 0 {
			break
		}

		fmt.Println()
		for k, v := range m {
			cv, err := ReflectCast(v, reflect.New(typeMapElement).Elem())
			if err != nil {
				return r, errors.WithMessagef(err, "can't dynamic cast element of map [%d][%v]", k, v)
			}
			r.SetMapIndex(reflect.ValueOf(k), cv)
		}

	default:
		r = reflect.ValueOf(reflect.ValueOf(vI))
	}
	return r, nil
}
