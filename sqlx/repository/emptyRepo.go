package repository

import (
	"database/sql"
	"database/sql/driver"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/imperiuse/golib/sqlx/repository/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
)

//nolint golint
const (
	FakeStringAns = "fake_string_ans"
)

//nolint dupl
var (
	// ErrEmptyRepo err empty repo.
	ErrEmptyRepo = errors.New("repository: emptyRepo, can't found Repo by name or obj, please check arguments for methods Repo AutoRepo") // nolint lll

	emptyRepo = func() *repository {
		return &repository{
			logger: zap.NewNop(),
			db:     badMockDBConn,
			name:   "_emptyRepo_",
		}
	}()

	badMockDBConn = func() SqlxDBConnectorI {
		mockCtxArg := mock.AnythingOfType("*context.cancelCtx")
		mockStringArg := mock.AnythingOfType("string")

		conn := mocks.SqlxDBConnectorI{}
		conn.On("DriverName").Return(FakeStringAns)
		conn.On("Rebind", mockStringArg).Return(FakeStringAns)
		conn.On("BindNamed", mockStringArg, mock.Anything).Return(FakeStringAns, nil, ErrEmptyRepo)
		conn.On("QueryContext", mockCtxArg, mockStringArg, mock.Anything).Return(&sql.Rows{}, ErrEmptyRepo)
		conn.On("QueryxContext", mockCtxArg, mockStringArg, mock.Anything).
			Return(&sqlx.Rows{}, ErrEmptyRepo)
		conn.On("QueryRowxContext", mockCtxArg, mockStringArg, mock.Anything).Return(&sqlx.Row{})
		conn.On("ExecContext", mockCtxArg, mockStringArg, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).
			Return(driver.RowsAffected(0), ErrEmptyRepo)
		conn.On("PrepareContext", mockCtxArg, mockStringArg).Return(&sql.Stmt{}, ErrEmptyRepo)
		conn.On("BeginTxx", mockCtxArg, mock.Anything).Return(&sqlx.Tx{}, ErrEmptyRepo)

		return &conn
	}()

	goodMockDBConn = func() SqlxDBConnectorI {
		mockCtxArg := mock.AnythingOfType("*context.cancelCtx")
		mockStringArg := mock.AnythingOfType("string")

		conn := mocks.SqlxDBConnectorI{}
		conn.On("DriverName").Return(FakeStringAns)
		conn.On("Rebind", mockStringArg).Return(FakeStringAns)
		conn.On("BindNamed", mockStringArg, mock.Anything).Return(FakeStringAns, nil, nil)
		conn.On("QueryContext", mockCtxArg, mockStringArg, mock.Anything).Return(&sql.Rows{}, nil)
		conn.On("QueryxContext", mockCtxArg, mockStringArg, mock.Anything).Return(&sqlx.Rows{}, nil)
		conn.On("QueryRowxContext", mockCtxArg, mockStringArg, mock.Anything).Return(&sqlx.Row{})
		conn.On("ExecContext", mockCtxArg, mockStringArg, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).
			Return(driver.RowsAffected(0), nil)
		conn.On("PrepareContext", mockCtxArg, mockStringArg).Return(&sql.Stmt{}, nil)
		conn.On("BeginTxx", mockCtxArg, mock.Anything).Return(&sqlx.Tx{}, nil)

		return &conn
	}()
)
