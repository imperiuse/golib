package orm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type (
	BaseDTO struct {
		ID        int64       `db:"id"          orm_use_in:"select"`
		CreatedAt time.Time   `db:"created_at"  orm_use_in:"select"`
		UpdatedAt time.Time   `db:"updated_at"  orm_use_in:"select,update"`
		_         interface{} `orm_use_in:"select,create,update"`
	}

	A struct {
		BaseDTO
		SelectOnly   string `db:"select_field"   orm_use_in:"select"`
		CreateOnly   string `db:"create_field"   orm_use_in:"create"`
		UpdateOnly   int    `db:"update_field"   orm_use_in:"update"`
		NoDbTagField string `orm_use_for:"update"`
		NoTagField   string
		_            interface{} `orm_table_name:"A" orm_alias:"a"`
	}

	B struct {
		BaseDTO
		CUS  float64     `db:"cus_field"   orm_use_in:"create,update,select"`
		CUS2 int         `db:"cus2_field"  orm_use_in:"create,update,select"`
		_    interface{} `orm_table_name:"B" orm_alias:"b"`
	}

	C struct {
		A `orm_alias:"a"`
		B `orm_alias:"b"`
		_ bool `orm_join:"ON a.id = b.id"`
	}

	D struct {
		BaseDTO
		CUS  float64     `db:"cus_field"   orm_use_in:"create,update,select"`
		CUS2 int         `db:"cus2_field"  orm_use_in:"create,update,select"`
		_    interface{} `orm_table_name:"D"`
	}

	BadStruct struct {
		*A
		_              struct{ a int }
		bad_name_field interface{} `orm_table_name:"B" orm_alias:"b"`
	}
)

const (
	SelectOnly = "SelectOnly"
	CreateOnly = "CreateOnly"
	UpdateOnly = 123
)

type OrmTestSuit struct {
	suite.Suite
}

func (suite *OrmTestSuit) SetupSuite() {
	InitMetaTagInfoCache([]interface{}{&A{}, &B{}, &C{}}...)
	InitMetaTagInfoCache(new(interface{}))
	InitMetaTagInfoCache(1, "", 1.23, &struct{}{})
	InitMetaTagInfoCache(nil)
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *OrmTestSuit) TearDownSuite() {
}

// The SetupTest method will be run before every test in the suite.
func (suite *OrmTestSuit) SetupTest() {

}

// The TearDownTest method will be run after every test in the suite.
func (suite *OrmTestSuit) TearDownTest() {
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSuite(t *testing.T) {
	suite.Run(t, new(OrmTestSuit))
}

func (suite *OrmTestSuit) Test_GetOrmDataForCreate() {
	t := suite.T()
	cols, args := GetDataForCreate(&A{
		SelectOnly: SelectOnly,
		CreateOnly: CreateOnly,
		UpdateOnly: UpdateOnly,
	})
	assert.NotNil(t, cols)
	assert.NotNil(t, args)
	assert.Equal(t, len(cols), len(args))
	assert.Equal(t, []string{"create_field"}, cols)
	assert.NotNil(t, []string{"CreateOnly"}, args)
}

func (suite *OrmTestSuit) Test_GetOrmDataForSelect() {
	t := suite.T()
	cols, join := GetDataForSelect(&C{})
	assert.Equal(t, "ON a.id = b.id", join)
	assert.Equal(t, []string{"a.id as \"a.id\"", "a.created_at as \"a.created_at\"", "a.updated_at as \"a.updated_at\"", "a.select_field as \"a.select_field\"", "b.id as \"b.id\"", "b.created_at as \"b.created_at\"", "b.updated_at as \"b.updated_at\"", "b.cus_field as \"b.cus_field\"", "b.cus2_field as \"b.cus2_field\""}, cols)
}

func (suite *OrmTestSuit) Test_GetUpdateColumnsValues() {
	t := suite.T()
	cv := GetDataForUpdate(&A{
		SelectOnly: SelectOnly,
		CreateOnly: CreateOnly,
		UpdateOnly: UpdateOnly,
	})
	assert.NotNil(t, cv)
	assert.Equal(t, 2, len(cv))
	assert.Equal(t, map[string]interface{}{"update_field": 123, "updated_at": time.Time{}}, cv)
}

func (suite *OrmTestSuit) Test_BadGetOrmDataForCreate() {
	t := suite.T()

	col, args := GetDataForCreate(&BadStruct{})
	assert.Equal(t, []string{}, col)
	assert.Equal(t, []interface{}{}, args)

	col, args = GetDataForCreate(nil)
	assert.Equal(t, []string{}, col)
	assert.Equal(t, []interface{}{}, args)

	col, args = GetDataForCreate(new(interface{}))
	assert.Equal(t, []string{}, col)
	assert.Equal(t, []interface{}{}, args)

	col, args = GetDataForCreate("123")
	assert.Equal(t, []string{}, col)
	assert.Equal(t, []interface{}{}, args)

	col, args = GetDataForCreate(123)
	assert.Equal(t, []string{}, col)
	assert.Equal(t, []interface{}{}, args)
}

func (suite *OrmTestSuit) Test_BadGetOrmDataForSelect() {
	t := suite.T()
	col, join := GetDataForSelect(&BadStruct{})
	assert.Equal(t, []string{}, col)
	assert.Equal(t, "", join)

	col, join = GetDataForSelect(nil)
	assert.Equal(t, []string{}, col)
	assert.Equal(t, "", join)

	col, join = GetDataForSelect(new(interface{}))
	assert.Equal(t, []string{}, col)
	assert.Equal(t, "", join)

	col, join = GetDataForSelect(1234)
	assert.Equal(t, []string{}, col)
	assert.Equal(t, "", join)

	col, join = GetDataForSelect("")
	assert.Equal(t, []string{}, col)
	assert.Equal(t, "", join)

}

func (suite *OrmTestSuit) Test_GetTableName() {
	t := suite.T()

	assert.Equal(t, "A", GetTableName(&A{}))
	assert.Equal(t, "B", GetTableName(&B{}))
	assert.Equal(t, "D", GetTableName(&D{}))
	assert.Equal(t, "", GetTableName(&C{}))
	assert.Equal(t, "", GetTableName(nil))
	assert.Equal(t, "", GetTableName(&BadStruct{}))
}

func (suite *OrmTestSuit) Test_GetTableAlias() {
	t := suite.T()

	assert.Equal(t, "a", GetTableAlias(&A{}))
	assert.Equal(t, "b", GetTableAlias(&B{}))
	assert.Equal(t, "D", GetTableAlias(&D{}))
	assert.Equal(t, "", GetTableAlias(&C{}))
	assert.Equal(t, "", GetTableAlias(nil))
	assert.Equal(t, "", GetTableAlias(&BadStruct{}))
}

func (suite *OrmTestSuit) Test_GetTableNameWithAlias() {
	t := suite.T()

	_ = BadStruct{}.bad_name_field

	assert.Equal(t, " A as a ", GetTableNameWithAlias(&A{}))
	assert.Equal(t, " B as b ", GetTableNameWithAlias(&B{}))
	assert.Equal(t, " D as D ", GetTableNameWithAlias(&D{}))
	assert.Equal(t, "", GetTableNameWithAlias(&C{}))
	assert.Equal(t, "", GetTableNameWithAlias(nil))
	assert.Equal(t, "", GetTableNameWithAlias(&BadStruct{}))
}
