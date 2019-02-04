package gobeans

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/jinzhu/copier"

	"github.com/imperiuse/golib/colors"
	"github.com/imperiuse/golib/concat"
	"github.com/imperiuse/golib/jsonnocomment"
)

// BeanID - уникальная строка идентификатор Bean
type BeanID string

// BeansSettingsType - Slice of Beans
type BeansSettingsType []BeanSettings

// BeanSettings - Описатель одного Bean
type BeanSettings struct {
	ID           BeanID        `json:"id"`           // unique id Bean
	Enable       bool          `json:"enable"`       // status Bean, enable or disable
	Struct       string        `json:"struct"`       // name of Go struct
	StructFields []StructField `json:"structFields"` // anonymous struct fields descriptions
	Description  string        `json:"description"`  // description bean
	Properties   []Properties  `json:"properties"`   // bean properties
}

// StructField - Описатель одного поля анонимной структуры @see:reflect.StructField
type StructField struct {
	Name string `json:"name"` // name field
	Type string `json:"type"` // golang type
	Tag  string `json:"tag"`  // golang tag
}

// Типы свойств по аналогии с JAVA BEANS
const (
	Natural      = "nat"  // Простые типы: int, float, string или []T или map[string]T - где T простой тип
	DeepCopyObj  = "copy" // Глубокая копия сложный объект
	PointerToObj = "link" // Ссылка на объект
	BeansObj     = "obj"  // @NOT_USED  (Reserved for recursive initial by Bean inside another Bean obj)
)

// AnonStructPrefixTypeName - Префикс обозначения типа анонимных структур
const AnonStructPrefixTypeName = "github.com/imperiuse/gobeans_anon_struct_"

// Properties - Описатель поля Bean (структуры)
type Properties struct {
	Type  string      `json:"type"`  // тип инициализации см. выше const Natural, DeepCopyObj, PointerToObj, AnonumusObj
	Name  string      `json:"name"`  // имя поля
	Value interface{} `json:"value"` // значение поля, либо ссылка , либо объект
}

// reflectInstance - Описатель объекта в "рефлексионном" представлении
type reflectInstance struct {
	Obj  reflect.Value // содержит сам объект
	Pobj reflect.Value // содержит указатель на объект
	Type reflect.Type  // содержит рефлексионный тип данных объекта
}

// BeanInstance - Описатель объекта в представлении пустого интерфейса
type BeanInstance struct {
	JFI interface{}     // содержит изначальный объект после парсинга Json !!! ВНИМАНИЕ СТАТИЧНЫЙ ОБЪЕКТ !!!!
	PIO interface{}     // содержит указатель на интерфейс объекта Bean ОСНОВНОЙ ОБЪЕКТ BEAN РЕКОМЕНДУЕТСЯ РАБОТАТЬ С НИМ!
	r   reflectInstance // рефлексионное представление объекта
}

// ClonePIO - return clone value of PIO
func (b *BeanInstance) ClonePIO() interface{} {
	r := reflect.New(b.r.Type).Interface()
	if err := copier.Copy(r, b.PIO); err != nil {
		fmt.Print("Can't Clone PIO!")
		return nil
	}
	return r
}

// CloneJFI - return clone value of JFI
func (b *BeanInstance) CloneJFI() interface{} {
	r := reflect.New(b.r.Type).Interface()
	if err := copier.Copy(r, b.JFI); err != nil {
		fmt.Print("Can't Clone JFI!")
		return nil
	}
	return r
}

// MapBeansType - map созданных Bean объектов
type MapBeansType map[BeanID]*BeanInstance

// MapRegistryType - map зарегистрированных типов (рефлексионных)
type MapRegistryType map[string]reflect.Type

// BeansStorage - Главный объект библиотеки - Beans (Хранилище Beans)
type BeansStorage struct {
	typeMap  MapRegistryType
	beansMap MapBeansType
}

// CreateBeanStorage - главный конструктор, главной структуры - хранилища Bean
func CreateBeanStorage() BeansStorage {
	beanStorage := BeansStorage{make(MapRegistryType), make(MapBeansType)}

	// регистрация стандартных необходимых типов (напрямую)
	// (можно было заюзать и вот эту функцию: beanStorage.RegType()) - но хочется экспериментов :)
	beanStorage.typeMap["string"] = reflect.TypeOf("123")
	beanStorage.typeMap["int"] = reflect.TypeOf(123)
	beanStorage.typeMap["float"] = reflect.TypeOf(1.23)

	return beanStorage
}

// GetAllNamesRegistryTypes - функция которая возращает slice имен зарегистрированных типов
func (bs BeansStorage) GetAllNamesRegistryTypes() (nameRegistryType []string) {
	for nameType := range bs.typeMap {
		nameRegistryType = append(nameRegistryType, nameType)
	}
	return
}

// GetIDBeans - функция которая возращает slice BeanID (имён бинов) хранимых bean типов
func (bs BeansStorage) GetIDBeans() (beansIDs []BeanID) {
	for beanID := range bs.beansMap {
		beansIDs = append(beansIDs, beanID)
	}
	return
}

// GetBean - получить объект Bean по ID
func (bs BeansStorage) GetBean(id BeanID) *BeanInstance {
	return bs.beansMap[id]
}

// GetBeanPIO - получить объект PIO Bean-а по его ID
func (bs BeansStorage) GetBeanPIO(id BeanID) interface{} {
	return bs.beansMap[id].PIO
}

// GetCloneBeanPIO - получить клонированный объект PIO Bean-а по его ID
func (bs BeansStorage) GetCloneBeanPIO(id BeanID) interface{} {
	return bs.beansMap[id].ClonePIO()
}

// GetBeanJFI - получить объект JFI Bean-а по его ID
func (bs BeansStorage) GetBeanJFI(id BeanID) interface{} {
	return bs.beansMap[id].JFI
}

// GetCloneBeanJFI - получить клонированный объект JFI Bean-а по его ID
func (bs BeansStorage) GetCloneBeanJFI(id BeanID) interface{} {
	return bs.beansMap[id].CloneJFI()
}

// GetMapBeans - получить map Beans
func (bs BeansStorage) GetMapBeans() MapBeansType {
	return bs.beansMap
}

// GetReflectTypeByName - получить reflect.Type по typeName
func (bs BeansStorage) GetReflectTypeByName(typeName string) reflect.Type {
	return bs.typeMap[typeName]
}

// FoundAndGetReflectTypeByName - получить reflect.Type по typeName
func (bs BeansStorage) FoundAndGetReflectTypeByName(typeName string) (reflect.Type, bool) {
	typ, found := bs.typeMap[typeName]
	return typ, found
}

// RegType - Метод регистратор объектов в глобальную переменную typeRegistry  ( map[string]reflect.Type )
func (bs BeansStorage) RegType(typesNil ...interface{}) error {
	for _, typeT := range typesNil {
		if err := bs.regType(typeT); err != nil {
			return err
		}
	}
	return nil
}

// ShowRegTypes - Метод возращает список зарегистрированных названий типов
func (bs BeansStorage) ShowRegTypes() []string {
	return bs.GetAllNamesRegistryTypes()
}

// regType - промежуточная оберточная функция для перехвата возможной panic
func (bs BeansStorage) regType(typeT interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error while registrate %v. Err: %v ", typeT, r)
		}
	}()
	bs.unsafeRegType(typeT)
	return nil
}

// unsafeRegType - базовая функция регистрации типа, может паниковать! поэтому небезопасна!
// !ATTENTION! can panic use at list func regType(typeT interface{}) (err error)    @see upper
func (bs BeansStorage) unsafeRegType(typeT interface{}) {
	// 3 - it's mean support only *T pointer's (поддержка указателей на объект с именем T)
	pkgName, typeName := recursiveGetPkgAndTypeName(reflect.TypeOf(typeT).Elem().PkgPath(),
		reflect.TypeOf(typeT).Elem().Name(), reflect.TypeOf(typeT).Elem(), 3)
	if pkgName != "" {
		bs.typeMap[pkgName+"."+typeName] = reflect.TypeOf(typeT).Elem()
	} else {
		bs.typeMap[typeName] = reflect.TypeOf(typeT).Elem()
	}
}

// recursiveGetPkgAndTypeName - функция для рекурсивного получения имени типа и имени пакета по его reflect.Type
// @param:
// 			 typeName                 string        - Начальное предполагаемое имя типа (T или *T или ***T),
//                                                    полученное с помощью метода reflect.TypeOf(typeT).Elem().Name()
// 			 typ                      reflect.Type  - Значение типа спец. тип рефлексии типа
// 			 maxCntRecEstimate        int           - максимальный уровень рекурсии (3 соотв. *T)
// @return:
// 			 pckgName                 string        - Имя пакета
// 			 typeName                 string        - Имя типа
// @other:
//        Уровень рекурсии ограничен 3 т.к. это соотв. распарсиванию указателя на T, т.е. *T,
//        а больше для JSON конфига и не нужно
//nolint
func recursiveGetPkgAndTypeName(pkgname, typeName string, typ reflect.Type, maxCntRecEstimate int) (string, string) {
	maxCntRecEstimate--
	if maxCntRecEstimate > 0 {
		if typeName == "" {
			typ = typ.Elem()
			var t string
			pkgname, t = recursiveGetPkgAndTypeName(typ.PkgPath(), typ.Name(), typ, maxCntRecEstimate)
			typeName = typeName + "*" + t
		} else {
			return pkgname, typeName
		}
	} else {
		panic(fmt.Errorf("estimate max try recursion call func for param %v ", typ))
	}
	return pkgname, typeName
}

// CreateBeansFromJSON - Метод заполнения хранилища Beans, Bean созданными по данным из JSON файла
// Читает из конфига строки, Unmarshal в структуру BeansSettingsType, далее построение Bean по описанию
// @param
//		pathFile       string    - путь к файлу json описания Beans
// @return
// 		err            error     - ошибка при создании объектов по данным JSON или ошибка при чтении JSON файла
func (bs BeansStorage) CreateBeansFromJSON(pathFile string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error while create map of beans. Err: %v ", r)
		}
	}()

	// Чтения из файла pathFile Beans Settings
	BeansSettingsFile, err := jsonnocomment.ReadFileAndCleanComment(pathFile)
	colors.CheckErrorFunc(err, fmt.Sprintf("Read File: %v", pathFile))
	if err != nil {
		return fmt.Errorf("can't open and read file: %v. Err: %v ", pathFile, err)
	}

	BeansSlice := BeansSettingsType{}
	// Unmarshal байтового представления Bean File в структуру
	err = json.Unmarshal(BeansSettingsFile, &BeansSlice)
	colors.CheckErrorFunc(err, fmt.Sprintf("Unmarshal Beans Settings from file: %v", pathFile))
	if err != nil {
		return fmt.Errorf("can't parse: %v. Err: %v. \nFile: %v ", pathFile, err, BeansSettingsFile)
	}

	// Пробегаемся по слайсу первый раз строим объекты и простые поля
	for _, beanSettings := range BeansSlice {
		if beanSettings.Enable { // если Bean включен
			bs.addNewBeanInstance(false, beanSettings)
		}
	}

	// Пробегаемся второй раз строим ссылки на сложные объекты или делаем их deep copy
	for _, beanSettings := range BeansSlice {
		if beanSettings.Enable { // если Bean включен
			bs.addNewBeanInstance(true, beanSettings)
		}
	}

	return nil
}

// addNewBeanInstance  - Метод добавления нового Bean экземляра заданного согласно BeanSettings
// @param
// 			processingReferenceOn  bool          -  режим обработки (флаг - "Обработка ссылок на объекты" - Вкл(True)/Выкл(False))
// 			beanSettings           BeanSetting   -  настройки генерируемого Bean инстанса
// @return
//         nil
//         Подробнее читай Язык ПРограммирования Go  А. Донован, Б. Керниган стр. 397.
// @other help info:
//// 			bean.id                     BeanID        -  уникальное имя объекта Bean
//// 			bean.structName             string        -  имя типа GO для Bean
//// 			bean.StructFields           []StructField -  анонимной runtime генерирующейся структуры
//// 			bean.properties             []Properties  -  свойства Bean которыми инициализурется объект после создания
//nolint
func (bs BeansStorage) addNewBeanInstance(processingFieldsAndReference bool, beanSettings BeanSettings) {
	var typ reflect.Type
	var s reflect.Value
	if beanSettings.Struct == "" { // Анонимная структура
		structFields := make([]reflect.StructField, len(beanSettings.StructFields))
		for i, descStructField := range beanSettings.StructFields {
			if tempType, found := bs.FoundAndGetReflectTypeByName(descStructField.Type); found {
				structFields[i] = reflect.StructField{
					Name: descStructField.Name,
					Type: tempType,
					Tag:  reflect.StructTag(descStructField.Tag)}
			} else {
				panic(fmt.Errorf("Not found Struct by Name (Anonumous Struct): %v", descStructField.Type))
			}
		}
		typ = reflect.StructOf(structFields)
		// TODO!!! Регистрация типа анонимной структуры
		bs.typeMap[concat.Strings(AnonStructPrefixTypeName, string(beanSettings.ID))] = typ
	} else { // Если объект не является анонимной структурой
		if tempType, found := bs.FoundAndGetReflectTypeByName(beanSettings.Struct); found {
			typ = tempType
		} else {
			panic(fmt.Errorf("Not found Struct by Name: %v", beanSettings.Struct))
		}
	}

	if s = reflect.New(typ).Elem(); s.Kind() == reflect.Struct {
		if processingFieldsAndReference {
			for _, p := range beanSettings.Properties {
				switch p.Type {
				case DeepCopyObj:
					x := bs.beansMap[BeanID(p.Value.(string))]
					s.FieldByName(p.Name).Set(x.r.Obj)
					//fmt.Println(s.FieldByName(p.Name).Type())  // test get field Type Name
					//fmt.Println(s.FieldByName(p.Name).Type().Name())
				case PointerToObj:
					x := bs.beansMap[BeanID(p.Value.(string))]
					s.FieldByName(p.Name).Set(x.r.Pobj)
					//fmt.Println(s.FieldByName(p.Name).Type().Elem().Name())
					//fmt.Println(s.FieldByName(p.Name).Type().Name())   // test get field Type Name
				case Natural:
					f := s.FieldByName(p.Name)
					// Поле структуры к которой обращаемся должно быть экспортируемо, т.е. быть public (с большой буквы)
					if f.IsValid() {
						// A Value can be changed only if it is
						// addressable and was not obtained by
						// the use of unexported struct fields.
						if f.CanSet() {
							// Ниже определение "макро" тип поля структуры
							switch f.Kind() {
							case reflect.Bool:
								if x, ok := p.Value.(bool); ok {
									f.SetBool(x)
								}
							case reflect.Int:
								if xf, ok := p.Value.(float64); ok { // float64 НЕ ошибка, так надо!
									x := int64(xf)
									if !f.OverflowInt(x) { // Проверка, что значение не переполняет тип
										f.SetInt(x)
									}
								}
							case reflect.Float64:
								if x, ok := p.Value.(float64); ok {
									if !f.OverflowFloat(x) { // Проверка, что значение не переполняет тип
										f.SetFloat(x)
									}
								}
							case reflect.String:
								if x, ok := p.Value.(string); ok {
									f.SetString(x)
								}
							case reflect.Slice: // ОГРАНИЧЕНИЯ для SLICE - только: []string, []float64
								pI := p.Value.([]interface{})
								f.Set(reflect.MakeSlice(f.Type(), len(pI), len(pI)))
								for i, v := range pI {
									f.Index(i).Set(reflect.ValueOf(v))
								}
							case reflect.Map: // ОГРАНИЧЕНИЯ для MAP - только: map[string]T
								f.Set(reflect.MakeMap(f.Type()))
								for k, v := range p.Value.(map[string]interface{}) {
									f.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
								}
							}
						}
					}
				case BeansObj: // TODO @UNUSED, пока сделано что вначале строится список всех BEAN
					// (флаг processingFieldsAndReference) , затем идет, заполнение значений
				}
			}
		}
	}
	bs.beansMap[beanSettings.ID] = &BeanInstance{s.Interface(), s.Addr().Interface(), reflectInstance{Obj: s, Pobj: s.Addr(), Type: typ}}
	return
}
