package filters

import (
	"net/http"
	"testing"
)

func ExampleBaseFilter_Run() {

	// Define Own Custom Filter
	type MyCustomFilter struct {
		BaseFilter
	}

	// Build filter struct (better use gobeans way create obj @see here --> github.com/imperiuse/golib/gobeans )
	CustomFilter := MyCustomFilter{BaseFilter{"CustomFilter", nil, nil}}

	// Get interface by pointer filter struct
	InterfaceCustomFilter := Filterer(&CustomFilter)

	InterfaceCustomFilter.SetSelfPointer(&InterfaceCustomFilter) // ATTENTION! ALWAYS SET THIS! Self pinter! IMPORTANT!
	InterfaceCustomFilter.SetNextFilter(nil)                     // next filter pointer

	// Start filter
	var response http.ResponseWriter
	var request *http.Request
	bf := func(http.ResponseWriter, *http.Request) { return }
	InterfaceCustomFilter.Run(response, request, bf)
}

func TestBaseFilter_Run_Positive(t *testing.T) {
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
	bf := func(http.ResponseWriter, *http.Request) { return }
	InterfaceCustomFilter1.Run(response, request, bf)

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
