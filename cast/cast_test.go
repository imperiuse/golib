package cast

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

//go test -coverprofile=coverage.out && go tool cover -html=coverage.out

func ExampleDynamicTypeAssertion() {
	// Назначение: typedVar, err := interfaceVar.(T)

	// 1. Представим что в процессе работы, получаем некий слайс вида []interface{}
	t := []interface{}{}
	for _, v := range []string{"1.1", "2", "3.0", "4"} {
		t = append(t, v)
	}

	// 2. Хотим восстановить исходный тип
	temp := []int{}

	r, err := DynamicTypeAssertion(t, reflect.ValueOf(temp))
	if err != nil {
		fmt.Printf("can't cast! :(")
		fmt.Println(err)
	}
	fmt.Printf("Input: %+v %T\nType: %T\nResult: %+v %T\n", t, t, temp, r, r)

	// 2. Хотим привести к другому типу
	temp1 := []float64{}

	r, err = DynamicTypeAssertion(t, reflect.ValueOf(temp1))
	if err != nil {
		fmt.Printf("can't cast! :(")
		fmt.Println(err)
	}
	fmt.Printf("Input: %+v %T\nType: %T\nResult: %+v %T\n", t, t, temp1, r, r)
}

// 1. CastToFloat64()

// 1.1. POSITIVE TEST CASE
func TestCastToFloat64Must(t *testing.T) {
	testCasesMustCast := []interface{}{false, true, math.MaxInt64, int8(8), int16(16), int(32), int32(32), int64(64), float32(32), 0, 42, -42, 1.234, math.MaxFloat32, math.MaxFloat64, -42.42, "123", "123.42", "-42.42", "0"}
	testCasesMustCastAns := []float64{0, 1, math.MaxInt64, 8, 16, 32, 32, 64, 32, 0, 42, -42, 1.234, math.MaxFloat32, math.MaxFloat64, -42.42, 123, 123.42, -42.42, 0}
	var r float64
	var ok bool
	for i, v := range testCasesMustCast {
		if r, ok = CastToFloat64(v); !ok {
			t.Errorf("Failed status test #%d cast: %v to float", i, v)
		}
		if r != testCasesMustCastAns[i] {
			t.Errorf("Failed result test #%d cast: %v to float", i, v)
		}
	}
}

// 1.2. NEGATIVE TEST CASE
func TestCastToFloat64Fail(t *testing.T) {
	testCasesNotMustCast := []interface{}{"", nil, new(interface{}), struct{}{}}
	var ok bool
	for i, v := range testCasesNotMustCast {
		if _, ok = CastToFloat64(v); ok {
			t.Errorf("Failed status negative test #%d cast: %v to float", i, v)
		}
	}
}

// 2. numCast()

// 2.1. POSITIVE TEST CASE

func TestNumCast(t *testing.T) {
	testCase := []interface{}{"1.1", 0, -1}
	testKind := []interface{}{int(32), int8(8), int16(16), int32(32), int64(64),
		uint(32), uint8(8), uint16(16), uint32(32), uint64(64), float32(32), float64(64)}
	testCaseAns := [][]interface{}{{int(1), int8(1), int16(1), int32(1), int64(1),
		uint(1), uint8(1), uint16(1), uint32(1), uint64(1), float32(1.1), float64(1.1)},
		{int(0), int8(0), int16(0), int32(0), int64(0),
			uint(0), uint8(0), uint16(0), uint32(0), uint64(0), float32(0), float64(0)},
		{int(-1), int8(-1), int16(-1), int32(-1), int64(-1),
			uint(18446744073709551615), uint8(0), uint16(0), uint32(0), uint64(18446744073709551615), float32(-1), float64(-1)}}

	var r reflect.Value
	var err error
	for i, v := range testCase {
		for j, kind := range testKind {
			r, err = numCast(v, reflect.ValueOf(kind))
			if err != nil {
				t.Errorf("#1 Failed positive test#%d. Origin.Kind #%d %v. On test example#%d: %v. With err: %v", i, j, kind, i, v, err)
			}
			if r.Interface() != testCaseAns[i][j] {
				t.Errorf("#2 Failed positive test#%d. Origin.Kind #%d %v. On test example#%d: %v. Unexpected result: %v", i, j, kind, i, v, r)
			}
		}
	}
}

// 2.2. NEGATIVE TEST CASE

func TestNumCastFail(t *testing.T) {
	testCase := []interface{}{"", nil, new(interface{}), struct{}{}}
	testKind := []interface{}{int(32), int8(8), int16(16), int32(32), int64(64),
		uint(32), uint8(8), uint16(16), uint32(32), uint64(64), float32(32), float64(64)}

	var err error
	for i, v := range testCase {
		for j, kind := range testKind {
			_, err = numCast(v, reflect.ValueOf(kind))
			if err == nil {
				t.Errorf("#1 Failed negative test#%d. Origin.Kind #%d %v. On test example: %v. NO ERROR!!", i, j, kind, v)
			}
		}
	}
}

// 3. DynamicTypeAssertion()

// 3.1. POSITIVE TEST CASE

// 3.2. NEGATIVE TEST CASE
