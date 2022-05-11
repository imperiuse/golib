package mocks

import (
	"database/sql"
	"database/sql/driver"

	"github.com/imperiuse/golib/db"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
)

const (
	FakeStringAns = "fake_string_ans"
)

var (
	BadMockDBConn = func() db.PureSqlxConnection {
		mockCtxArg := mock.AnythingOfType("*context.cancelCtx")
		mockCtxEmpty := mock.AnythingOfType("*context.emptyCtx")
		mockSqlTxOptions := mock.AnythingOfType("*sql.TxOptions")
		mockStringArg := mock.AnythingOfType("string")

		conn := PureSqlxConnection{}
		conn.On("DriverName").Return(FakeStringAns)
		conn.On("Rebind", mockStringArg).Return(FakeStringAns)
		conn.On("BindNamed", mockStringArg, mock.Anything).Return(FakeStringAns, nil, db.ErrInvalidRepoEmptyRepo)
		conn.On("QueryContext", mockCtxArg, mockStringArg, mock.Anything).Return(&sql.Rows{}, db.ErrInvalidRepoEmptyRepo)
		conn.On("QueryContext", mockCtxEmpty, mockStringArg, mock.Anything).Return(&sql.Rows{}, db.ErrInvalidRepoEmptyRepo)
		conn.On("QueryxContext", mockCtxArg, mockStringArg, mock.Anything).
			Return(&sqlx.Rows{}, db.ErrInvalidRepoEmptyRepo)
		conn.On("QueryxContext", mockCtxEmpty, mockStringArg, mock.Anything).
			Return(&sqlx.Rows{}, db.ErrInvalidRepoEmptyRepo)
		conn.On("QueryRowxContext", mockCtxArg, mockStringArg, mock.Anything).Return(&sqlx.Row{})
		conn.On("QueryRowxContext", mockCtxEmpty, mockStringArg, mock.Anything).Return(&sqlx.Row{})
		conn.On("ExecContext", mockCtxArg, mockStringArg, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).
			Return(driver.RowsAffected(0), db.ErrInvalidRepoEmptyRepo)
		conn.On("ExecContext", mockCtxEmpty, mockStringArg, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).
			Return(driver.RowsAffected(0), db.ErrInvalidRepoEmptyRepo)
		conn.On("PrepareContext", mockCtxArg, mockStringArg).Return(&sql.Stmt{}, db.ErrInvalidRepoEmptyRepo)
		conn.On("BeginTxx", mockCtxArg, mock.Anything).Return(&sqlx.Tx{}, db.ErrInvalidRepoEmptyRepo)
		conn.On("BeginTxx", mockCtxEmpty, mockSqlTxOptions).Return(&sqlx.Tx{}, db.ErrInvalidRepoEmptyRepo)

		return &conn
	}()

	GoodMockDBConn = func() db.PureSqlxConnection {
		mockCtxArg := mock.AnythingOfType("*context.cancelCtx")
		mockCtxEmpty := mock.AnythingOfType("*context.emptyCtx")
		mockSqlTxOptions := mock.AnythingOfType("*sql.TxOptions")
		mockStringArg := mock.AnythingOfType("string")

		conn := PureSqlxConnection{}
		conn.On("DriverName").Return(FakeStringAns)
		conn.On("Rebind", mockStringArg).Return(FakeStringAns)
		conn.On("BindNamed", mockStringArg, mock.Anything).Return(FakeStringAns, nil, nil)
		conn.On("QueryContext", mockCtxArg, mockStringArg, mock.Anything).Return(&sql.Rows{}, nil)
		conn.On("QueryContext", mockCtxEmpty, mockStringArg, mock.Anything).Return(&sql.Rows{}, nil)
		conn.On("QueryxContext", mockCtxArg, mockStringArg, mock.Anything).Return(&sqlx.Rows{}, nil)
		conn.On("QueryxContext", mockCtxEmpty, mockStringArg, mock.Anything).Return(&sqlx.Rows{}, nil)
		conn.On("QueryRowxContext", mockCtxArg, mockStringArg, mock.Anything).Return(&sqlx.Row{})
		conn.On("QueryRowxContext", mockCtxEmpty, mockStringArg, mock.Anything).Return(&sqlx.Row{})
		conn.On("ExecContext", mockCtxArg, mockStringArg, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).
			Return(driver.RowsAffected(0), nil)
		conn.On("ExecContext", mockCtxEmpty, mockStringArg, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).
			Return(driver.RowsAffected(0), nil)
		conn.On("PrepareContext", mockCtxArg, mockStringArg).Return(&sql.Stmt{}, nil)
		conn.On("BeginTxx", mockCtxArg, mock.Anything).Return(&sqlx.Tx{}, nil)
		conn.On("BeginTxx", mockCtxEmpty, mockSqlTxOptions).Return(&sqlx.Tx{}, nil)

		return &conn
	}()
)
