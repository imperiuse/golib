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
	JoinCond = string
	Alias    = string
	Argument = interface{}

	structName       = string
	ormUseInTagValue = string

	MetaDTO = struct {
		ColsMap    map[ormUseInTagValue][]Column
		JoinCond   JoinCond
		TableName  Table
		TableAlias Alias
		StructName Typ
	}
)

const (
	Undefined = ""
)

var (
	m            sync.RWMutex
	cacheMetaDTO = map[structName]*MetaDTO{}
)

// custom tag for "sugar" columns values prepare for Update squirrel library staff
const (
	tagDB       = "db" // must be for all DTO structs fields, this tag also used by sqlx
	tagOrmUseIN = "orm_use_in"

	underscored     = "_" // special name fo field contains tag orn_tab_name, orm_alias, orm_join
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

func GetDataForSelectOnlyCols(obj interface{}) []Column {
	meta := GetMetaDTO(obj)
	return meta.ColsMap[ormUseInSelect]
}

func GetDataForSelect(obj interface{}) ([]Column, JoinCond) {
	meta := GetMetaDTO(obj)
	return meta.ColsMap[ormUseInSelect], meta.JoinCond
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

// GetTableName - return table name
func GetTableName(obj interface{}) Table {
	meta := GetMetaDTO(obj)
	return meta.TableName
}

// GetTableAlias - return alias of table, if not set, return table name
func GetTableAlias(obj interface{}) Alias {
	meta := GetMetaDTO(obj)
	if meta.TableAlias == "" {
		return meta.TableName
	}

	return meta.TableAlias
}

// GetTableNameWithAlias - return string: `table_name` as `alias_name`
//(if alias_name not set, return `table_name` as `table_name`)
func GetTableNameWithAlias(obj interface{}) string {
	meta := GetMetaDTO(obj)
	if meta.TableName == "" {
		return ""
	}

	alias := meta.TableAlias
	if meta.TableAlias == "" {
		alias = meta.TableName
	}

	return fmt.Sprintf(" %s as %s ", meta.TableName, alias)
}

func getNoneCacheMetaDTO(obj interface{}) *MetaDTO {
	meta := &MetaDTO{
		ColsMap:    map[ormUseInTagValue][]Column{ormUseInSelect: {}, ormUseInCreate: {}, ormUseInUpdate: {}},
		JoinCond:   Undefined,
		TableName:  Undefined,
		TableAlias: Undefined,
		StructName: getObjTypeNameByReflect(obj),
	}
	if obj == nil {
		return meta
	}

	meta.JoinCond = getMetaInfoForOrmTagOnlyOne(tagOrmJoin, obj)

	meta.TableName = getMetaInfoForOrmTagOnlyOne(tagOrmTableName, obj)

	meta.TableAlias = getMetaInfoForOrmTagOnlyOne(tagOrmAlias, obj)

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

		if tagValue := field.Tag.Get(value); !isTagEmpty(tagValue) && field.Name == underscored {
			return tagValue // we search first usage of tag, for high root component only,
		}
	}

	return ""
}

func isTagEmpty(tag string) bool {
	return tag == "" || tag == "-"
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
