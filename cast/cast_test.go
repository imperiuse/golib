package cast

import (
	"fmt"
	"reflect"
)

//go test -coverprofile=coverage.out && go tool cover -html=coverage.out

func ExampleCast() {

	// Назначение: typedVar, err := interfaceVar.(T)

	// 1. Представим что в процессе работы, получаем некий слайс вида []interface{}
	t := []interface{}{}
	for _, v := range []int{1, 2, 3, 4} {
		t = append(t, v)
	}

	// 2. Хотим восстановить исходный тип
	b := []int{}

	r, err := DynamicTypeAssetion(t, reflect.ValueOf(b))
	if err != nil {
		fmt.Printf("can't cast! :(")
		fmt.Println(err)
	}

	fmt.Printf("%+v %T\n%T\n%+v %T\n", t, t, b, r, r)
}
