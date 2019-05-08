package gobeans

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/imperiuse/golib/cast"
	"github.com/mitchellh/mapstructure"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"

	"github.com/imperiuse/golib/concat"
	"github.com/imperiuse/golib/jsonnocomment"
)

// BeanDescription - структура описывающая Bean
//nolint
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
	Type reflect.Type  // содержит рефлексионный тип данных объекта
}

// BeanInstance - структура описывающая Bean объект
type BeanInstance struct {
	jfi interface{}      // содержит изначальный объект после парсинга Json !!! ВНИМАНИЕ СТАТИЧНЫЙ ОБЪЕКТ !!!!
	PIO interface{}      // содержит указатель на интерфейс объекта Bean ОСНОВНОЙ ОБЪЕКТ BEAN РЕКОМЕНДУЕТСЯ РАБОТАТЬ С НИМ!
	r   *reflectInstance // рефлексионное представление объекта (для служебного использования)
	d   *BeanDescription // структура описание bean (для служебного использования)
}

// ClonePIO - метод возращающий clone value of PIO
func (b *BeanInstance) ClonePIO() (interface{}, error) {
	r := reflect.New(b.r.Type).Interface()
	if err := copier.Copy(r, b.PIO); err != nil {
		return nil, errors.WithMessage(err, "Can't clone PIO object by copier.Copy()")
	}
	return r, nil
}

//// CloneJFI - метод возращающий clone value of jfi
//func (b *BeanInstance) CloneJFI() (interface{}, error) {
//	r := reflect.New(b.r.Type).Interface()
//	if err := copier.Copy(r, b.jfi); err != nil {
//		return nil, errors.WithMessage(err, "Can't clone jfi object by copier.Copy()")
//	}
//	return r, nil
//}

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

// GetMapBeans - метод возращающий map Beans
func (bs *BeansStorage) GetMapBeans() MapBeansType {
	return bs.beansMap
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

// getBeanByInterfaceID - метод возращающий объект Bean по ID
func (bs *BeansStorage) getBeanByInterfaceID(v interface{}) (*BeanInstance, error) {
	id, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("can't convert value to string BeanID %v", v)
	}
	bean := bs.GetBean(id)
	if bean == nil {
		return nil, fmt.Errorf("not found Bean by ID: %s", id)
	}
	return bean, nil
}

// GetBeanPIO - метод возращающий объект PIO Bean-а по его ID
func (bs *BeansStorage) GetBeanPIO(id string) interface{} {
	return bs.beansMap[id].PIO
}

// GetCloneBeanPIO - получить клонированный объект PIO Bean-а по его ID
func (bs *BeansStorage) GetCloneBeanPIO(id string) (interface{}, error) {
	return bs.beansMap[id].ClonePIO()
}

//// GetBeanJFI - метод возращающий объект jfi Bean-а по его ID
//func (bs *BeansStorage) GetBeanJFI(id string) interface{} {
//	return bs.beansMap[id].jfi
//}
//
//// GetCloneBeanJFI - метод возращающий клонированный объект jfi Bean-а по его ID
//func (bs *BeansStorage) GetCloneBeanJFI(id string) (interface{}, error) {
//	return bs.beansMap[id].CloneJFI()
//}

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

// BuildBeansInstance - метод создающий, заполняющий и связывающий экземпляры Bean по их описанию - BeanDescription
func (bs *BeansStorage) BuildBeansInstance(beanDescriptions []BeanDescription) error {

	// Построение каркасов Bean
	for i := range beanDescriptions {
		beanDesc := &beanDescriptions[i]
		if beanDesc.Enable {
			if err := bs.addNewBeanInstance(beanDesc); err != nil {
				return err
			}
		}
	}

	// Заполнение каркасов значениями согласно JSON описанию и связывание Bean между собой
	for i := range beanDescriptions {
		beanDesc := &beanDescriptions[i]
		if beanDesc.Enable {
			if err := bs.fillAndLinkBean(beanDesc); err != nil {
				return err
			}
		}
	}

	// Очистка ненужной системной информации
	for _, bean := range bs.GetMapBeans() {
		bean.d = nil
		bean.jfi = nil
	}

	return nil
}

// addNewBeanInstance  -  метод создающий и добавляющий в мапу beanStorage новый Bean согласно его описанию - BeanDescription
func (bs *BeansStorage) addNewBeanInstance(beanDescription *BeanDescription) error {

	s, typ, err := bs.createEmptyBean(beanDescription)
	if err != nil {
		return err
	}

	bs.saveBean(beanDescription, s, typ)

	return nil
}

// saveBean  -  метод сохраняющий в мапу Bean-ов новый Bean
func (bs *BeansStorage) saveBean(d *BeanDescription, s reflect.Value, t reflect.Type) {
	bs.beansMap[d.ID] = &BeanInstance{
		jfi: s.Interface(),
		PIO: s.Addr().Interface(),
		r: &reflectInstance{
			Obj:  s,
			Type: t},
		d: d,
	}
}

// createEmptyBean - метод создающий пустой Bean объект по его описанию - BeanDescription
func (bs *BeansStorage) createEmptyBean(beanDescription *BeanDescription) (reflect.Value, reflect.Type, error) {
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
				return val, typ, fmt.Errorf("not found reflect type by name: %s  [createEmptyBean in Anonymous if]", v.Type)
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
			return val, typ, fmt.Errorf("not found reflect type by name: %s  [createEmptyBean in usual if]", beanDescription.StructName)
		}
	}

	return reflect.New(typ).Elem(), typ, nil
}

// fillAndLinkBean  - метод заполняющий Bean на основе данных и связывающий объект Bean с другими
func (bs *BeansStorage) fillAndLinkBean(beanDescription *BeanDescription) error {

	bean, found := bs.beansMap[beanDescription.ID]
	if !found {
		return fmt.Errorf("not found Bean by ID: %s [fillAndLinkBean]", beanDescription.ID)
	}

	s := bean.r.Obj

	if s.Kind() == reflect.Struct {
		for i, p := range beanDescription.Properties {

			var f = s.FieldByName(p.Name)
			if !f.IsValid() || !f.CanSet() {
				// Поле структуры к которой обращаемся должно быть экспортируемо, т.е. быть public (с большой буквы)
				// A Value can be changed only if it is addressable and was not obtained by  the use of unexported struct fields.
				continue
			}

			var x reflect.Value

			switch p.Type {
			case DeepCopyObj:
				b, err := bs.getBeanByInterfaceID(p.Value)
				if err != nil {
					return errors.WithMessagef(err, "p.Name: %s p.Value: %v beanDescription.ID: %s", p.Name, p.Value, beanDescription.ID)
				}
				x = b.r.Obj

			case PointerToObj:
				b, err := bs.getBeanByInterfaceID(p.Value)
				if err != nil {
					return errors.WithMessagef(err, "p.Name: %s p.Value: %v beanDescription.ID: %s", p.Name, p.Value, beanDescription.ID)
				}
				x = b.r.Obj.Addr()

			case Natural:
				var err error
				x, err = cast.DynamicTypeAssertion(p.Value, f)
				if err != nil {
					return errors.WithMessagef(err, "Can't get reflect value of p.Name: %s, p.Value: %+v BeanID: %s", p.Name, p.Value, beanDescription.ID)
				}

			case BeansObj:
				var bDesc BeanDescription
				var typ reflect.Type
				err := mapstructure.Decode(p.Value, &bDesc)
				if err != nil {
					return errors.WithMessagef(err, "err while convert property[%d] to BeanDescription struct: %+v. p.Name: %s, p.Value: %+v BeanID: %s", i, beanDescription, p.Name, p.Value, beanDescription.ID)
				}

				x, typ, err = bs.createEmptyBean(&bDesc) // создаем внутренний Bean
				if err != nil {
					return errors.WithMessagef(err, "can't create inner Bean [fillAndLinkBean] p.Name: %s, p.Value: %+v BeanID: %s", p.Name, p.Value, beanDescription.ID)
				}

				bs.saveBean(&bDesc, x, typ) // сохраняем внутренний Bean

				err = bs.fillAndLinkBean(&bDesc) // рекурсивно связываем внутренний Bean
				if err != nil {
					return errors.WithMessagef(err, "can't recursive link inner bean [fillAndLinkBean] BeanID: %s", bDesc.ID)
				}
			}
			f.Set(x)
		}
	}

	bs.saveBean(beanDescription, s, bean.r.Type) // обновляем сохраненное

	return nil
}
