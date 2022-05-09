package intergation

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/imperiuse/golib/db/genrepo/emptygen"

	"math/rand"
	"testing"

	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v4"        // for pgx driver import.
	_ "github.com/jackc/pgx/v4/stdlib" // for pgx driver import.

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/imperiuse/golib/db"
	"github.com/imperiuse/golib/db/connector"
	"github.com/imperiuse/golib/db/example/simple/config"
	"github.com/imperiuse/golib/db/example/simple/dto"
	"github.com/imperiuse/golib/db/repo"
	"github.com/imperiuse/golib/db/repo/empty"
	"github.com/imperiuse/golib/reflect/orm"
)

const (
	PostgresUser     = "test"
	PostgresPassword = "test"
	PostgresDB       = "test"
	PostgresHost     = "localhost"
	PostgresPort     = "5433"
)

type RepositoryTestSuit struct {
	suite.Suite
	ctx       context.Context
	ctxCancel context.CancelFunc
	logger    db.Logger

	db                              *sqlx.DB
	connector                       db.Connector[config.SimpleTestConfig]
	connectorWithValidation         db.Connector[config.SimpleTestConfig]
	connectorWithValidationAndCache db.Connector[config.SimpleTestConfig]
}

var DTOs = []db.DTO{&dto.User[dto.ID]{}, &dto.Role[dto.ID]{}, &dto.Paginator[dto.ID]{}}

func GetTableNames(dtos []db.DTO) []db.Table {
	names := make([]db.Table, 0, len(dtos))
	for _, v := range dtos {
		names = append(names, v.Repo())
	}

	return names
}

// The SetupSuite method will be run by testify once, at the very
// start of the testing suite, before any tests are run.
func (suite *RepositoryTestSuit) SetupSuite() {
	suite.ctx, suite.ctxCancel = context.WithCancel(context.Background())
	suite.logger = zap.NewNop()

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		PostgresUser,
		PostgresPassword,
		PostgresHost,
		PostgresPort,
		PostgresDB,
	)

	dbConn, err := sqlx.Connect("pgx", dsn)
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), dbConn)

	suite.db = dbConn

	a := make([]any, 0, len(DTOs))
	for _, v := range DTOs {
		a = append(a, v)
	}

	orm.InitMetaTagInfoCache(a...)
	orm.InitMetaTagInfoCache(&dto.BaseDTO[dto.ID]{}, &dto.UsersRole[dto.ID]{})

	tables := []string{}
	// create table
	for _, obj := range DTOs {
		table := orm.GetTableName(obj)
		assert.NotEqual(suite.T(), "", table)
		tables = append(tables, table)

		_, err = dbConn.ExecContext(suite.ctx, dto.DSL[table])
		assert.Nil(suite.T(), err)
	}

	// Refresh DB
	for _, table := range tables {
		_, err = dbConn.ExecContext(suite.ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", table))
		assert.Nil(suite.T(), err)
	}

	suite.connector = connector.New[config.SimpleTestConfig](config.New(
		squirrel.Dollar, false, false), zap.NewNop(), dbConn)
	suite.connectorWithValidation = connector.New[config.SimpleTestConfig](
		config.New(squirrel.Dollar, true, false), zap.NewNop(), dbConn)
	suite.connectorWithValidationAndCache = connector.New[config.SimpleTestConfig](
		config.New(squirrel.Dollar, true, true), zap.NewNop(), dbConn)

	assert.NotNil(suite.T(), suite.connector)
	assert.NotNil(suite.T(), suite.connectorWithValidation)
	assert.NotNil(suite.T(), suite.connectorWithValidationAndCache)

	suite.connector.AddAllowsRepos(GetTableNames(DTOs)...) // not needed only for test not error or panic here ...
	suite.connectorWithValidation.AddAllowsRepos(GetTableNames(DTOs)...)
	suite.connectorWithValidationAndCache.AddAllowsRepos(GetTableNames(DTOs)...)

	assert.Nil(suite.T(), appendTestDataToTables(suite))
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *RepositoryTestSuit) TearDownSuite() {
	for _, obj := range DTOs {
		_, err := suite.connector.Repo(obj).Delete(suite.ctx, 1)
		assert.Nil(suite.T(), err)
	}

	assert.Nil(suite.T(), suite.db.Close())
}

// The SetupTest method will be run before every test in the suite.
func (suite *RepositoryTestSuit) SetupTest() {
	suite.ctx, suite.ctxCancel = context.WithCancel(context.Background())
}

// The TearDownTest method will be run after every test in the suite.
func (suite *RepositoryTestSuit) TearDownTest() {
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSuite(t *testing.T) {
	initTables() // prevent errors when run first containers in github CI
	suite.Run(t, new(RepositoryTestSuit))
}

func initTables() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		PostgresUser,
		PostgresPassword,
		PostgresHost,
		PostgresPort,
		PostgresDB,
	)

	dbConn, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		panic(err)
	}

	tables := []string{}
	// create table
	for _, obj := range DTOs {
		table := orm.GetTableName(obj)
		tables = append(tables, table)

		_, _ = dbConn.ExecContext(context.Background(), dto.DSL[table])
	}

	// Refresh DB
	for _, table := range tables {
		_, _ = dbConn.ExecContext(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", table))
	}

}

func appendTestDataToTables(s *RepositoryTestSuit) error {
	const cntRecords = 200

	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	randStringRunes := func(n int) string {
		b := make([]rune, n)
		for i := range b {
			b[i] = letterRunes[rand.Intn(len(letterRunes))]
		}
		return string(b)
	}

	for i := 0; i < cntRecords; i++ {
		_, err := s.connector.AutoCreate(s.ctx, &dto.Paginator[dto.ID]{
			Name: randStringRunes(10),
		})

		if err != nil {
			return err
		}
	}

	return nil
}

type notRegisterDTO[I db.ID] struct{}

func (n notRegisterDTO[I]) Identity() db.ID { return 0 }
func (n notRegisterDTO[I]) ID() I           { return *new(I) }
func (n notRegisterDTO[I]) Repo() db.Table  { return "not_register_dto" }

func (suite *RepositoryTestSuit) Test_Connector_Repo_Creation() {
	t := suite.T()

	u := dto.User[dto.ID]{}

	for _, c := range []db.Connector[config.SimpleTestConfig]{
		suite.connector,
		suite.connectorWithValidation,
		suite.connectorWithValidationAndCache,
	} {
		// test _system methods
		assert.NotNil(t, c.Connection())
		assert.NotNil(t, c.Logger())
		assert.NotNil(t, c.Config())
		assert.Equal(t, suite.db, c.Connection())

		r := c.Repo(u)
		assert.NotNil(t, r)
		assert.Equal(t, dto.User[dto.ID]{}.Repo(), r.Name())

		genericRepo := repo.NewGen[dto.ID, dto.User[dto.ID]](c)
		assert.NotNil(t, genericRepo)
		assert.Equal(t, dto.User[dto.ID]{}.Repo(), genericRepo.Name())
	}

	// without validation, we should create new repo
	r := suite.connector.Repo(notRegisterDTO[dto.ID]{})
	assert.NotNil(t, r)
	assert.Equal(t, notRegisterDTO[dto.ID]{}.Repo(), r.Name())

	genericRepo := repo.NewGen[dto.ID, notRegisterDTO[dto.ID]](suite.connector)
	assert.NotNil(t, genericRepo)
	assert.Equal(t, notRegisterDTO[dto.ID]{}.Repo(), genericRepo.Name())

	// without validation, we should NOT create new repo // we must return empty repo
	r = suite.connectorWithValidation.Repo(notRegisterDTO[dto.ID]{})
	assert.NotNil(t, r)
	assert.Equal(t, empty.Repo, r)

	genericRepo = repo.NewGen[dto.ID, notRegisterDTO[dto.ID]](suite.connectorWithValidation)
	assert.NotNil(t, genericRepo)
	assert.Equal(t, emptygen.NewGen[dto.ID, notRegisterDTO[dto.ID]](), genericRepo)

	r = suite.connectorWithValidationAndCache.Repo(notRegisterDTO[dto.ID]{})
	assert.NotNil(t, r)
	assert.Equal(t, empty.Repo, r)

	genericRepo = repo.NewGen[int, notRegisterDTO[dto.ID]](suite.connectorWithValidationAndCache)
	assert.NotNil(t, genericRepo)
	assert.Equal(t, emptygen.NewGen[dto.ID, notRegisterDTO[dto.ID]](), genericRepo)
}

func (suite *RepositoryTestSuit) Test_Connector_AutoCRUD() {
	t := suite.T()

	r := dto.Role[dto.ID]{
		BaseDTO: dto.BaseDTO[dto.ID]{Id: 1},
		Name:    "User",
		Rights:  1,
	}

	u := dto.User[dto.ID]{
		BaseDTO:  dto.BaseDTO[dto.ID]{Id: 1},
		Name:     "User1",
		Email:    "user@mail.com",
		Password: "p@ssw0rd",
		RoleID:   1,
	}

	for i, c := range []db.Connector[config.SimpleTestConfig]{
		suite.connector,
		suite.connectorWithValidation,
		suite.connectorWithValidationAndCache,
	} {
		nr := notRegisterDTO[dto.ID]{}
		// For second two connectors check Validation  (Could not create Repo)
		if i > 0 {
			// check we returned empty repo for all
			assert.Equal(t, empty.Repo, c.Repo(nr))

			// 		a) repo way
			_, err := c.AutoCreate(suite.ctx, nr)
			assert.Equal(t, db.ErrInvalidRepoEmptyRepo, err)

			// 		b) generic repo way
			_, err = repo.NewGen[dto.ID, notRegisterDTO[dto.ID]](c).Create(suite.ctx, nr)
			assert.Equal(t, db.ErrInvalidRepoEmptyRepo, err)
		}

		// I. Part
		// 	1) Could not create User (foreign constraint)
		// 		a) repo way
		{
			_, err := c.AutoCreate(suite.ctx, u)
			assert.NotNil(t, err) // actual  : *fmt.wrapError(&fmt.wrapError{msg:"err while Rollback. error: <nil>, ERROR: insert or update on table \"users\" violates foreign key constraint \"fkey__r\" (SQLSTATE 23503)", err:(*pgconn.PgError)(0xc0002602d0)})
		}
		// 		b) generic repo way
		{
			_, err := repo.NewGen[dto.ID, dto.User[dto.ID]](c).Create(suite.ctx, u)
			assert.NotNil(t, err) // actual  : *fmt.wrapError(&fmt.wrapError{msg:"err while Rollback. error: <nil>, ERROR: insert or update on table \"users\" violates foreign key constraint \"fkey__r\" (SQLSTATE 23503)", err:(*pgconn.PgError)(0xc0002602d0)})
		}
		// 		c) old school (pure connection) way
		{
			res, err := c.Connection().ExecContext(
				suite.ctx, fmt.Sprintf("INSERT INTO %s (name) VAlUES($1);", u.Repo()), u.Name)
			assert.Nil(t, res)
			assert.NotNil(t, err) // actual  : *fmt.wrapError(&fmt.wrapError{msg:"err while Rollback. error: <nil>, ERROR: insert or update on table \"users\" violates foreign key constraint \"fkey__r\" (SQLSTATE 23503)", err:(*pgconn.PgError)(0xc0002602d0)})
		}
		// 	2) Check that we could not Get User (User not exists)
		// 		a) repo way
		u2 := dto.User[dto.ID]{BaseDTO: dto.BaseDTO[dto.ID]{Id: u.ID()}}
		err := c.AutoGet(suite.ctx, &u2)
		assert.Equal(t, "", u2.Name)
		assert.NotEqual(t, u, u2)
		assert.Equal(t, sql.ErrNoRows, err)

		r2 := dto.Role[dto.ID]{BaseDTO: dto.BaseDTO[dto.ID]{Id: u.ID()}}
		err = c.AutoGet(suite.ctx, &r2)
		assert.Equal(t, "", r2.Name)
		assert.NotEqual(t, r, r2)
		assert.Equal(t, sql.ErrNoRows, err)

		// 		b) generic repo way
		u2, err = repo.NewGen[dto.ID, dto.User[dto.ID]](c).Get(suite.ctx, u.ID())
		assert.Equal(t, "", u2.Name)
		assert.NotEqual(t, u, u2)
		assert.Equal(t, sql.ErrNoRows, err)

		r2, err = repo.NewGen[dto.ID, dto.Role[dto.ID]](c).Get(suite.ctx, r.ID())
		assert.Equal(t, "", r2.Name)
		assert.NotEqual(t, r, r2)
		assert.Equal(t, sql.ErrNoRows, err)

		// 		c) old school (pure connection) way
		res, err := c.Connection().QueryContext(
			suite.ctx, fmt.Sprintf("SELECT * FROM %s WHERE id=$1 LIMIT 1;", u.Repo()), u.ID())
		assert.NotNil(t, res)
		assert.Nil(t, err)

		res, err = c.Connection().QueryContext(
			suite.ctx, fmt.Sprintf("SELECT * FROM %s WHERE id=$1 LIMIT 1;", r.Repo()), r.ID())
		assert.NotNil(t, res)
		assert.Nil(t, err)

		// 	3) Check we could not update User (User not exist)
		// 		a) repo way
		// 		b) generic repo way
		// 		c) old school (pure connection) way

		// 	4) Check we delete without error User. (User not exist)
		// 		a) repo wa
		n, err := c.AutoDelete(suite.ctx, u)
		assert.Equal(t, int64(0), n)
		assert.Nil(t, err)

		n, err = c.AutoDelete(suite.ctx, r)
		assert.Equal(t, int64(0), n)
		assert.Nil(t, err)

		// 		b) generic repo way
		n, err = repo.NewGen[dto.ID, dto.User[dto.ID]](c).Delete(suite.ctx, u.Identity().(dto.ID))
		assert.Equal(t, int64(0), n)
		assert.Nil(t, err)

		n, err = c.AutoDelete(suite.ctx, r)
		assert.Equal(t, int64(0), n)
		assert.Nil(t, err)

		// 		c) old school (pure connection) way
		_, err = c.Connection().ExecContext(
			suite.ctx, fmt.Sprintf("DELETE FROM %s WHERE id=$1;", r.Repo()), r.ID())
		assert.Nil(t, err)

		_, err = c.Connection().ExecContext(
			suite.ctx, fmt.Sprintf("DELETE FROM %s WHERE id=$1;", r.Repo()), r.ID())
		assert.Nil(t, err)

		// II. Part
		// 	1) Create Role and User
		id, err := c.AutoCreate(suite.ctx, r)
		r.Id = dto.ID(id)
		assert.Nil(t, err)

		u.RoleID = r.Id
		id, err = c.AutoCreate(suite.ctx, u)
		u.Id = dto.ID(id)
		assert.Nil(t, err)

		// III. Part delete all records
		n, err = c.AutoDelete(suite.ctx, u)
		assert.Equal(t, int64(1), n)
		assert.Nil(t, err)

		n, err = c.AutoDelete(suite.ctx, r)
		assert.Equal(t, int64(1), n)
		assert.Nil(t, err)
	}
}

//func (suite *RepositoryTestSuit) Test_Repo_AutoRepo_SqlxDBConnectorI() {
//	t := suite.T()
//
//	for _, obj := range DTOs {
//		assert.NotNil(t, suite.repos.AutoReflectRepo(obj).PureConnector())
//		assert.Equal(t, suite.db, suite.repos.AutoReflectRepo(obj).PureConnector())
//		assert.NotEqual(t, suite.repos.AutoReflectRepo(obj), suite.repos.Repo("_UNKNOWN_"))
//		assert.Equal(t, emptyRepo, suite.repos.Repo("_UNKNOWN_"))
//		assert.Equal(t, emptyRepo, suite.repos.Repo("_UNKNOWN_2"))
//		assert.Equal(t, emptyRepo, suite.repos.AutoReflectRepo(&BaseDTO{}))
//		assert.Equal(t, emptyRepo, suite.repos.AutoReflectRepo(&NotDTO{}))
//	}
//
//	assert.Equal(t, suite.repos.Repo(orm.GetTableName(&User{})), suite.repos.AutoRepo(&User{}))
//	assert.NotNil(t, suite.repos.Repo(orm.GetTableName(&User{})))
//	assert.Equal(t, suite.repos.Repo(orm.GetTableName(&Role{})), suite.repos.AutoRepo(&Role{}))
//	assert.NotNil(t, suite.repos.Repo(orm.GetTableName(&Role{})))
//}
//
//func (suite *RepositoryTestSuit) Test_EmptyRepo_NotPanic() {
//	t := suite.T()
//
//	r := suite.repos.Repo("_UNKNOWN_")
//	assert.NotNil(t, r)
//	assert.Equal(t, r, emptyRepo)
//
//	assert.NotNil(t, r.PureConnector())
//	assert.NotEqual(t, suite.db, r.PureConnector())
//
//	emptyCon := r.PureConnector()
//
//	assert.Equal(t, FakeStringAns, emptyCon.DriverName())
//
//	assert.Equal(t, FakeStringAns, emptyCon.Rebind(""))
//
//	r1, r2, r3 := emptyCon.BindNamed("", nil)
//	assert.Equal(t, FakeStringAns, r1)
//	assert.Nil(t, r2)
//	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, errors.Cause(r3))
//
//	row, err := emptyCon.QueryxContext(suite.ctx, "select * from users;")
//	assert.NotNil(t, row)
//	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, errors.Cause(err))
//
//	row2, err := emptyCon.QueryContext(suite.ctx, "select * from users;")
//	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, errors.Cause(err))
//	assert.NotNil(t, row2)
//
//	assert.Equal(t, &sqlx.Row{}, emptyCon.QueryRowxContext(suite.ctx, "select * from users;"))
//
//	res, err := emptyCon.ExecContext(suite.ctx, "select * from users;")
//	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, errors.Cause(err))
//	assert.NotNil(t, res)
//
//	stmt, err := emptyCon.PrepareContext(suite.ctx, "select * from users;")
//	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, errors.Cause(err))
//	assert.NotNil(t, stmt)
//
//	tx, err := emptyCon.BeginTxx(suite.ctx, nil)
//	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, errors.Cause(err))
//	assert.NotNil(t, tx)
//
//	// CHECK REPO CRUD
//
//	var role Role
//	assert.Equal(t, sql.ErrNoRows, errors.Cause(r.Get(suite.ctx, 1, &role)))
//
//	id, err := r.Create(suite.ctx, &role)
//	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, errors.Cause(err))
//	assert.Equal(t, int64(0), id)
//
//	cnt, err := r.Update(suite.ctx, 1, &role)
//	assert.NotNil(t, err)
//	assert.Equal(t, int64(0), cnt)
//
//	var roles []Role
//	err = r.Select(suite.ctx, squirrel.Select("*").From(r.Name()).OrderBy("id DESC"), &roles)
//	assert.NotNil(t, err)
//	assert.Equal(t, 0, len(roles))
//
//	err = r.SelectWithCursorOnPKPagination(suite.ctx, squirrel.Select("*").From(r.Name()), CursorPaginationParams{
//		Limit:     10,
//		Cursor:    0,
//		DescOrder: false,
//	}, &roles)
//	assert.NotNil(t, err)
//	assert.Equal(t, 0, len(roles))
//
//	ppr, err := r.SelectWithPagePagination(suite.ctx, squirrel.Select("*").From(r.Name()), PagePaginationParams{PageNumber: 1, PageSize: 10}, &roles)
//	assert.NotNil(t, err)
//	assert.NotNil(t, ppr)
//	assert.Equal(t, 0, len(roles))
//
//	cnt, err = r.Delete(suite.ctx, 1)
//	assert.NotNil(t, err)
//	assert.Equal(t, int64(0), cnt)
//
//	var temp any
//	err = r.FindBy(suite.ctx, []Column{"*"}, squirrel.Eq{"id": 1}, &temp)
//	assert.NotNil(t, err)
//
//	err = r.FindOneBy(suite.ctx, []Column{"*"}, squirrel.Eq{"id": 1}, &temp)
//	assert.NotNil(t, err)
//
//	ucnt, err := r.CountByQuery(suite.ctx, squirrel.Select("*").From("unknown"))
//	assert.NotNil(t, err)
//	assert.Equal(t, uint64(0), ucnt)
//
//	err = r.FindByWithInnerJoin(suite.ctx, []Column{"*"}, "al as al", "ON al.id = p.id", squirrel.Eq{"id": 1}, &temp)
//	assert.NotNil(t, err)
//
//	err = r.FindOneByWithInnerJoin(suite.ctx, []Column{"*"}, "al as al", "ON al.id = p.id", squirrel.Eq{"id": 1}, &temp)
//	assert.NotNil(t, err)
//
//	rows1, err := r.GetRowsByQuery(suite.ctx, squirrel.Select("*").From("unknown"))
//	assert.NotNil(t, err)
//	assert.NotNil(t, rows1)
//
//	id, err = r.Insert(suite.ctx, []Column{"name"}, []any{"test"})
//	assert.NotNil(t, err)
//	assert.Equal(t, int64(0), id)
//
//	cnt, err = r.UpdateCustom(suite.ctx, map[string]any{}, squirrel.Eq{"id": 1})
//	assert.NotNil(t, err)
//	assert.Equal(t, int64(0), cnt)
//}
//
//func (suite *RepositoryTestSuit) Test_CRUD() {
//	t := suite.T()
//
//	assert.Equal(t, "Roles", suite.repos.AutoRepo(&Role{}).Name())
//
//	const (
//		rights       = 10
//		role1        = "role1"
//		updatedRole1 = "role1_updated"
//	)
//
//	var role = Role{
//		Name:   role1,
//		Rights: rights,
//	}
//
//	id, err := suite.repos.AutoRepo(&role).Create(suite.ctx, &role)
//	assert.Nil(t, err)
//	assert.NotNil(t, id)
//
//	// Alternative way
//	id2, err := suite.repos.AutoCreate(suite.ctx, &role)
//	assert.Nil(t, err)
//	assert.NotNil(t, id2)
//	// also you can directly get Repo Role  and exec method Create
//	assert.NotNil(t, suite.repos.Repo("Roles"))
//	//suite.repos.Repo("Roles").Create(.....)
//
//	var temp Role
//	err = suite.repos.AutoRepo(&temp).Get(suite.ctx, id, &temp)
//	assert.Nil(t, err)
//	assert.Equal(t, role.Name, temp.Name)
//	assert.Equal(t, role.Rights, temp.Rights)
//	assert.NotNil(t, temp.UpdatedAt)
//	assert.NotNil(t, temp.CreatedAt)
//	assert.NotNil(t, temp.ID)
//	assert.Equal(t, id, temp.ID)
//
//	// Alternative way to get obj from db
//	var temp2 = Role{
//		BaseDTO: BaseDTO{ID: id.(Integer)},
//	}
//	err = suite.repos.AutoGet(suite.ctx, &temp2)
//	assert.Nil(t, err)
//	assert.Equal(t, temp, temp2)
//
//	var temp3 = Role{
//		BaseDTO: BaseDTO{ID: id.(Integer)},
//	}
//	err = suite.repos.AutoReflectRepo(temp3).Get(suite.ctx, id, &temp3)
//	assert.Nil(t, err)
//	assert.Equal(t, temp, temp3)
//
//	var temp4 Role
//	err = suite.repos.Repo(orm.GetTableName(&temp3)).Get(suite.ctx, id, &temp4)
//	assert.Nil(t, err)
//	assert.Equal(t, temp, temp4)
//
//	role.Name = updatedRole1
//	cnt, err := suite.repos.AutoRepo(&role).Update(suite.ctx, id, &role)
//	assert.Nil(t, err)
//	assert.Equal(t, int64(1), cnt)
//	err = suite.repos.AutoRepo(&temp).Get(suite.ctx, id, &temp)
//	assert.Nil(t, err)
//	assert.Equal(t, role.Name, temp.Name)
//
//	cnt, err = suite.repos.AutoRepo(&role).Delete(suite.ctx, id)
//	assert.Nil(t, err)
//	assert.Equal(t, int64(1), cnt)
//
//	err = suite.repos.Repo("Roles").Get(suite.ctx, id2, &temp)
//	assert.Nil(t, err)
//	assert.NotNil(t, temp.ID)
//
//	temp.Name = ""
//	cnt, err = suite.repos.AutoUpdate(suite.ctx, &temp)
//	assert.Nil(t, err)
//	assert.Equal(t, int64(1), cnt)
//
//	temp2.ID = temp.ID
//	assert.Nil(t, suite.repos.AutoGet(suite.ctx, &temp2))
//	assert.Equal(t, temp, temp2)
//
//	cnt, err = suite.repos.AutoDelete(suite.ctx, &temp)
//	assert.Nil(t, err)
//	assert.Equal(t, int64(1), cnt)
//}
//
//func (suite *RepositoryTestSuit) Test_Advance_RepoFunc() {
//	t := suite.T()
//	ctx := suite.ctx
//
//	// Create one role, than create one user, than check Join, Cnt
//	var role = Role{
//		Name:   "new_test_role",
//		Rights: 100,
//	}
//	roleID, err := suite.repos.AutoCreate(ctx, &role)
//	defer func() { _, _ = suite.repos.AutoRepo(&role).Delete(ctx, roleID) }()
//	assert.Nil(t, err)
//	assert.NotNil(t, roleID)
//
//	const newName = "newTestRole"
//	// Check Update Custom
//	cnt, err := suite.repos.AutoRepo(&role).UpdateCustom(ctx, map[string]interface{}{"name": newName}, squirrel.Eq{"id": roleID})
//	assert.Nil(t, err)
//	assert.Equal(t, int64(1), cnt)
//
//	var roles []Role
//	assert.Nil(t, suite.repos.AutoRepo(&role).FindBy(ctx, []Column{"rights"}, squirrel.Eq{"id": roleID}, &roles))
//	assert.Equal(t, 1, len(roles))
//	assert.NotEqual(t, newName, roles[0].Name, "role name must not change, because we have not updated name yet ")
//
//	var roleOne Role
//	assert.Nil(t, suite.repos.AutoRepo(&role).FindOneBy(ctx, []Column{"rights"}, squirrel.Eq{"id": roleID}, &roleOne))
//	assert.NotEqual(t, newName, roleOne, "role name must not change, because we have not updated name yet ")
//
//	roles = roles[1:]
//	assert.Nil(t, suite.repos.AutoRepo(&role).FindBy(ctx, []Column{"name"}, squirrel.Eq{"id": roleID}, &roles))
//	assert.Equal(t, 1, len(roles))
//	assert.Equal(t, newName, roles[0].Name, "role name must change, because we have  updated name already ")
//
//	var user = User{
//		Name:     "testUser",
//		Email:    "test@test.ru",
//		Password: "123456",
//		RoleID:   roleID.(int64),
//	}
//
//	userID, err := suite.repos.AutoCreate(ctx, &user)
//	defer func() { _, _ = suite.repos.AutoRepo(&user).Delete(ctx, userID) }()
//	assert.Nil(t, err)
//	assert.NotNil(t, userID)
//
//	cols, joinCond := orm.GetDataForSelect(&UsersRole{})
//	nameWithAlias := orm.GetTableNameWithAlias(&user)
//	joinCond = orm.GetTableNameWithAlias(&Role{}) + " " + joinCond
//	var ur UsersRole
//	err = suite.repos.AutoRepo(&user).FindOneByWithInnerJoin(ctx, cols, nameWithAlias, joinCond, squirrel.Eq{"u.id": userID}, &ur)
//	assert.Nil(t, err)
//	assert.NotNil(t, ur)
//	assert.Equal(t, userID, ur.User.ID)
//	assert.Equal(t, user.Name, ur.User.Name)
//	assert.Equal(t, user.Email, ur.User.Email)
//	assert.Equal(t, user.Password, ur.User.Password)
//	assert.Equal(t, user.RoleID, ur.Role.ID)
//	assert.Equal(t, role.Rights, ur.Role.Rights)
//
//	// Insert one more user with same role
//
//	var role2 = Role{
//		Name:   "role2",
//		Rights: 50,
//	}
//	cols, args := orm.GetDataForCreate(role2)
//	role2ID, err := suite.repos.AutoRepo(&role2).Insert(ctx, cols, args)
//	assert.Nil(t, err)
//	assert.NotNil(t, role2ID)
//	defer func() { _, _ = suite.repos.AutoRepo(&role).Delete(ctx, role2ID) }()
//
//	{
//		cnt, err := suite.repos.AutoRepo(&role2).CountByQuery(ctx, squirrel.Select("count(1)").From(orm.GetTableName(&role2)))
//		assert.Nil(t, err)
//		assert.Equal(t, uint64(2), cnt)
//
//		_, err = suite.repos.AutoRepo(&role2).CountByQuery(ctx, squirrel.Select("*").From(orm.GetTableName(&role2)))
//		assert.NotNil(t, err)
//
//		_, err = suite.repos.AutoRepo(&role2).CountByQuery(ctx, squirrel.Select("*"))
//		assert.NotNil(t, err)
//	}
//
//	rows, err := suite.repos.AutoRepo(&role2).GetRowsByQuery(ctx, squirrel.Select("*").From(orm.GetTableName(&role2)))
//	assert.Nil(t, err)
//	assert.NotNil(t, rows)
//
//	roles = []Role{}
//	repo := suite.repos.AutoRepo(&role2)
//
//	err = repo.Select(suite.ctx, squirrel.Select("*").From(repo.Name()).OrderBy("id DESC"), &roles)
//	assert.Nil(t, err)
//	assert.Equal(t, 2, len(roles))
//
//	roles = []Role{}
//	err = repo.SelectWithCursorOnPKPagination(suite.ctx, squirrel.Select("*").From(repo.Name()), CursorPaginationParams{
//		Limit:     10,
//		Cursor:    0,
//		DescOrder: false,
//	}, &roles)
//	assert.Nil(t, err)
//	assert.Equal(t, 2, len(roles))
//
//	roles = []Role{}
//	ppr, err := repo.SelectWithPagePagination(suite.ctx, squirrel.Select("*").From(repo.Name()), PagePaginationParams{PageNumber: 1, PageSize: 10}, &roles)
//	assert.Nil(t, err)
//	assert.NotNil(t, ppr)
//	assert.Equal(t, 2, len(roles))
//}
//
//func (suite *RepositoryTestSuit) Test_GetAllPossibleErrors() {
//	t := suite.T()
//	ctx := suite.ctx
//
//	_, err := suite.repos.AutoRepo(&Role{}).CountByQuery(ctx, squirrel.Select())
//	assert.NotNil(t, err)
//
//	_, err = suite.repos.AutoRepo(&Role{}).GetRowsByQuery(ctx, squirrel.Select())
//	assert.NotNil(t, err)
//
//	{
//		err = suite.repos.AutoRepo(&User{}).FindByWithInnerJoin(ctx, []string{}, "", "", squirrel.Eq{}, nil)
//		assert.NotNil(t, err)
//	}
//
//	assert.NotNil(t, suite.repos.AutoRepo(&User{}).FindBy(ctx, []string{}, squirrel.Eq{}, nil))
//
//	assert.NotNil(t, suite.repos.AutoRepo(&User{}).FindOneBy(ctx, []string{}, squirrel.Eq{}, nil))
//
//	{
//		id, err := suite.repos.AutoRepo(&User{}).Insert(ctx, []string{}, []interface{}{})
//		assert.NotNil(t, err)
//		assert.Equal(t, int64(0), id)
//	}
//
//	{
//		id, err := suite.repos.AutoRepo(&User{}).Create(ctx, new(interface{}))
//		assert.NotNil(t, err)
//		assert.Equal(t, int64(0), id)
//
//		id, err = suite.repos.AutoRepo(&User{}).Create(ctx, nil)
//		assert.NotNil(t, err)
//		assert.Equal(t, int64(0), id)
//
//		id, err = suite.repos.AutoRepo(&User{}).Create(ctx, &NotDTO{})
//		assert.NotNil(t, err)
//		assert.Equal(t, int64(0), id)
//	}
//
//}

//func (suite *RepositoryTestSuit) Test_SelectWithPagePagination() {
//	t := suite.T()
//
//	type Test struct {
//		Params  db.PagePaginationParams
//		Results db.PagePaginationResults
//		LenData int
//	}
//
//	testTable := []Test{{
//		Params: db.PagePaginationParams{
//			PageNumber: 0,
//			PageSize:   50,
//		},
//		Results: db.PagePaginationResults{
//			CurrentPageNumber: 0,
//			NextPageNumber:    0,
//			CntPages:          4,
//		},
//		LenData: 50,
//	},
//		{
//			Params: db.PagePaginationParams{
//				PageNumber: 1,
//				PageSize:   49,
//			},
//			Results: db.PagePaginationResults{
//				CurrentPageNumber: 1,
//				NextPageNumber:    0,
//				CntPages:          5,
//			},
//			LenData: 49,
//		},
//
//		{
//			Params: db.PagePaginationParams{
//				PageNumber: 2,
//				PageSize:   50,
//			},
//			Results: db.PagePaginationResults{
//				CurrentPageNumber: 2,
//				NextPageNumber:    0,
//				CntPages:          4,
//			},
//			LenData: 50,
//		},
//
//		{
//			Params: db.PagePaginationParams{
//				PageNumber: 5,
//				PageSize:   50,
//			},
//			Results: db.PagePaginationResults{
//				CurrentPageNumber: 5,
//				NextPageNumber:    0,
//				CntPages:          4,
//			},
//			LenData: 0,
//		},
//	}
//
//	cols, _ := orm.GetDataForSelect(&Paginator{})
//	table := orm.GetTableName(&Paginator{})
//
//	for _, test := range testTable {
//		res := make([]Paginator, 0, test.Params.PageSize)
//
//		paginationRes, err := suite.realConnector.Repo(Paginator{}).SelectWithPagePagination(
//			suite.ctx,
//			squirrel.Select(cols...).From(table).OrderBy("id DESC"),
//			test.Params,
//			&res)
//
//		assert.Nil(t, err)
//		assert.NotNil(t, res)
//		assert.NotNil(t, paginationRes)
//		assert.Equal(t, test.LenData, len(res))
//		assert.Equal(t, test.Results, paginationRes)
//	}
//}
