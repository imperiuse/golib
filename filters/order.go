package filters

import (
	"fmt"
	"net/http"
)

// OrderFilterer - order of filter interface
type OrderFilterer interface {
	// Append some new filter to chains. !ATTENTION! MODIFY FILTER ! SetNextFilterPointer() to each input Filter!!!
	AppendFilter(...Filterer)
	// Get N-й filter. 0 - First, ... n - Last
	GetFilterN(int) Filterer
	// Generate func - handle fun - filtered handle func of action
	GenerateFilteredHandleFunc(func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request)
}

// OrderFilters - order of filter struct
type OrderFilters struct {
	Order   []string
	filters []Filterer
}

// AppendFilter - Append some new filter to chains.
// !ATTENTION! MODIFY FILTER ! SetNextFilterPointer() to each input Filter!!!
// @param
// 				filters     ...Filterer    -  some Filters interfaces that will chain of filters
func (filterOrder *OrderFilters) AppendFilter(filters ...Filterer) {
	if cntFilters := len(filters); cntFilters > 0 {
		if filterOrder.filters == nil {
			filterOrder.filters = []Filterer{}
		} else {
			// if already filterOrder have some filter so set next filter pointer
			filterOrder.filters[len(filterOrder.filters)].SetNextFilter(&filters[0])
		}
		for i := range filters {
			filterOrder.filters = append(filterOrder.filters, filters[i]) // Save interface of filter
			filters[i].SetSelfPointer(&filters[i])                        // IMPORTANT! Save self interface pointer
			if i+1 != cntFilters {
				filters[i].SetNextFilter(&filters[i+1]) // Save pointer to interface to next filter
			} else {
				filters[i].SetNextFilter(nil) // flag end of filter
			}
		}
	}
}

// GetFilterN - Get N-й filter. 0 - First, ... n - Last
//@param
//				n   int    -  number of filter in chain
func (filterOrder *OrderFilters) GetFilterN(n int) Filterer {
	if len(filterOrder.filters) > n {
		return filterOrder.filters[n]
	}
	return nil
}

// GenerateFilteredHandleFunc - Generate func - handle fun - filtered handle func of action
// ATTENTION CAN PANIC!
//nolint
//	@param
//				handleFunc 	func(http.ResponseWriter, *http.Request)      -  handle of action
//	@return
// 							func(w http.ResponseWriter, r *http.Request)  -  total handle func
func (filterOrder *OrderFilters) GenerateFilteredHandleFunc(handleFunc func(http.ResponseWriter, *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	if StartFilter := filterOrder.GetFilterN(0); StartFilter != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			StartFilter.Run(w, r, handleFunc)
		}
	}
	panic(fmt.Errorf("`GenerateFilteredHandleFunc()` --> fo.GetFilterN(0) return nil"))
}
