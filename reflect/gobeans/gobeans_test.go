package gobeans

//go test -covermode=count -coverprofile=coverage.cov && go tool cover -html=coverage.cov

import (
	"fmt"
	"reflect"
	"testing"
)

// TODO REFACTOR -> USE TESTIFY

// TestNatural - первая тестовая структура
//nolint
type TestNatural struct {
	FBool    bool
	FStr     string
	FInt     int
	FInt8    int8
	FInt16   int16
	FInt32   int32
	FInt64   int64
	FUint    uint
	FUint8   uint8
	FUint16  uint16
	FUint32  uint32
	FUint64  uint64
	FFloat32 float32
	FFloat64 float64

	SBool    []bool
	SString  []string
	SInt     []int
	SInt8    []int8
	SInt16   []int16
	SInt32   []int32
	SInt64   []int64
	SUint    []uint
	SUint8   []uint8
	SUint16  []uint16
	SUint32  []uint32
	SUint64  []uint64
	SFloat32 []float32
	SFloat64 []float64

	MstringBool    map[string]bool
	MstringString  map[string]string
	MstringInt     map[string]int
	MstringInt8    map[string]int8
	MstringInt16   map[string]int16
	MstringInt32   map[string]int32
	MstringInt64   map[string]int64
	MstringUint    map[string]uint
	MstringUint8   map[string]uint8
	MstringUint16  map[string]uint16
	MstringUint32  map[string]uint32
	MstringUint64  map[string]uint64
	MstringFloat32 map[string]float32
	MstringFloat64 map[string]float64
}

// TestStruct2 - вторая тестовая структура
type TestAgrStruct struct {
	TestNatural
	FS        TestNatural
	PFS       *TestNatural
	InnerBean TestNatural
	Self      *TestAgrStruct
}

func ExampleCreateStorage() {

	Beans, err := CreateStorage()
	if err != nil {
		print("\nError while gobeans.CreateStorage: %v\n", err)
	} else {
		fmt.Print("\nStorage successful created!\n")
	}

	err = Beans.RegNamedTypes((*TestNatural)(nil), "TestNatural", (**TestNatural)(nil), "*TestNatural")

	if err != nil {
		print("\nError while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}

	if err = Beans.CreateBeansFromJSON("./test_json/beansTest.json"); err != nil {
		print("Error while create Beans from json file", err)
	} else {
		for id, bean := range Beans.GetMapBeans() {
			fmt.Printf("Bean name: [%+v] \n\n\tPIO: %v\n", id, bean.pio)
		}
	}
}

func TestCreateBeanStorage(t *testing.T) {
	Beans, err := CreateStorage()
	if err != nil {
		t.Errorf("\nError while gobeans.CreateStorage: %v\n", err)
	} else {
		fmt.Print("\nStorage successful created!\n")
	}
	if Beans.typeMap == nil || Beans.beansMap == nil {
		t.Errorf("Unxpected value of field Beans!")
	}
}

func TestRegType(t *testing.T) {
	Beans, err := CreateStorage()
	if err != nil {
		t.Errorf("\nError while gobeans.CreateStorage: %v\n", err)
	} else {
		fmt.Print("\nStorage successful created!\n")
	}
	err = Beans.RegNamedTypes((*TestNatural)(nil), "TestNatural", (*TestAgrStruct)(nil), "TestAgrStruct")
	if err != nil {
		t.Errorf("error while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}

	err = Beans.RegTypes((*TestNatural)(nil), (*TestAgrStruct)(nil), (*TestAgrStruct)(nil))
	if err != nil {
		t.Errorf("error while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}
}

func TestGetAllNamesRegistryTypes(t *testing.T) {
	Beans, err := CreateStorage()
	if err != nil {
		t.Errorf("\nError while gobeans.CreateStorage: %v\n", err)
	} else {
		fmt.Print("\nStorage successful created!\n")
	}
	err = Beans.RegNamedTypes((*TestNatural)(nil), "TestNatural", (*TestAgrStruct)(nil), "TestAgrStruct")
	if err != nil {
		t.Errorf("error while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}
	Beans.ShowRegTypes()
}

func TestGetAllNamesRegistryTypesNegative(t *testing.T) {
	Beans, err := CreateStorage()
	if err != nil {
		t.Errorf("\nError while gobeans.CreateStorage: %v\n", err)
	} else {
		fmt.Print("\nStorage successful created!\n")
	}
	err = Beans.RegTypes((************TestNatural)(nil))
	if err == nil {
		t.Errorf("\nNO Error while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}
	err = Beans.RegTypes((************TestAgrStruct)(nil))
	if err == nil {
		t.Errorf("\nNO Error while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}
	Beans.ShowRegTypes()
}

func TestCreateBeansFromJSON(t *testing.T) {
	Beans, err := CreateStorage()
	if err != nil {
		t.Errorf("\nError while gobeans.CreateStorage: %v\n", err)
	} else {
		fmt.Print("\nStorage successful created!\n")
	}
	err = Beans.RegNamedTypes((*TestNatural)(nil), "TestNatural", (**TestNatural)(nil), "*TestNatural",
		(*TestAgrStruct)(nil), "TestAgrStruct")
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
	Beans, err := CreateStorage()
	if err != nil {
		t.Errorf("\nError while gobeans.CreateStorage: %v\n", err)
	} else {
		fmt.Print("\nStorage created!\n")
	}
	err = Beans.RegNamedTypes((*TestNatural)(nil), "TestNatural", (**TestNatural)(nil), "*TestNatural",
		(*TestAgrStruct)(nil), "TestAgrStruct")
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
	Beans, err := CreateStorage()
	if err != nil {
		t.Errorf("\nError while gobeans.CreateStorage: %v\n", err)
	} else {
		fmt.Print("\nStorage created!\n")
	}
	err = Beans.RegNamedTypes((*TestNatural)(nil), "TestNatural", (**TestNatural)(nil), "*TestNatural",
		(*TestAgrStruct)(nil), "TestAgrStruct")
	if err != nil {
		t.Errorf("\nError while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}
	if err = Beans.CreateBeansFromJSON("./test_json/beansTest.json"); err != nil {
		t.Errorf("Error while create Beans from json file. %v", err)
	}

	Beans.GetBean("natural")
	Beans.GetMapBeans()
	Beans.GetAllBeansID()
	if typ := Beans.GetReflectTypeByName("TestNatural"); typ != reflect.TypeOf((*TestNatural)(nil)).Elem() {
		t.Errorf("Error! Unexpected value of reflect type: TestStruct2")
	}
	if typ, found := Beans.FoundReflectTypeByName("TestAgrStruct"); !found {
		t.Errorf("Error while get type of exist registrated struct: TestStruct3 %v", err)
	} else if typ != reflect.TypeOf((*TestAgrStruct)(nil)).Elem() {
		t.Errorf("Error! Unexpected value of reflect type: TestStruct3")
	}
	if _, found := Beans.FoundReflectTypeByName("UnknownType12345"); found {
		t.Errorf("No Error! For un registrated type: UnknownTupe12345")
	}
}

func TestClonesFunc(t *testing.T) {
	Beans, err := CreateStorage()
	if err != nil {
		t.Errorf("\nError while gobeans.CreateStorage: %v\n", err)
	} else {
		fmt.Print("\nStorage created!\n")
	}
	err = Beans.RegNamedTypes((*TestNatural)(nil), "TestNatural", (**TestNatural)(nil), "*TestNatural",
		(*TestAgrStruct)(nil), "TestAgrStruct")
	if err != nil {
		t.Errorf("\nError while gobeans.RegType: %v\n", err)
	} else {
		fmt.Printf("\nRegistrate types: %v\n", Beans.ShowRegTypes())
	}
	if err = Beans.CreateBeansFromJSON("./test_json/beansTest.json"); err != nil {
		t.Errorf("Error while create Beans from json file. %v", err)
	}

	pcI, _ := Beans.CloneBean("natural")

	pc := pcI.(*TestNatural)

	pc.FInt = -123

	p := Beans.GetBean("natural").(*TestNatural)

	p.FInt = 5

	if pc == p || p.FInt == pc.FInt {
		t.Errorf("\nBeans not clones!")
	}
}
