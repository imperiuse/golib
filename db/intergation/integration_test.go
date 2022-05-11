package intergation

import (
	"context"
	"database/sql"
	"fmt"
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
	"github.com/imperiuse/golib/db/genrepo/emptygen"
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
			N:    i + 1,
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

	for i, c := range []db.Connector[config.SimpleTestConfig]{
		suite.connector,
		suite.connectorWithValidation,
		suite.connectorWithValidationAndCache,
	} {
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
		n, err := c.AutoUpdate(suite.ctx, u)
		assert.Equal(t, int64(0), n)
		assert.Nil(t, err)

		// 		b) generic repo way
		n, err = repo.NewGen[dto.ID, dto.User[dto.ID]](c).Update(suite.ctx, u.ID(), u)
		assert.Equal(t, int64(0), n)
		assert.Nil(t, err)

		// 		c) old school (pure connection) way
		res2, err := c.Connection().ExecContext(suite.ctx,
			fmt.Sprintf("UPDATE %s SET name=$1 WHERE id=$2;", u.Repo()), u.Name, u.ID())
		assert.NotNil(t, res2)
		assert.Nil(t, err)

		// 	4) Check we delete without error User. (User not exist)
		// 		a) repo way
		n, err = c.AutoDelete(suite.ctx, u)
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
		// 		a) repo way
		id, err := c.AutoCreate(suite.ctx, r)
		r.Id = dto.ID(id)
		assert.Nil(t, err)

		// 		b) generic repo way
		u.RoleID = r.Id
		u.Id, err = repo.NewGen[dto.ID, dto.User[dto.ID]](c).Create(suite.ctx, u)
		assert.Nil(t, err)

		// 2) Get Role and User
		// 		a) repo way
		r2 = dto.Role[dto.ID]{BaseDTO: dto.BaseDTO[dto.ID]{Id: r.ID()}}
		assert.Nil(t, c.AutoGet(suite.ctx, &r2))
		assert.Equal(t, r.Name, r2.Name)
		assert.Equal(t, r.Rights, r2.Rights)
		assert.NotEqual(t, r.CreatedAt, r2.CreatedAt) // db set up it auto

		// 		b) generic repo way
		u3, err := repo.NewGen[dto.ID, dto.User[dto.ID]](c).Get(suite.ctx, u.ID())
		assert.Equal(t, u.Name, u3.Name)
		assert.Equal(t, u.Email, u3.Email)
		assert.NotEqual(t, u.CreatedAt, u3.CreatedAt) // db set up it auto
		assert.Nil(t, err)

		// 3) Update Role and User
		// 		a) repo way
		r.Name = "New Role"
		r.Rights = 7
		n, err = c.AutoUpdate(suite.ctx, r)
		assert.Equal(t, int64(1), n)
		assert.Nil(t, err)

		// 		b) generic repo way
		u.Name = "New User"
		u.Email = "new-user@mail.com"

		n, err = repo.NewGen[dto.ID, dto.User[dto.ID]](c).Update(suite.ctx, u.ID(), u)
		assert.Equal(t, int64(1), n)
		assert.Nil(t, err)

		// 4) Get again and check update is apply
		//      c) old school (pure connection) way I
		var r4 = dto.Role[dto.ID]{BaseDTO: dto.BaseDTO[dto.ID]{Id: r2.ID()}}
		assert.Nil(t,
			sqlx.GetContext(
				suite.ctx,
				c.Connection(),
				&r4,
				fmt.Sprintf("SELECT * FROM %s WHERE id=$1 LIMIT 1;", r4.Repo()), r4.ID(),
			),
		)

		assert.Equal(t, r.Name, r4.Name)
		assert.Equal(t, r.Rights, r4.Rights)
		assert.NotEqual(t, r.CreatedAt, r4.CreatedAt) // db set up it auto

		//      c) old school (pure connection) way II
		var u4 = dto.User[dto.ID]{BaseDTO: dto.BaseDTO[dto.ID]{Id: u.Id}}
		assert.Nil(t,
			sqlx.GetContext(
				suite.ctx,
				c.Connection(),
				&u4,
				fmt.Sprintf("SELECT * FROM %s WHERE id=$1 LIMIT 1;", u4.Repo()), u4.ID(),
			),
		)
		assert.Equal(t, u.Name, u4.Name)
		assert.Equal(t, u.Email, u4.Email)
		assert.NotEqual(t, u.CreatedAt, u4.CreatedAt) // db set up it auto

		// III. Tests FindOneByWithInnerJoin
		{
			cols, joinCond := orm.GetDataForSelect(&dto.UsersRole[dto.ID]{})
			nameWithAlias := orm.GetTableNameWithAlias(&dto.User[dto.ID]{})
			joinCond = orm.GetTableNameWithAlias(&dto.Role[dto.ID]{}) + " " + joinCond

			var ur dto.UsersRole[dto.ID]
			err := c.Repo(u).FindOneByWithInnerJoin(suite.ctx, cols, nameWithAlias, joinCond, squirrel.Eq{"u.id": u.ID()}, &ur)
			assert.Nil(t, err)
			assert.NotNil(t, ur)
			assert.Equal(t, u.ID(), ur.User.ID())
			assert.Equal(t, u.Name, ur.User.Name)
			assert.Equal(t, u.Email, ur.User.Email)
			assert.Equal(t, u.Password, ur.User.Password)
			assert.Equal(t, u.RoleID, ur.Role.ID())
			assert.Equal(t, r.Rights, ur.Role.Rights)

			var url = make([]dto.UsersRole[dto.ID], 0)
			err = c.Repo(u).FindByWithInnerJoin(suite.ctx, cols, nameWithAlias, joinCond, squirrel.Eq{"u.id": u.ID()}, &url)
			assert.Nil(t, err)
			assert.NotNil(t, url)
			ur = url[0]
			assert.Equal(t, u.ID(), ur.User.ID())
			assert.Equal(t, u.Name, ur.User.Name)
			assert.Equal(t, u.Email, ur.User.Email)
			assert.Equal(t, u.Password, ur.User.Password)
			assert.Equal(t, u.RoleID, ur.Role.ID())
			assert.Equal(t, r.Rights, ur.Role.Rights)
		}

		// IV. Part delete all records
		n, err = c.AutoDelete(suite.ctx, u)
		assert.Equal(t, int64(1), n)
		assert.Nil(t, err)

		n, err = c.AutoDelete(suite.ctx, r)
		assert.Equal(t, int64(1), n)
		assert.Nil(t, err)
	}
}

func (suite *RepositoryTestSuit) Test_GetRowsByQuery() {
	t := suite.T()

	for _, c := range []db.Connector[config.SimpleTestConfig]{
		suite.connector,
		suite.connectorWithValidation,
		suite.connectorWithValidationAndCache,
	} {
		r, err := c.Repo(dto.User[dto.ID]{}).GetRowsByQuery(suite.ctx, squirrel.SelectBuilder{}.Columns("*").Where("1=1"))
		assert.Nil(t, err)
		assert.NotNil(t, r)

		r, err = repo.NewGen[dto.ID, dto.User[dto.ID]](c).GetRowsByQuery(suite.ctx, squirrel.SelectBuilder{}.Columns("*").Where("1=1"))
		assert.Nil(t, err)
		assert.NotNil(t, r)
	}

}

func (suite *RepositoryTestSuit) Test_CountByQuery() {
	t := suite.T()

	for _, c := range []db.Connector[config.SimpleTestConfig]{
		suite.connector,
		suite.connectorWithValidation,
		suite.connectorWithValidationAndCache,
	} {
		n, err := c.Repo(dto.User[dto.ID]{}).CountByQuery(suite.ctx, squirrel.Select("count(1)"))
		assert.Nil(t, err)
		assert.Equal(t, uint64(0), n)

		n, err = c.Repo(dto.Paginator[dto.ID]{}).CountByQuery(suite.ctx, squirrel.Select("count(1)"))
		assert.Nil(t, err)
		assert.Equal(t, uint64(200), n)

		n, err = repo.NewGen[dto.ID, dto.User[dto.ID]](c).CountByQuery(suite.ctx, squirrel.Select("count(1)"))
		assert.Nil(t, err)
		assert.Equal(t, uint64(0), n)

		n, err = repo.NewGen[dto.ID, dto.Paginator[dto.ID]](c).CountByQuery(suite.ctx, squirrel.Select("count(1)"))
		assert.Nil(t, err)
		assert.Equal(t, uint64(200), n)
	}
}

func (suite *RepositoryTestSuit) Test_FindBy() {
	t := suite.T()

	for _, c := range []db.Connector[config.SimpleTestConfig]{
		suite.connector,
		suite.connectorWithValidation,
		suite.connectorWithValidationAndCache,
	} {
		cols := orm.GetDataForSelectOnlyCols(dto.Paginator[dto.ID]{})

		var res = make([]dto.Paginator[dto.ID], 0)
		err := c.Repo(dto.Paginator[dto.ID]{}).FindBy(suite.ctx, cols, squirrel.Eq{"1": "1"}, &res)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, 200, len(res))

		res, err = repo.NewGen[dto.ID, dto.Paginator[dto.ID]](c).FindBy(suite.ctx, cols, squirrel.Eq{"1": "1"})

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, 200, len(res))

	}
}

func (suite *RepositoryTestSuit) Test_FindOne() {
	t := suite.T()

	for _, c := range []db.Connector[config.SimpleTestConfig]{
		suite.connector,
		suite.connectorWithValidation,
		suite.connectorWithValidationAndCache,
	} {
		cols := orm.GetDataForSelectOnlyCols(dto.Paginator[dto.ID]{})

		var res dto.Paginator[dto.ID]
		err := c.Repo(dto.Paginator[dto.ID]{}).FindOneBy(suite.ctx, cols, squirrel.Eq{"id": "77"}, &res)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, 77, res.ID())

		res, err = repo.NewGen[dto.ID, dto.Paginator[dto.ID]](c).FindOneBy(suite.ctx, cols, squirrel.Eq{"id": "77"})

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, 77, res.ID())

	}
}

func (suite *RepositoryTestSuit) Test_Select() {
	t := suite.T()

	for _, c := range []db.Connector[config.SimpleTestConfig]{
		suite.connector,
		suite.connectorWithValidation,
		suite.connectorWithValidationAndCache,
	} {
		var res = make([]dto.Paginator[dto.ID], 0)
		err := c.Repo(dto.Paginator[dto.ID]{}).Select(suite.ctx, squirrel.Select("*").Where("1=1"), &res)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, 200, len(res))

		res, err = repo.NewGen[dto.ID, dto.Paginator[dto.ID]](c).Select(suite.ctx, squirrel.Select("*").Where("1=1"))

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, 200, len(res))
	}
}

func (suite *RepositoryTestSuit) Test_SelectWithPagePagination() {
	t := suite.T()

	type Test struct {
		Params     db.PagePaginationParams
		Results    db.PagePaginationResults
		GenResults db.PagePaginationResults // for generic way (we could get page numbers!)
		FirstN     int
		LastN      int
		LenData    int
	}

	testTable := []Test{
		{
			Params: db.PagePaginationParams{
				PageNumber: 0,
				PageSize:   50,
			},
			Results: db.PagePaginationResults{
				CurrentPageNumber: 0,
				NextPageNumber:    0,
				CntPages:          4,
			},
			GenResults: db.PagePaginationResults{
				CurrentPageNumber: 0,
				NextPageNumber:    1,
				CntPages:          4,
			},
			FirstN:  200,
			LastN:   151,
			LenData: 50,
		},
		{
			Params: db.PagePaginationParams{
				PageNumber: 1,
				PageSize:   49,
			},
			Results: db.PagePaginationResults{
				CurrentPageNumber: 1,
				NextPageNumber:    0,
				CntPages:          5,
			},
			GenResults: db.PagePaginationResults{
				CurrentPageNumber: 1,
				NextPageNumber:    2,
				CntPages:          5,
			},
			FirstN:  200,
			LastN:   152,
			LenData: 49,
		},

		{
			Params: db.PagePaginationParams{
				PageNumber: 2,
				PageSize:   50,
			},
			Results: db.PagePaginationResults{
				CurrentPageNumber: 2,
				NextPageNumber:    0,
				CntPages:          4,
			},
			GenResults: db.PagePaginationResults{
				CurrentPageNumber: 2,
				NextPageNumber:    3,
				CntPages:          4,
			},
			FirstN:  150,
			LastN:   101,
			LenData: 50,
		},
		{
			Params: db.PagePaginationParams{
				PageNumber: 5,
				PageSize:   50,
			},
			Results: db.PagePaginationResults{
				CurrentPageNumber: 5,
				NextPageNumber:    0,
				CntPages:          4,
			},
			GenResults: db.PagePaginationResults{
				CurrentPageNumber: 5,
				NextPageNumber:    0,
				CntPages:          4,
			},
			FirstN:  0,
			LastN:   50,
			LenData: 0,
		},
	}

	cols, _ := orm.GetDataForSelect(&dto.Paginator[dto.ID]{})

	for _, c := range []db.Connector[config.SimpleTestConfig]{
		suite.connector,
		suite.connectorWithValidation,
		suite.connectorWithValidationAndCache,
	} {

		for i, test := range testTable {
			res := make([]dto.Paginator[dto.ID], 0, test.Params.PageSize)

			// classic repo way
			paginationRes, err := c.Repo(dto.Paginator[dto.ID]{}).SelectWithPagePagination(
				suite.ctx,
				squirrel.Select(cols...).OrderBy("id DESC"),
				test.Params,
				&res,
			)

			assert.Nil(t, err)
			assert.NotNil(t, res)
			assert.NotNil(t, paginationRes)
			assert.Equal(t, test.LenData, len(res))
			assert.Equal(t, test.Results, paginationRes)
			if l := len(res) - 1; l > 0 {
				assert.Equalf(t, test.FirstN, res[0].N, "repo test:", i)
				assert.Equalf(t, test.LastN, res[l].N, "repo: test", i)
			}

			// generic way
			res = make([]dto.Paginator[dto.ID], 0, test.Params.PageSize)
			res, paginationRes, err = repo.NewGen[dto.ID, dto.Paginator[dto.ID]](c).SelectWithPagePagination(
				suite.ctx,
				squirrel.Select(cols...).OrderBy("id DESC"),
				test.Params,
			)

			assert.Nil(t, err)
			assert.NotNil(t, res)
			assert.NotNil(t, paginationRes)
			assert.Equal(t, test.LenData, len(res))
			assert.Equal(t, test.GenResults, paginationRes, "gen: test:", i)
			if l := len(res) - 1; l > 0 {
				assert.Equalf(t, test.FirstN, res[0].N, "gen: test", i)
				assert.Equalf(t, test.LastN, res[l].N, "gen: test", i)
			}
		}
	}
}

func (suite *RepositoryTestSuit) Test_SelectWithCursorPagination() {
	t := suite.T()

	type Test struct {
		Params  db.CursorPaginationParams
		FirstN  int
		LastN   int
		LenData int
	}

	testTable := []Test{
		{
			Params: db.CursorPaginationParams{
				Limit:     30,
				Cursor:    0,
				DescOrder: false,
			},
			FirstN:  1,
			LastN:   30,
			LenData: 30,
		},
		{
			Params: db.CursorPaginationParams{
				Limit:     30,
				Cursor:    201,
				DescOrder: true,
			},
			FirstN:  200,
			LastN:   171,
			LenData: 30,
		},
		{
			Params: db.CursorPaginationParams{
				Limit:     30,
				Cursor:    10,
				DescOrder: false,
			},
			FirstN:  11,
			LastN:   40,
			LenData: 30,
		},
		{
			Params: db.CursorPaginationParams{
				Limit:     20,
				Cursor:    10,
				DescOrder: true,
			},
			FirstN:  9,
			LastN:   1,
			LenData: 9,
		},
		{
			Params: db.CursorPaginationParams{
				Limit:     250,
				Cursor:    0,
				DescOrder: true,
			},
			FirstN:  200,
			LastN:   1,
			LenData: 0,
		},
		{
			Params: db.CursorPaginationParams{
				Limit:     250,
				Cursor:    250,
				DescOrder: true,
			},
			FirstN:  200,
			LastN:   1,
			LenData: 200,
		},
		{
			Params: db.CursorPaginationParams{
				Limit:     250,
				Cursor:    0,
				DescOrder: false,
			},
			FirstN:  1,
			LastN:   200,
			LenData: 200,
		},
		{
			Params: db.CursorPaginationParams{
				Limit:     100,
				Cursor:    0,
				DescOrder: false,
			},
			FirstN:  1,
			LastN:   100,
			LenData: 100,
		},
		{
			Params: db.CursorPaginationParams{
				Limit:     100,
				Cursor:    50,
				DescOrder: false,
			},
			FirstN:  51,
			LastN:   150,
			LenData: 100,
		},
		{
			Params: db.CursorPaginationParams{
				Limit:     100,
				Cursor:    50,
				DescOrder: true,
			},
			FirstN:  49,
			LastN:   1,
			LenData: 49,
		},
	}

	cols, _ := orm.GetDataForSelect(&dto.Paginator[dto.ID]{})

	for _, c := range []db.Connector[config.SimpleTestConfig]{
		suite.connector,
		suite.connectorWithValidation,
		suite.connectorWithValidationAndCache,
	} {

		for i, test := range testTable {
			res := make([]dto.Paginator[dto.ID], 0, test.Params.Limit)

			// classic repo way
			err := c.Repo(dto.Paginator[dto.ID]{}).SelectWithCursorOnPKPagination(
				suite.ctx,
				squirrel.Select(cols...),
				test.Params,
				&res,
			)

			assert.Nil(t, err)
			assert.NotNil(t, res)
			assert.Equalf(t, test.LenData, len(res), "case: ", i)
			if l := len(res) - 1; l > 0 {
				assert.Equal(t, test.FirstN, res[0].N)
				assert.Equal(t, test.LastN, res[len(res)-1].N)
			}

			// generic way
			res = make([]dto.Paginator[dto.ID], 0, test.Params.Limit)
			res, err = repo.NewGen[dto.ID, dto.Paginator[dto.ID]](c).SelectWithCursorOnPKPagination(
				suite.ctx,
				squirrel.Select(cols...),
				test.Params,
			)

			assert.Nil(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, test.LenData, len(res))
			if l := len(res) - 1; l > 0 {
				assert.Equal(t, test.FirstN, res[0].N)
				assert.Equal(t, test.LastN, res[len(res)-1].N)
			}
		}
	}
}
