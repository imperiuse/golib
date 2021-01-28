package orm

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// !DISCLAIMER!
// I know, that use reflect in most cases bad and slow.
// Also I realize that yet-another ORM, it is not go-way.
// But for me, using this small library for create micro ORM good choice in many cases
// Of course, if you do not think about serious optimization
//
// This micro library use the same approach like sqlx with `db` tag parsing
//
// Before start work with this package
// You should call function InitMetaTagInfoCache() for warm and init cache (cacheMetaDTO)

type (
	Column   = string
	Table    = string
	Typ      = string
	Join     = string
	Alias    = string
	Argument = interface{}

	structName       = string
	ormUseInTagValue = string

	MetaDTO = struct {
		ColsMap    map[ormUseInTagValue][]Column
		Aliases    []Alias
		Join       Join
		TableName  Table
		StructName Typ
	}
)

var (
	m            sync.RWMutex
	cacheMetaDTO = map[structName]*MetaDTO{}
)

// custom tag for "sugar" columns values prepare for Update squirrel library staff
const (
	tagDB           = "db" // must be for all DTO structs fields, this tag also used by sqlx
	tagOrmUseIN     = "orm_use_in"
	tagOrmAlias     = "orm_alias"
	tagOrmJoin      = "orm_join"
	tagOrmTableName = "orm_table_name"

	ormUseInSelect = "select"
	ormUseInCreate = "create"
	ormUseInUpdate = "update"

	emptyRootAlias = ""
)

func InitMetaTagInfoCache(objs ...interface{}) {
	m.Lock()
	defer m.Unlock()

	for _, obj := range objs {
		if obj == nil {
			continue
		}
		cacheMetaDTO[getObjTypeNameByReflect(obj)] = getNoneCacheMetaDTO(obj)
	}
}

func GetMetaDTO(obj interface{}) *MetaDTO {
	return getMetaDTO(getObjTypeNameByReflect(obj), obj)
}

func getMetaDTO(structName string, obj interface{}) *MetaDTO {
	m.RLock()
	if _, found := cacheMetaDTO[structName]; !found {
		m.RUnlock()
		m.Lock()
		defer m.Unlock()
		cacheMetaDTO[structName] = getNoneCacheMetaDTO(obj)
	} else {
		m.RUnlock()
	}

	return cacheMetaDTO[structName]
}

func GetDataForSelect(obj interface{}) ([]Column, []Alias, Join) {
	meta := GetMetaDTO(obj)
	return meta.ColsMap[ormUseInSelect], meta.Aliases, meta.Join
}

func GetDataForCreate(obj interface{}) ([]Column, []Argument) {
	cols, args := getMetaInfoUseInTag(obj, ormUseInCreate, emptyRootAlias)
	return cols, args
}

func GetDataForUpdate(obj interface{}) map[Column]Argument {
	cols, args := getMetaInfoUseInTag(obj, ormUseInUpdate, emptyRootAlias)

	cv := make(map[Column]Argument, len(cols))
	for i, v := range cols {
		cv[v] = args[i]
	}
	return cv
}

func GetTableName(obj interface{}) string {
	meta := GetMetaDTO(obj)
	return meta.TableName
}

func getNoneCacheMetaDTO(obj interface{}) *MetaDTO {
	meta := &MetaDTO{
		ColsMap:    map[ormUseInTagValue][]Column{ormUseInSelect: {}, ormUseInCreate: {}, ormUseInUpdate: {}},
		Aliases:    []Alias{},
		Join:       "",
		TableName:  "",
		StructName: getObjTypeNameByReflect(obj),
	}
	if obj == nil {
		return meta
	}

	meta.Join = getMetaInfoForOrmTagOnlyOne(tagOrmJoin, obj)

	meta.TableName = getMetaInfoForOrmTagOnlyOne(tagOrmTableName, obj)

	meta.Aliases = getMetaInfoForOrmAliasTag(obj)

	for _, v := range []string{ormUseInSelect, ormUseInCreate, ormUseInUpdate} {
		meta.ColsMap[v], _ = getMetaInfoUseInTag(obj, v, emptyRootAlias)
	}

	return meta
}

func getObjTypeNameByReflect(obj interface{}) string {
	if obj == nil {
		return "nil"
	}
	return reflect.Indirect(reflect.ValueOf(obj)).Type().Name()
}

func getMetaInfoForOrmTagOnlyOne(value ormUseInTagValue, obj interface{}) string {
	v := reflect.Indirect(reflect.ValueOf(obj))
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return ""
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if tagValue := field.Tag.Get(value); !isTagEmpty(tagValue) {
			return tagValue // we search first usage of tag, for high root component only,
		}
	}

	return ""
}

func isTagEmpty(tag string) bool {
	return tag == "" || tag == "-"
}

func getMetaInfoForOrmAliasTag(obj interface{}) []Alias {
	aliases := []Alias{}

	v := reflect.Indirect(reflect.ValueOf(obj))
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return aliases
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if v.Field(i).Kind() == reflect.Struct {
			aliasTagValue := field.Tag.Get(tagOrmAlias)
			if !isTagEmpty(aliasTagValue) {
				aliases = append(aliases, aliasTagValue)
			}
		}
	}

	return aliases
}

func getMetaInfoUseInTag(obj interface{}, useInTag ormUseInTagValue, alias Alias) ([]Column, []Argument) {
	cols, args := []Column{}, []Argument{}

	if obj == nil {
		return cols, args
	}

	v := reflect.Indirect(reflect.ValueOf(obj))
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return cols, args
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if tagValue := field.Tag.Get(tagOrmUseIN); !isTagEmpty(tagValue) {
			if !strings.Contains(tagValue, useInTag) {
				continue
			}

			dbTagValue := field.Tag.Get(tagDB)
			if isTagEmpty(dbTagValue) {
				continue
			}

			colValue := dbTagValue
			if alias != "" && useInTag == ormUseInSelect {
				colValue = fmt.Sprintf("%s.%s", alias, dbTagValue)
				colValue = fmt.Sprintf("%s as \"%s\"", colValue, colValue)
			}

			cols, args = append(cols, colValue), append(args, v.Field(i).Interface())
			continue
		}

		if v.Field(i).Kind() == reflect.Struct {
			if aliasTagValue := field.Tag.Get(tagOrmAlias); !isTagEmpty(aliasTagValue) {
				alias = field.Tag.Get(tagOrmAlias)
			}

			c, a := []Column{}, []Argument{}
			if v.Field(i).CanAddr() && v.Field(i).Addr().CanInterface() {
				c, a = getMetaInfoUseInTag(v.Field(i).Addr().Interface(), useInTag, alias)
			}
			cols, args = append(cols, c...), append(args, a...)
		}

	}

	return cols, args
}
