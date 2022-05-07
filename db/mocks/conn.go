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
		mockStringArg := mock.AnythingOfType("string")

		conn := PureSqlxConnection{}
		conn.On("DriverName").Return(FakeStringAns)
		conn.On("Rebind", mockStringArg).Return(FakeStringAns)
		conn.On("BindNamed", mockStringArg, mock.Anything).Return(FakeStringAns, nil, db.ErrInvalidRepo)
		conn.On("QueryContext", mockCtxArg, mockStringArg, mock.Anything).Return(&sql.Rows{}, db.ErrInvalidRepo)
		conn.On("QueryxContext", mockCtxArg, mockStringArg, mock.Anything).
			Return(&sqlx.Rows{}, db.ErrInvalidRepo)
		conn.On("QueryRowxContext", mockCtxArg, mockStringArg, mock.Anything).Return(&sqlx.Row{})
		conn.On("ExecContext", mockCtxArg, mockStringArg, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).
			Return(driver.RowsAffected(0), db.ErrInvalidRepo)
		conn.On("PrepareContext", mockCtxArg, mockStringArg).Return(&sql.Stmt{}, db.ErrInvalidRepo)
		conn.On("BeginTxx", mockCtxArg, mock.Anything).Return(&sqlx.Tx{}, db.ErrInvalidRepo)

		return &conn
	}()

	GoodMockDBConn = func() db.PureSqlxConnection {
		mockCtxArg := mock.AnythingOfType("*context.cancelCtx")
		mockStringArg := mock.AnythingOfType("string")

		conn := PureSqlxConnection{}
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
