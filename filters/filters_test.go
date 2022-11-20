package filters

import (
	"fmt"
	"net/http"
	"testing"
)

// Define Own Custom Filter
type MyFirstCustomFilter struct {
	BaseFilter
	cntIn     int // counter for Before method
	cntOut    int // counter for After method
	cntFilter int // counter for Filter method
}

// Implement some of method Filterer interface

// Let's three simple filter methods and one Info method

// Before - first filter method before method Filter(...)
func (f *MyFirstCustomFilter) Before(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[MyCustomFilter] Before()")
	f.cntIn++
	fmt.Println((*f.selfPointer).Info())
}

// After - last filter method after Filter(...)
func (f *MyFirstCustomFilter) After(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[MyCustomFilter] After()")
	f.cntOut++
	fmt.Println((*f.selfPointer).Info())
}

// Filter - main filter method
func (f *MyFirstCustomFilter) Filter(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[MyCustomFilter] Filter()")
	f.cntFilter++
	fmt.Println((*f.selfPointer).Info())
}

// Info - information about method
func (f *MyFirstCustomFilter) Info() string {
	return fmt.Sprintf("[MyCustomFilter]: %+v", *f)
}

func ExampleBaseFilter_Run() {

	// Define some struct and methods @see upper

	// Build filter struct (better use gobeans way create obj @see here --> github.com/imperiuse/golib/gobeans )
	CustomFilter := MyFirstCustomFilter{}
	CustomFilter.Name = "MyCustomFilter"

	CustomFilter2 := MyFirstCustomFilter{}
	CustomFilter2.Name = "MyCustomFilter2"

	// Get interface by pointer filter struct
	InterfaceCustomFilter := Filterer(&CustomFilter)   // IMPORTANT! Better work with interface type of Filterer
	InterfaceCustomFilter2 := Filterer(&CustomFilter2) // IMPORTANT! Better work with interface type of Filterer

	InterfaceCustomFilter.SetSelfPointer(&InterfaceCustomFilter) // ATTENTION! ALWAYS SET THIS! Self pinter! IMPORTANT!
	InterfaceCustomFilter.SetNextFilter(&InterfaceCustomFilter2) // next filter pointer

	InterfaceCustomFilter2.SetSelfPointer(&InterfaceCustomFilter2) // ATTENTION! ALWAYS SET THIS! Self pinter! IMPORTANT!
	InterfaceCustomFilter2.SetNextFilter(nil)                      // next filter pointer

	// Let's test

	// Our child Info method
	InterfaceCustomFilter.Info() //print: [MyCustomFilter]: ...

	// Parent method
	InterfaceCustomFilter.GetBaseFilter().Info() //print: [BaseFilter] Before():

	// Start filter
	var response http.ResponseWriter
	var request *http.Request
	bf := func(http.ResponseWriter, *http.Request) {} // empty func

	InterfaceCustomFilter.Run(response, request, bf) // Run all Filter in rigth order
	// MyCustomFilter.Before()->MyCustomFilter.Filter()->MyCustomFilter.GetNextFilter() == return MyCustomFilter2
	// MyCustomFilter2.Run()->MyCustomFilter2.Before()->MyCustomFilter2.Filter()->->MyCustomFilter.GetNextFilter() == nil
	// MyCustomFilter2.After()->MyCustomFilter.After()-> end.
}

func TestBaseFilter_Run_Positive(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic was!")
		}
	}()

	CustomFilter := MyFirstCustomFilter{}
	CustomFilter.Name = "MyCustomFilter"

	CustomFilter2 := MyFirstCustomFilter{}
	CustomFilter2.Name = "MyCustomFilter2"

	// Get interface by pointer filter struct
	InterfaceCustomFilter := Filterer(&CustomFilter)   // IMPORTANT! Better work with interface type of Filterer
	InterfaceCustomFilter2 := Filterer(&CustomFilter2) // IMPORTANT! Better work with interface type of Filterer

	InterfaceCustomFilter.SetSelfPointer(&InterfaceCustomFilter) // ATTENTION! ALWAYS SET THIS! Self pinter! IMPORTANT!
	InterfaceCustomFilter.SetNextFilter(&InterfaceCustomFilter2) // next filter pointer

	InterfaceCustomFilter2.SetSelfPointer(&InterfaceCustomFilter2) // ATTENTION! ALWAYS SET THIS! Self pinter! IMPORTANT!
	InterfaceCustomFilter2.SetNextFilter(nil)                      // next filter pointer

	// Let's test

	// Our child Info method
	InterfaceCustomFilter.Info() //print: [MyCustomFilter]: ...

	// Parent method
	InterfaceCustomFilter.GetBaseFilter().Info() //print: [BaseFilter] Before():

	// Start filter
	var response http.ResponseWriter
	var request *http.Request
	bf := func(http.ResponseWriter, *http.Request) {} // empty func

	InterfaceCustomFilter.Run(response, request, bf)

}

func TestBaseFilter_Run_Negative(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic was!")
		}
	}()

	// Define Own Custom Filter
	type MyCustomFilter struct {
		BaseFilter
	}

	// Build filter struct (better use gobeans way create obj)
	CustomFilter1 := MyCustomFilter{BaseFilter{"CustomFilter1", nil, nil}}
	CustomFilter2 := MyCustomFilter{BaseFilter{"CustomFilter2", nil, nil}}

	// Get interface by pointer filter struct
	InterfaceCustomFilter1 := Filterer(&CustomFilter1)
	InterfaceCustomFilter2 := Filterer(&CustomFilter2)

	InterfaceCustomFilter1.SetSelfPointer(&InterfaceCustomFilter1) // ATTENTION! ALWAYS SET THIS! Self pinter! IMPORTANT!
	InterfaceCustomFilter1.SetNextFilter(&InterfaceCustomFilter2)  // next filter pointer

	InterfaceCustomFilter2.SetSelfPointer(&InterfaceCustomFilter2) // ATTENTION! ALWAYS SET THIS! Self pinter! IMPORTANT!
	InterfaceCustomFilter2.SetNextFilter(nil)                      // next filter pointer

	// Start filter
	var response http.ResponseWriter
	var request *http.Request
	InterfaceCustomFilter1.Run(response, request, nil)

}
