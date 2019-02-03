package filters

import (
	"net/http"
	"testing"
)

// ExampleOrderFilters_AppendFilter - an example of use OrderFilterer
func ExampleOrderFilters_AppendFilter() {
	// At the begin, we have coupled of filters in map of Filter
	// better use MapBeansType --> @see github.com/imperiuse/gobeans   Type: MapBeansType
	mapOfFilterer := map[string]Filterer{"f1": &BaseFilter{}, "f2": &BaseFilter{}, "f3": &BaseFilter{}}
	// Create new Filter Order
	fOrder := OrderFilterer(&OrderFilters{[]string{"f1", "f2", "f3"}, nil})
	// Configure Filterer to chain of Filters
	for _, nameFilter := range fOrder.GetOrderFilters() {
		fOrder.AppendFilter(mapOfFilterer[nameFilter])
	}
	// Next, we have some business func, which we want "decorate" by own filters with special order define upper
	businesFunc := func(http.ResponseWriter, *http.Request) { return }
	// Do this!
	summaryFilteredBusinesFuncf := fOrder.GenerateFilteredHandleFunc(businesFunc)
	_ = summaryFilteredBusinesFuncf
}

// TestOrderFilters_GetOrderFilters - test right return Order
func TestOrderFilters_GetOrderFilters(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unknown Panic in TestOrderFilters_GetOrderFilters")
		}
	}()
	order := []string{"f1", "f2", "f3"}
	fOrder := OrderFilterer(&OrderFilters{order, nil})
	for i, nameF := range fOrder.GetOrderFilters() {
		if nameF != order[i] {
			t.Errorf("Missmath name of filters in Order slice")
		}
	}
}

func TestOrderFilters_AppendFilter(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unknown Panic in TestOrderFilters_AppendFilter")
		}
	}()

	mapOfFilterer := map[string]Filterer{"f1": &BaseFilter{}, "f2": &BaseFilter{}, "f3": &BaseFilter{}}
	fOrder := OrderFilterer(&OrderFilters{[]string{"f1", "f2", "f3"}, nil})

	fOrder.AppendFilter(mapOfFilterer["f1"])
	if fOrder.GetFilterN(0) != mapOfFilterer["f1"] {
		t.Errorf("filter f1 not set")
	}

	fOrder.AppendFilter(mapOfFilterer["f2"])
	if fOrder.GetFilterN(1) != mapOfFilterer["f2"] {
		t.Errorf("filter f2 not set")
	}

	if fOrder.GetFilterN(0) != mapOfFilterer["f1"] {
		t.Errorf("change filter order")
	}

	fOrder = OrderFilterer(&OrderFilters{[]string{"f1", "f2", "f3"}, nil})
	fOrder.AppendFilter(mapOfFilterer["f1"], mapOfFilterer["f2"])

	if fOrder.GetFilterN(0) != mapOfFilterer["f1"] {
		t.Errorf("filter f1 not set")
	}

	fOrder.AppendFilter(mapOfFilterer["f2"])
	if fOrder.GetFilterN(1) != mapOfFilterer["f2"] {
		t.Errorf("filter f2 not set")
	}

}

func TestOrderFilters_GetFilterN(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unknown Panic in TestOrderFilters_GetFilterN")
		}
	}()
	order := []string{"f1", "f2", "f3"}
	filters := []Filterer{
		&BaseFilter{"f1", nil, nil},
		&BaseFilter{"f2", nil, nil},
		&BaseFilter{"f3", nil, nil}}
	fOrder := OrderFilterer(&OrderFilters{order, filters})
	for i := range fOrder.GetOrderFilters() {
		if filters[i].Info() != fOrder.GetFilterN(i).Info() {
			t.Errorf("Missmath filters! %v !=%v", filters[i].Info(), fOrder.GetFilterN(i).Info())
		}
	}
}

func TestOrderFilters_GenerateFilteredHandleFunc(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unknown Panic in TestOrderFilters_GenerateFilteredHandleFunc")
		}
	}()

	mapOfFilterer := map[string]Filterer{"f1": &BaseFilter{}, "f2": &BaseFilter{}, "f3": &BaseFilter{}}
	fOrder := OrderFilterer(&OrderFilters{[]string{"f1", "f2", "f3"}, nil})
	fOrder.AppendFilter(mapOfFilterer["f1"], mapOfFilterer["f2"])

	f := func(http.ResponseWriter, *http.Request) { return }
	ff := fOrder.GenerateFilteredHandleFunc(f)

	if ff == nil {
		t.Errorf("return nil func")
	}

}
