package gobeans

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"

	"github.com/imperiuse/golib/concat"
	"github.com/imperiuse/golib/jsonnocomment"
)

// BeanDescription - структура описывающая Bean
type BeanDescription struct {
	// IMPORTANT FIELDS
	ID         string       `json:"id"`         // Bean ID - MUST BE UNIQUE!
	Enable     bool         `json:"enable"`     // Enable for create (build) this Bean obj ?
	StructName string       `json:"structName"` // Registered structure name (@see func RegTypes and RegNamedTypes)
	Properties []Properties `json:"properties"` // bean properties

	// OTHERS FIELDS
	Description string `json:"description"` // Simple text description of Bean (doesn't matter)

	// Anonymous structs
	Anonymous    bool          `json:"anonymous"`    // Bean base on Anonymous struct ? (Default = false)
	StructFields []StructField `json:"structFields"` // anonymous struct fields descriptions (if Struct == "")
}

// StructField - структура описывающая одно поле анонимной структуры @see:reflect.StructField
type StructField struct {
	Name string `json:"name"` // name field
	Type string `json:"type"` // golang type name
	Tag  string `json:"tag"`  // golang tag
}

// Типы свойств по аналогии с JAVA BEANS
const (
	Natural      = "nat"  // Простые типы: int, float, string или []T или map[string]T - где T простой тип
	DeepCopyObj  = "copy" // Глубокая копия сложный объект
	PointerToObj = "link" // Ссылка на объект
	BeansObj     = "obj"  // Включение вложенного Bean объекта
)

// AnonStructPrefixTypeName - Префикс обозначения типа анонимных структур
const AnonStructPrefixTypeName = "github.com/imperiuse/gobeans_anon_struct_"

// Properties - структура описывающая поля Bean
type Properties struct {
	Type  string      `json:"type"`  // тип инициализации см. выше const Natural, DeepCopyObj, PointerToObj, BeansObj
	Name  string      `json:"name"`  // имя поля
	Value interface{} `json:"value"` // значение поля, либо ссылка , либо объект
}

// reflectInstance - структура описывающая объект Bean в "рефлексионном" представлении
type reflectInstance struct {
	Obj  reflect.Value // содержит сам объект
	Pobj reflect.Value // содержит указатель на объект
	Type reflect.Type  // содержит рефлексионный тип данных объекта
}

// BeanInstance - структура описывающая Bean объект
type BeanInstance struct {
	JFI interface{}     // содержит изначальный объект после парсинга Json !!! ВНИМАНИЕ СТАТИЧНЫЙ ОБЪЕКТ !!!!
	PIO interface{}     // содержит указатель на интерфейс объекта Bean ОСНОВНОЙ ОБЪЕКТ BEAN РЕКОМЕНДУЕТСЯ РАБОТАТЬ С НИМ!
	r   reflectInstance // рефлексионное представление объекта
}

// ClonePIO - метод возращающий clone value of PIO
func (b *BeanInstance) ClonePIO() (interface{}, error) {
	r := reflect.New(b.r.Type).Interface()
	if err := copier.Copy(r, b.PIO); err != nil {
		return nil, errors.WithMessage(err, "Can't clone PIO object by copier.Copy()")
	}
	return r, nil
}

// CloneJFI - метод возращающий clone value of JFI
func (b *BeanInstance) CloneJFI() (interface{}, error) {
	r := reflect.New(b.r.Type).Interface()
	if err := copier.Copy(r, b.JFI); err != nil {
		return nil, errors.WithMessage(err, "Can't clone JFI object by copier.Copy()")
	}
	return r, nil
}

// MapBeansType - map созданных Bean объектов
type MapBeansType map[string]*BeanInstance

// MapRegistryType - map зарегистрированных типов (рефлексионных)
type MapRegistryType map[string]reflect.Type

// BeansStorage - Главный объект библиотеки - Beans (Хранилище Beans)
type BeansStorage struct {
	typeMap  MapRegistryType
	beansMap MapBeansType
}

// CreateBeanStorage - главный конструктор, главной структуры - хранилища Bean
func CreateBeanStorage() (BeansStorage, error) {

	// Регистрация базовых стандартных типов
	golangInlineTypes := []interface{}{
		(*int8)(nil), (*int16)(nil), (*int32)(nil), (*int64)(nil), (*int)(nil),
		(*uint8)(nil), (*uint16)(nil), (*uint32)(nil), (*uint64)(nil), (*uint)(nil),
		(*float32)(nil), (*float64)(nil), (*complex64)(nil), (*complex128)(nil),
		(*byte)(nil), (*rune)(nil), (*string)(nil), (*bool)(nil)}

	beanStorage := BeansStorage{make(MapRegistryType, len(golangInlineTypes)), make(MapBeansType)}

	// Пример. Регистрация стандартных необходимых типов (напрямую)
	//beanStorage.typeMap["string"] = reflect.TypeOf("123")
	//beanStorage.typeMap["int"] = reflect.TypeOf(123)
	//beanStorage.typeMap["float"] = reflect.TypeOf(1.23)

	// Но будем делать это в общем виде :)
	err := beanStorage.RegTypes(golangInlineTypes...)

	return beanStorage, err
}

// GetAllNamesRegistryTypes - метод возращающий slice имен зарегистрированных типов
func (bs *BeansStorage) GetAllNamesRegistryTypes() (nameRegistryType []string) {
	for nameType := range bs.typeMap {
		nameRegistryType = append(nameRegistryType, nameType)
	}
	return
}

// GetIDBeans - метод возращающий slice BeanID (имён бинов) хранимых bean типов
func (bs *BeansStorage) GetIDBeans() []string {
	beansIDs := make([]string, len(bs.beansMap))
	for beanID := range bs.beansMap {
		beansIDs = append(beansIDs, beanID)
	}
	return beansIDs
}

// GetBean - метод возращающий объект Bean по ID
func (bs *BeansStorage) GetBean(id string) *BeanInstance {
	return bs.beansMap[id]
}

// GetBeanPIO - метод возращающий объект PIO Bean-а по его ID
func (bs *BeansStorage) GetBeanPIO(id string) interface{} {
	return bs.beansMap[id].PIO
}

// GetCloneBeanPIO - получить клонированный объект PIO Bean-а по его ID
func (bs *BeansStorage) GetCloneBeanPIO(id string) (interface{}, error) {
	return bs.beansMap[id].ClonePIO()
}

// GetBeanJFI - метод возращающий объект JFI Bean-а по его ID
func (bs *BeansStorage) GetBeanJFI(id string) interface{} {
	return bs.beansMap[id].JFI
}

// GetCloneBeanJFI - метод возращающий клонированный объект JFI Bean-а по его ID
func (bs *BeansStorage) GetCloneBeanJFI(id string) (interface{}, error) {
	return bs.beansMap[id].CloneJFI()
}

// GetMapBeans - метод возращающий map Beans
func (bs *BeansStorage) GetMapBeans() MapBeansType {
	return bs.beansMap
}

// GetReflectTypeByName - метод возращающий reflect.Type по typeName
func (bs *BeansStorage) GetReflectTypeByName(typeName string) reflect.Type {
	return bs.typeMap[typeName]
}

// FoundReflectTypeByName - метод возращающий reflect.Type по typeName и проверящий его наличие
func (bs *BeansStorage) FoundReflectTypeByName(typeName string) (reflect.Type, bool) {
	typ, found := bs.typeMap[typeName]
	return typ, found
}

// RegTypes - метод  регистрирующий типы в MapRegistryType, именует согласно пути пакета
func (bs *BeansStorage) RegTypes(typesNil ...interface{}) error {
	for _, typeT := range typesNil {
		if err := bs.regType(typeT, ""); err != nil {
			return err
		}
	}
	return nil
}

// RegNamedTypes -  метод регистрирующий типы в MapRegistryType, и именует согласно переданному имени, нечетные типы, четные имя типа
func (bs *BeansStorage) RegNamedTypes(typesAndNames ...interface{}) error {
	lenTaN := len(typesAndNames)
	if lenTaN%2 != 0 {
		return fmt.Errorf("mistmatch count type and their names. len args not even!: %d", lenTaN)
	}

	for i := 0; ; {
		typeT := typesAndNames[i]
		if nameT, ok := typesAndNames[i+1].(string); ok {
			if err := bs.regType(typeT, nameT); err != nil {
				return err
			}
			if i += 2; i >= lenTaN {
				break
			}
		} else {
			return fmt.Errorf("can't type assertion to string this arg[%d]: %+v", i+1, typesAndNames[i+1])
		}
	}
	return nil
}

// ShowRegTypes - метод возращающий список зарегистрированных названий типов
func (bs *BeansStorage) ShowRegTypes() []string {
	return bs.GetAllNamesRegistryTypes()
}

// regType - промежуточная оберточная функция для перехвата возможной panic
func (bs *BeansStorage) regType(typeT interface{}, nameT string) (err error) {

	if nameT != "" {
		// эта проверка вероятно важна, т.к. человек может перетереть уже сохраненный тип
		if _, found := bs.typeMap[nameT]; found {
			return fmt.Errorf("type with this name: '%s' - alredy registreted", nameT)
		}

		bs.typeMap[nameT] = reflect.TypeOf(typeT).Elem()
		return nil
	}

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
func (bs *BeansStorage) unsafeRegType(typeT interface{}) {
	// 3 - it's mean support only *T pointer's (поддержка указателей на объект с именем T)
	pkgName, typeName := recursiveGetPkgAndTypeName(reflect.TypeOf(typeT).Elem().PkgPath(),
		reflect.TypeOf(typeT).Elem().Name(), reflect.TypeOf(typeT).Elem(), 3)
	if pkgName != "" { // ниже проверка на существование типа (ключа) не важна, т.к. мы сами строим его
		bs.typeMap[pkgName+"."+typeName] = reflect.TypeOf(typeT).Elem()
	} else {
		bs.typeMap[typeName] = reflect.TypeOf(typeT).Elem()
	}
}

// recursiveGetPkgAndTypeName - функция для рекурсивного получения имени типа и имени пакета по его reflect.Type
// @param:
//           pkgName				  string        - имя пакета
// 			 typeName                 string        - Начальное предполагаемое имя типа (T или *T или ***T),
//                                                    полученное с помощью метода reflect.TypeOf(typeT).Elem().Name()
// 			 typ                      reflect.Type  - Значение типа спец. тип рефлексии типа
// 			 maxCntRecEstimate        int           - максимальный уровень рекурсии (3 соотв. *T)
// @return:
// 			 pckgName                 string        - Имя пакета
// 			 typeName                 string        - Имя типа
// @other:
//        Уровень рекурсии ограничен maxCntRecEstimate
//nolint
func recursiveGetPkgAndTypeName(pkgName, typeName string, typ reflect.Type, maxCntRecEstimate int) (string, string) {
	maxCntRecEstimate--
	if maxCntRecEstimate > 0 {
		if typeName == "" {
			typ = typ.Elem()
			t := ""
			pkgName, t = recursiveGetPkgAndTypeName(typ.PkgPath(), typ.Name(), typ, maxCntRecEstimate)
			typeName = concat.Strings(typeName, "*", t)
		} else {
			return pkgName, typeName
		}
	} else {
		// todo in future, may be better use error, but now (recursion case) fast and simple use panic
		panic(fmt.Errorf("estimate max try recursion call func for param %v ", typ))
	}
	return pkgName, typeName
}

// CreateBeansFromJSON - метод возращающий заполняющий хранилище Bean по данным из JSON файла
func (bs *BeansStorage) CreateBeansFromJSON(pathFile string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic! in CreateBeansFromJSON. Err: %v ", r)
		}
	}()

	BeansSettingsFile, err := jsonnocomment.ReadFileAndCleanComment(pathFile)
	if err != nil {
		return errors.WithMessagef(err, "can't open and read file: %v. ", pathFile)
	}

	BeanDescriptions := make([]BeanDescription, 0)
	err = json.Unmarshal(BeansSettingsFile, &BeanDescriptions)
	if err != nil {
		return errors.WithMessagef(err, "can't parse (Unmarshal) file: %v.  Data: %v ", pathFile, BeansSettingsFile)
	}

	err = bs.BuildBeansInstance(BeanDescriptions)
	if err != nil {
		return errors.WithMessagef(err, "can't build some Bean's instance")
	}

	return nil
}

// BuildBeansInstance - метод создающий экземпляры Bean по их описанию
func (bs *BeansStorage) BuildBeansInstance(beanDescriptions []BeanDescription) error {

	// TODO

	// Пробегаемся по слайсу первый раз строим объекты и простые поля
	for _, beanSettings := range beanDescriptions {
		if beanSettings.Enable { // если Bean включен
			if err := bs.addNewBeanInstance(beanSettings); err != nil {
				return err
			}
		}
	}

	// Пробегаемся второй раз строим ссылки на сложные объекты или делаем их deep copy
	for _, beanSettings := range beanDescriptions {
		if beanSettings.Enable { // если Bean включен
			if err := bs.fillAndLinkBean(beanSettings); err != nil {
				return err
			}
		}
	}

	return nil
}

// addNewBeanInstance  - todo name метод нового Bean экземляра заданного согласно его описанию - BeanDescription
func (bs *BeansStorage) addNewBeanInstance(beanDescription BeanDescription) error {

	// TODO проверить двойное создание объектов
	s, typ, err := bs.createEmptyBean(beanDescription)
	if err != nil {
		return err
	}

	// Сохраняем в Map-у Bean очередной объект
	bs.beansMap[beanDescription.ID] = &BeanInstance{
		s.Interface(),
		s.Addr().Interface(),
		reflectInstance{Obj: s, Pobj: s.Addr(), Type: typ}}

	return nil
}

// createEmptyBean - метод создающий пустой Bean объект по его описанию - BeanDescription
func (bs *BeansStorage) createEmptyBean(beanDescription BeanDescription) (reflect.Value, reflect.Type, error) {
	var typ reflect.Type
	var val reflect.Value

	if beanDescription.Anonymous {
		structFields := make([]reflect.StructField, len(beanDescription.StructFields))
		for i, v := range beanDescription.StructFields {
			if tempType, found := bs.FoundReflectTypeByName(v.Type); found {
				structFields[i] = reflect.StructField{
					Name: v.Name,
					Type: tempType,
					Tag:  reflect.StructTag(v.Tag)}
			} else {
				return val, typ, fmt.Errorf("not found reflect type by name: %v  [in Anonymous if]", v.Type)
			}
		}

		// Создаем новый объект по описанию structFields
		typ = reflect.StructOf(structFields)

		// Регистрация нового созданного типа - анонимной структура на месте
		if beanDescription.StructName == "" {
			beanDescription.StructName = concat.Strings(AnonStructPrefixTypeName, string(beanDescription.ID))
		}
		bs.typeMap[beanDescription.StructName] = typ

	} else { // Если объект не является анонимной структурой
		var found bool
		typ, found = bs.FoundReflectTypeByName(beanDescription.StructName)
		if !found {
			return val, typ, fmt.Errorf("not found reflect type by name: %v  [in usual if]", beanDescription.StructName)
		}
	}

	return reflect.New(typ).Elem(), typ, nil
}

// fillAndLinkBean  - метод заполняющий Bean на основе данных и связывающий объект Bean с другими
func (bs *BeansStorage) fillAndLinkBean(beanDescription BeanDescription) error {
	return nil
}

//
//	if s.Kind() == reflect.Struct {
//		for i, p := range beanDescription.Properties {
//			switch p.Type {
//			case DeepCopyObj:
//				if processingFieldsAndReference {
//					x := bs.beansMap[p.Value.(string)]
//					s.FieldByName(p.Name).Set(x.r.Obj)
//				}
//				//fmt.Println(s.FieldByName(p.Name).Type())  // test get field Type Name
//				//fmt.Println(s.FieldByName(p.Name).Type().Name())
//			case PointerToObj:
//				if processingFieldsAndReference {
//					x := bs.beansMap[p.Value.(string)]
//					s.FieldByName(p.Name).Set(x.r.Pobj)
//				}
//				//fmt.Println(s.FieldByName(p.Name).Type().Elem().Name())
//				//fmt.Println(s.FieldByName(p.Name).Type().Name())   // test get field Type Name
//			case Natural:
//				if processingFieldsAndReference {
//					f := s.FieldByName(p.Name)
//					// Поле структуры к которой обращаемся должно быть экспортируемо, т.е. быть public (с большой буквы)
//					if f.IsValid() {
//						// A Value can be changed only if it is
//						// addressable and was not obtained by
//						// the use of unexported struct fields.
//						if f.CanSet() {
//							// Ниже определение "макро" тип поля структуры
//							switch f.Kind() {
//							case reflect.Bool:
//								if x, ok := p.Value.(bool); ok {
//									f.SetBool(x)
//								}
//							case reflect.Int:
//								if xf, ok := p.Value.(float64); ok { // float64 НЕ ошибка, так надо!
//									x := int64(xf)
//									if !f.OverflowInt(x) { // Проверка, что значение не переполняет тип
//										f.SetInt(x)
//									}
//								}
//							case reflect.Float64:
//								if x, ok := p.Value.(float64); ok {
//									if !f.OverflowFloat(x) { // Проверка, что значение не переполняет тип
//										f.SetFloat(x)
//									}
//								}
//							case reflect.String:
//								if x, ok := p.Value.(string); ok {
//									f.SetString(x)
//								}
//							case reflect.Slice: // ОГРАНИЧЕНИЯ для SLICE - только: []string, []float64
//								pI := p.Value.([]interface{})
//								f.Set(reflect.MakeSlice(f.Type(), len(pI), len(pI)))
//								for i, v := range pI {
//									f.Index(i).Set(reflect.ValueOf(v))
//								}
//							case reflect.Map: // ОГРАНИЧЕНИЯ для MAP - только: map[string]T
//								f.Set(reflect.MakeMap(f.Type()))
//								for k, v := range p.Value.(map[string]interface{}) {
//									f.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
//								}
//							}
//						}
//					}
//				}
//			case BeansObj:
//				var beanSettings BeanDescription
//				if err := mapstructure.Decode(p.Value, &beanSettings); err != nil {
//					return errors.WithMessagef(err, "err while convert property[%d] to BeanDescription struct: %+v", i, p.Value)
//				}
//
//				if err := bs.addNewBeanInstance(processingFieldsAndReference, beanSettings); err != nil {
//					return err
//				}
//				// TODO подумать над этим местом тут копия или все таки ссылка ???????
//				// TODO подумать над абстрактной фабрикой объектов + подумать над реорганизацией кода
//				x := bs.beansMap[beanSettings.ID]
//				s.FieldByName(p.Name).Set(x.r.Obj)
//			}
//		}
//	}
//
//}
