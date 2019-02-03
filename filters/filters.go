package filters

import (
	"fmt"
	"net/http"
)

// BaseFilter - Базовая структура фильтра
type BaseFilter struct {
	Name        string    // наименование фильтра
	nextFilter  *Filterer // ссылка на следующий фильтр
	selfPointer *Filterer // ссылка на себя
}

// Filterer - базовый интерфейс фильтров
type Filterer interface {
	GetBaseFilter() Filterer // Метод получения родителя

	SetNextFilter(*Filterer)  // Метод установки указателя на следующий фильтр (интерфейс фильтра)
	SetSelfPointer(*Filterer) // Метод установки указателя на себя (указатель на интерфейс себя)

	GetNextFilter() *Filterer  // Метод получения указателя на следующий фильтр (интерфейс фильтра)
	GetSelfPointer() *Filterer // Метод получения указателя на себя (указатель на интерфейс себя)

	Info() string // Метод получения информации по фильтру

	Before(http.ResponseWriter, *http.Request) // 1 Вспомогательный метод - предварительный метод фильтра
	Filter(http.ResponseWriter, *http.Request) // Основной метод фильтра
	After(http.ResponseWriter, *http.Request)  // 2 Вспомогательный метод - заключительный метод фильтра

	// Метод стартующий выполнение методов фильтра, в которой последовательно вызваются методы:
	// 			Before()->Filter()->GetNextFilter().Run()->After()
	//  При возникновении ошибки вызывается метод ErrorHandler
	Run(http.ResponseWriter, *http.Request, func(http.ResponseWriter, *http.Request))
	GeneratorDeferRunFunc(http.ResponseWriter, *http.Request) func() // Генератор Defer для функции Run
	ErrorHandler(http.ResponseWriter, *http.Request, interface{})    // Метод вызывающийся в случае ошибки на уровне
	// фильтра т.е. при вызове функций Before, After, Filter
}

// GetBaseFilter - метод возразает родительскую структуру фильтра (BaseFilter)
func (f *BaseFilter) GetBaseFilter() Filterer {
	return f
}

// Info - метод возращает информацию по фильтру
func (f *BaseFilter) Info() string {
	return fmt.Sprintf("[BaseFilter] Info():\n%+v", f)
}

// SetNextFilter - метод для установки указателя на интерфейс след. фильтра
//nolint
func (f *BaseFilter) SetNextFilter(nf *Filterer) {
	f.nextFilter = nf
}

// SetSelfPointer - метод для установки указателя на себя (указатель на интерфейс себя)
//nolint
func (f *BaseFilter) SetSelfPointer(self *Filterer) {
	f.selfPointer = self
}

// GetNextFilter - метод получения текущего указателя на интерфейс след. фильтра
//nolint
func (f *BaseFilter) GetNextFilter() *Filterer {
	return f.nextFilter
}

// GetSelfPointer - метод получения текущего указателя себя (указатель на интерфейс себя)
//nolint
func (f *BaseFilter) GetSelfPointer() *Filterer {
	return f.selfPointer
}

// Before - метод фильтра вызывающийся до Filter
func (f *BaseFilter) Before(http.ResponseWriter, *http.Request) {
	//fmt.Println("[BaseFilter] Before()")
}

// After - метод фильтра вызывающийся после Filter
func (f *BaseFilter) After(http.ResponseWriter, *http.Request) {
	//fmt.Println("[BaseFilter] After()")
}

// Filter - основной метод фильтра
func (f *BaseFilter) Filter(http.ResponseWriter, *http.Request) {
	//fmt.Println("[BaseFilter] Filter()")
	//fmt.Println((*f.selfPointer).Info())
}

// ErrorHandler - метод фильтра вызывающийся в случае ошибки на уровне фильтра
func (f *BaseFilter) ErrorHandler(w http.ResponseWriter, r *http.Request, err interface{}) {
	//fmt.Println("[BaseFilter] ErrorHandler()", err)
}

// GeneratorDeferRunFunc - метод создания базового дефера с обработкой паники с помощью вызова функции ErrorHandler
func (f *BaseFilter) GeneratorDeferRunFunc(w http.ResponseWriter, r *http.Request) func() {
	return func() {
		if rec := recover(); rec != nil {
			fmt.Println(fmt.Sprintf("[BaseFilter] Generator Defer.\t"+
				"Problem in func filter: %v."+
				"\tErr: %v", (*f.GetSelfPointer()).Info(), rec))
			(*f.GetSelfPointer()).ErrorHandler(w, r, rec)
		}
	}
}

// Run - метод стартующий вызов других методов фильтра
func (f *BaseFilter) Run(w http.ResponseWriter, r *http.Request, businessFunc func(http.ResponseWriter, *http.Request)) {
	defer (*f.GetSelfPointer()).GeneratorDeferRunFunc(w, r)()

	(*f.GetSelfPointer()).Before(w, r)
	(*f.GetSelfPointer()).Filter(w, r)

	if NextFilterInterface := (*f.GetSelfPointer()).GetNextFilter(); NextFilterInterface != nil {
		(*NextFilterInterface).Run(w, r, businessFunc)
	} else {
		businessFunc(w, r)
	}

	(*f.GetSelfPointer()).After(w, r)
}
