package cast

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
)

// DynamicTypeAssertion - функция для динамического сопостовления типа делает примерно следующее
//                       r := vI.(origin.Type()) - это решило бы все проблемы!!!!
func DynamicTypeAssertion(vI interface{}, origin reflect.Value) (reflect.Value, error) {
	var r reflect.Value
	var err error

	if vI == nil {
		return r, errors.New("received nil vI param!")
	}

	switch origin.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

		r, err = numCast(vI, origin)
		if err != nil {
			return r, err
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
		sI := vI.([]interface{})
		lsI := len(sI)
		r = reflect.MakeSlice(origin.Type(), lsI, lsI)

		if lsI == 0 {
			break
		}

		for i, v := range sI {
			cv, err := DynamicTypeAssertion(v, r.Index(0))
			if err != nil {
				return r, errors.WithMessagef(err, "can't dynamic cast element of slice [%d]", i)
			}
			r.Index(i).Set(cv)

		}

	case reflect.Map:
		r = reflect.MakeMap(origin.Type())
		typeMapElement := reflect.TypeOf(r.Interface()).Elem()

		m := vI.(map[string]interface{})

		if len(m) == 0 {
			break
		}

		fmt.Println()
		for k, v := range m {
			cv, err := DynamicTypeAssertion(v, reflect.New(typeMapElement).Elem())
			if err != nil {
				return r, errors.WithMessagef(err, "can't dynamic cast element of map [%v][%v]", k, v)
			}
			r.SetMapIndex(reflect.ValueOf(k), cv)
		}

	default:
		r = reflect.ValueOf(reflect.ValueOf(vI))
	}
	return r, nil
}

// numCast - вспомогательная функция для приведения числовых значений
func numCast(vI interface{}, origin reflect.Value) (reflect.Value, error) {
	var r reflect.Value

	f64, ok := CastToFloat64(vI)
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
	return r, nil
}

// CastToFloat64 - по возможности приводит интерфейс к типу float64
func CastToFloat64(v interface{}) (float64, bool) {
	switch i := v.(type) {
	case nil:
		return float64(0), false
	case int64:
		return float64(int64(i)), true
	case int:
		return float64(int64(i)), true
	case int16:
		return float64(int64(i)), true
	case int8:
		return float64(int64(i)), true
	case float64:
		return float64(i), true
	case float32:
		return float64(i), true
	case bool:
		if bool(i) {
			return float64(1), true
		}
		return float64(0), true
	case string:
		f, err := strconv.ParseFloat(string(i), 64)
		if err != nil {
			return float64(0), false
		}
		return f, true
	default:
		return 0, false
	}
}
