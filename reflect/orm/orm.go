package reflect

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// !DISCLAIMER!
// I know, that use reflect in most cases bad and slow.
// I know, that another ORM it is not go-way.
// But for me, using this small library for create micro ORM good choise in many case
// if you do not think about serious optimization
//
// This micro library use the same approach like sqlx with `db` tag parsing

type (
	Column   = string
	Join     = string
	Alias    = string
	Argument = interface{}

	structName = string
	ormUseType = string

	metaDTO = struct {
		colsMap map[ormUseType][]Column
		aliases []Alias
		join    Join
	}
)

var (
	cacheMetaDTO = map[structName]*metaDTO{}
)

var (
	ErrNotFoundInCache = errors.New("not found in cacheMetaDTO, check InitCacheForOrmMetaInfo, if all DTO register? ")
)

// custom tag for "sugar" columns values prepare for Update squirrel library staff
const (
	tagDB       = "db" // must be for all DTO structs fields, this tag also used by sqlx
	tagOrmUseIN = "orm_use_in"
	tagOrmAlias = "orm_alias"
	tagOrmJoin  = "orm_join"

	ormUseInSelect = "select"
	ormUseInCreate = "create"
	ormUseInUpdate = "update"
)

const emptyRootAlias = ""

func InitCacheForOrmMetaInfo(objs ...interface{}) error {
	for _, obj := range objs {
		meta := &metaDTO{
			colsMap: map[ormUseType][]Column{},
		}
		meta.join = getMetaInfoForOrmJoinTag(obj)
		meta.aliases = getMetaInfoForOrmAliasTag(obj)
		for _, v := range []string{ormUseInSelect, ormUseInCreate, ormUseInUpdate} {
			meta.colsMap[v], _ = getMetaInfoUseInTag(obj, v, emptyRootAlias)
		}
		cacheMetaDTO[getObjTypeNameByReflect(obj)] = meta
	}

	return nil
}

func getObjTypeNameByReflect(obj interface{}) string {
	return reflect.Indirect(reflect.ValueOf(obj)).Type().Name()
}

func getMetaInfoForOrmJoinTag(obj interface{}) Join {
	join := ""
	v := reflect.Indirect(reflect.ValueOf(obj))
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if tagValue := field.Tag.Get(tagOrmJoin); !isTagEmpty(tagValue) {
			join = tagValue
			continue // we search first join tag, for high root component only,
		}
	}
	return join
}

func isTagEmpty(tag string) bool {
	return tag == "" || tag == "-"
}

func getMetaInfoForOrmAliasTag(obj interface{}) []Alias {
	aliases := []Alias{}

	v := reflect.Indirect(reflect.ValueOf(obj))
	t := v.Type()

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

func getMetaInfoUseInTag(obj interface{}, useInTag ormUseType, alias Alias) ([]Column, []Argument) {
	cols, args := []Column{}, []Argument{}

	v := reflect.Indirect(reflect.ValueOf(obj))
	t := v.Type()

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

func GetOrmDataForSelect(obj interface{}) ([]Column, []Alias, Join, error) {
	meta, err := getMetaDTOInfo("", obj)
	if err != nil {
		return nil, nil, "", err
	}
	return meta.colsMap[ormUseInSelect], meta.aliases, meta.join, nil
}

func getMetaDTOInfo(typ string, obj interface{}) (*metaDTO, error) {
	// todo think about typ param
	if typ == "" {
		typ = getObjTypeNameByReflect(obj)
	}
	if m, found := cacheMetaDTO[typ]; found {
		return m, nil
	}
	return nil, ErrNotFoundInCache
}

func GetOrmDataForCreate(obj interface{}) ([]Column, []Argument, error) {
	cols, args := getMetaInfoUseInTag(obj, ormUseInCreate, emptyRootAlias)
	return cols, args, nil
}

func GetOrmDataForUpdate(obj interface{}) (map[Column]Argument, error) {
	cols, args := getMetaInfoUseInTag(obj, ormUseInUpdate, emptyRootAlias)

	cv := make(map[Column]Argument, len(cols))
	for i, v := range cols {
		cv[v] = args[i]
	}
	return cv, nil
}
