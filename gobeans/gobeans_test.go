package gobeans

import (
	"fmt"
	"reflect"
	"testing"
)

// TestStruct1 - первая тестовая структура
type TestStruct1 struct {
	FieldBool    bool
	FieldStr     string
	FieldInt     int
	FieldFloat64 float64
}

// TestStruct2 - вторая тестовая структура
type TestStruct2 struct {
	TestStruct1
	FieldStrOut string
}

// TestStruct3 - третья тестовая структура
type TestStruct3 struct {
	FieldSliceStr     []string
	FieldSliceInt     []int
	FieldSliceFloat64 []float64

	FieldMapStringString  map[string]string
	FieldMapStringInt     map[string]int
	FieldMapStringFloat64 map[string]float64

	T1  TestStruct1
	T2  TestStruct2
	PT1 *TestStruct1
	PT2 *TestStruct2
}

func ExampleCreateBeanStorage() {

	Beans := CreateBeanStorage()
	err := Beans.RegType((*float64)(nil), (*int)(nil), (*uint)(nil), (*string)(nil), (*TestStruct1)(nil),
		(*TestStruct2)(nil), (*TestStruct3)(nil))

	if err != nil {
		print("\nError while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}

	if err = Beans.CreateBeansFromJSON("./test_json/beansTest.json"); err != nil {
		print("Error while create Beans from json file", err)
	} else {
		for id, bean := range Beans.GetMapBeans() {
			fmt.Printf("Bean name: [%+v] \n\tJFI: %+v \n\tPIO: %v\n", id, bean.JFI,
				bean.PIO)
		}
	}
}

func TestCreateBeanStorage(t *testing.T) {
	Beans := CreateBeanStorage()
	if Beans.typeMap == nil || Beans.beansMap == nil {
		t.Errorf("Unxpected value of field Beans!")
	}
}

func TestRegType(t *testing.T) {
	Beans := CreateBeanStorage()
	err := Beans.RegType((*float64)(nil), (*int)(nil), (*uint)(nil), (*string)(nil), (*TestStruct1)(nil),
		(*TestStruct2)(nil), (*TestStruct3)(nil))
	if err != nil {
		t.Errorf("error while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}
}

func TestGetAllNamesRegistryTypes(t *testing.T) {
	Beans := CreateBeanStorage()
	err := Beans.RegType((*float64)(nil), (*int)(nil), (*uint)(nil), (*string)(nil), (*TestStruct1)(nil),
		(*TestStruct2)(nil), (*TestStruct3)(nil))
	if err != nil {
		t.Errorf("error while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}
	Beans.ShowRegTypes()
}

func TestGetAllNamesRegistryTypesNegative(t *testing.T) {
	Beans := CreateBeanStorage()
	err := Beans.RegType((************TestStruct1)(nil))
	if err == nil {
		t.Errorf("\nNO Error while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}
	err = Beans.RegType((************TestStruct2)(nil))
	if err == nil {
		t.Errorf("\nNO Error while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}
	Beans.ShowRegTypes()
}

func TestCreateBeansFromJSON(t *testing.T) {
	Beans := CreateBeanStorage()
	err := Beans.RegType((*float64)(nil), (*int)(nil), (*uint)(nil), (*string)(nil), (*TestStruct1)(nil),
		(*TestStruct2)(nil), (*TestStruct3)(nil), (**TestStruct1)(nil), (**TestStruct2)(nil))
	if err != nil {
		t.Errorf("\nError while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}
	if err = Beans.CreateBeansFromJSON("./test_json/beansTest.json"); err != nil {
		t.Errorf("Error while create Beans from json file. %v", err)
	}
}

func TestCreateBeansFromJSONNegative(t *testing.T) {
	Beans := CreateBeanStorage()
	err := Beans.RegType((*float64)(nil), (*int)(nil), (*uint)(nil), (*string)(nil), (*TestStruct1)(nil),
		(*TestStruct2)(nil), (*TestStruct3)(nil), (**TestStruct1)(nil), (**TestStruct2)(nil))
	if err != nil {
		t.Errorf("\nError while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}
	if err = Beans.CreateBeansFromJSON("./test_json/badBeansTest.json"); err == nil {
		t.Errorf("No Error while create Beans from BAD BEANS file badBeansTest.json")
	}
	if err = Beans.CreateBeansFromJSON("./test_json/badBeansTest2.json"); err == nil {
		t.Errorf("No Error while create Beans from BAD BEANS2 file badBeansTest2.json")
	}
	if err = Beans.CreateBeansFromJSON("./test_json/badBeansTest3.json"); err == nil {
		t.Errorf("No Error while create Beans from BAD BEANS3 file badBeansTest3.json")
	}
	if err = Beans.CreateBeansFromJSON("./test_json/badBeansTest4.json"); err == nil {
		t.Errorf("No Error while create Beans from BAD BEANS4 file badBeansTest4.json")
	}
	if err = Beans.CreateBeansFromJSON("./test_json/badJsonBeansTest.json"); err == nil {
		t.Errorf("No Error while create Beans from BAD JSON file badJsonBeansTest.json")
	}
	if err = Beans.CreateBeansFromJSON("./test_json/unknown.json"); err == nil {
		t.Errorf("No Error while create Beans from UNKNOWN json file unknown.json")
	}
}

//nolint
func TestGetBeansAndGetReflectType(t *testing.T) {
	Beans := CreateBeanStorage()
	err := Beans.RegType((*float64)(nil), (*int)(nil), (*uint)(nil), (*string)(nil), (*TestStruct1)(nil),
		(*TestStruct2)(nil), (*TestStruct3)(nil), (**TestStruct1)(nil), (**TestStruct2)(nil))
	if err != nil {
		t.Errorf("\nError while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}
	if err = Beans.CreateBeansFromJSON("./test_json/beansTest.json"); err != nil {
		t.Errorf("Error while create Beans from json file. %v", err)
	}

	Beans.GetBean("IDTestStruct1")
	Beans.GetMapBeans()
	Beans.GetIDBeans()
	if typ := Beans.GetReflectTypeByName("github.com/imperiuse/golib/gobeans.TestStruct2"); typ != reflect.TypeOf((*TestStruct2)(nil)).Elem() {
		t.Errorf("Error! Unexpected value of reflect type: TestStruct2")
	}
	if typ, found := Beans.FoundAndGetReflectTypeByName("github.com/imperiuse/golib/gobeans.TestStruct3"); !found {
		t.Errorf("Error while get type of exist registrated struct: TestStruct3 %v", err)
	} else if typ != reflect.TypeOf((*TestStruct3)(nil)).Elem() {
		t.Errorf("Error! Unexpected value of reflect type: TestStruct3")
	}
	if _, found := Beans.FoundAndGetReflectTypeByName("UnknownType12345"); found {
		t.Errorf("No Error! For un registrated type: UnknownTupe12345")
	}
}

func TestClonesFunc(t *testing.T) {
	Beans := CreateBeanStorage()
	err := Beans.RegType((*float64)(nil), (*int)(nil), (*uint)(nil), (*string)(nil), (*TestStruct1)(nil),
		(*TestStruct2)(nil), (*TestStruct3)(nil), (**TestStruct1)(nil), (**TestStruct2)(nil))
	if err != nil {
		t.Errorf("\nError while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}
	if err = Beans.CreateBeansFromJSON("./test_json/beansTest.json"); err != nil {
		t.Errorf("Error while create Beans from json file. %v", err)
	}

	pc := Beans.GetCloneBeanPIO("IDTestStruct1").(*TestStruct1)
	jc := Beans.GetCloneBeanJFI("IDTestStruct1").(*TestStruct1)
	pc.FieldInt = 0

	p := Beans.GetBeanPIO("IDTestStruct1").(*TestStruct1)
	j := Beans.GetBeanJFI("IDTestStruct1").(TestStruct1)

	p.FieldInt = 5

	if pc == p || p.FieldInt == pc.FieldInt {
		t.Errorf("\nBeans not clones!")
	}

	if j != *jc {
		t.Errorf("JFI change!")
	}

}
