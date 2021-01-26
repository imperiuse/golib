package reflect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type (
	BaseDTO struct {
		ID        int64     `db:"id"          orm_use_in:"select"`
		CreatedAt time.Time `db:"created_at"  orm_use_in:"select"`
		UpdatedAt time.Time `db:"updated_at"  orm_use_in:"select,update"`
	}

	A struct {
		BaseDTO
		SelectOnly   string `db:"select_field"   orm_use_in:"select"`
		CreateOnly   string `db:"create_field"   orm_use_in:"create"`
		UpdateOnly   int    `db:"update_field"   orm_use_in:"update"`
		NoDbTagField string `orm_use_for:"update"`
		NoTagField   string
	}

	B struct {
		BaseDTO
		CUS  float64 `db:"cus_field"   orm_use_in:"create,update,select"`
		CUS2 int     `db:"cus2_field"  orm_use_in:"create,update,select"`
	}

	C struct {
		A `orm_alias:"a"`
		B `orm_alias:"b"`
		_ string `orm_join:"ON a.id = b.id"`
	}

	BadStruct struct {
		*A
		_ struct{ a int }
	}
)

const (
	SelectOnly = "SelectOnly"
	CreateOnly = "CreateOnly"
	UpdateOnly = 123
)

type ReflectTestSuit struct {
	suite.Suite
}

func (suite *ReflectTestSuit) SetupSuite() {
	assert.Nil(suite.T(), InitCacheForOrmMetaInfo([]interface{}{&A{}, &B{}, &C{}}...))
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *ReflectTestSuit) TearDownSuite() {
}

// The SetupTest method will be run before every test in the suite.
func (suite *ReflectTestSuit) SetupTest() {

}

// The TearDownTest method will be run after every test in the suite.
func (suite *ReflectTestSuit) TearDownTest() {
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSuite(t *testing.T) {
	suite.Run(t, new(ReflectTestSuit))
}

func (suite *ReflectTestSuit) Test_GetOrmDataForCreate() {
	t := suite.T()
	cols, args, err := GetOrmDataForCreate(&A{
		SelectOnly: SelectOnly,
		CreateOnly: CreateOnly,
		UpdateOnly: UpdateOnly,
	})
	assert.Nil(t, err)
	assert.NotNil(t, cols)
	assert.NotNil(t, args)
	assert.Equal(t, len(cols), len(args))
	assert.Equal(t, []string{"create_field"}, cols)
	assert.NotNil(t, []string{"CreateOnly"}, args)
}

func (suite *ReflectTestSuit) Test_GetOrmDataForSelect() {
	t := suite.T()
	cols, aliases, join, err := GetOrmDataForSelect(&C{})
	assert.Nil(t, err)
	assert.NotNil(t, aliases)
	assert.Equal(t, []string{"a", "b"}, aliases)
	assert.Equal(t, "ON a.id = b.id", join)
	assert.Equal(t, []string{"a.id as \"a.id\"", "a.created_at as \"a.created_at\"", "a.updated_at as \"a.updated_at\"", "a.select_field as \"a.select_field\"", "b.id as \"b.id\"", "b.created_at as \"b.created_at\"", "b.updated_at as \"b.updated_at\"", "b.cus_field as \"b.cus_field\"", "b.cus2_field as \"b.cus2_field\""}, cols)
}

func (suite *ReflectTestSuit) Test_GetUpdateColumnsValues() {
	t := suite.T()
	cv, err := GetOrmDataForUpdate(&A{
		SelectOnly: SelectOnly,
		CreateOnly: CreateOnly,
		UpdateOnly: UpdateOnly,
	})
	assert.Nil(t, err)
	assert.NotNil(t, cv)
	assert.Equal(t, 2, len(cv))
	assert.Equal(t, map[string]interface{}{"update_field": 123, "updated_at": time.Time{}}, cv)
}

func (suite *ReflectTestSuit) Test_BadGetOrmDataForCreate() {
	t := suite.T()
	col, args, err := GetOrmDataForCreate(&BadStruct{})
	assert.Nil(t, err)
	assert.Equal(t, []string{}, col)
	assert.Equal(t, []interface{}{}, args)
}

func (suite *ReflectTestSuit) Test_BadGetOrmDataForSelect() {
	t := suite.T()
	col, args, join, err := GetOrmDataForSelect(&BadStruct{})
	assert.NotNil(t, err)
	assert.Nil(t, col)
	assert.Nil(t, args)
	assert.Equal(t, "", join)
	assert.Equal(t, ErrNotFoundInCache, err)
}
