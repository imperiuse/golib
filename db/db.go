package db

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/Masterminds/squirrel"
	"github.com/imperiuse/golib/db/transaction"
)

type (
	// Logger - Uber Zap logger
	Logger = *zap.Logger

	// Condition - squirrel.Sqlizer
	Condition = squirrel.Sqlizer // squirrel.Eq or squirrel.Gt or squirrel.And and etc

	// SelectBuilder - squirrel.SelectBuilder
	SelectBuilder = squirrel.SelectBuilder

	// PlaceholderFormat - squirrel.PlaceholderFormat
	PlaceholderFormat = squirrel.PlaceholderFormat

	// Query - sql query string.
	Query = string
	// Table - sql table
	Table = string
	// Column - sql column name.
	Column = string
	// Argument - anything ibj which have Valuer and Scan.
	Argument = any
	// Alias - alias for table.
	Alias = string
	// Join - join part of query.
	Join = string
	// Obj - any obj (any).
	Obj = any
	// ID - uniq ID
	ID = any
)

//go:generate mockery --name=PureSqlxConnection
type (
	// Config - configuration obj for Storage
	Config = interface {
		IsEnableValidationRepoNames() bool
		IsEnableReposCache() bool
		PlaceholderFormat() PlaceholderFormat // squirrel int code for placeholder
	}

	DTO interface {
		Repo() Table
		Identity() ID
	}

	// NB! Idea for future ->
	// type Resource[I any] interface {
	//      GetID() I
	//  }
	//
	//  type Storage[I any, R Resource[I]] interface {
	//       GetByID(id I) R
	//   }
	// DTO[I ID] interface {
	//	Repo() Table
	//	Identity() I
	// }

	PureSqlxConnection interface {
		sqlx.QueryerContext
		sqlx.ExecerContext
		sqlx.ExtContext
		sqlx.PreparerContext
		transaction.TxxI
	}

	// Storage - abstract Storage interface general storage interface.
	Storage[C Config] interface {
		Config() C

		Connect() error
		Reconnect() error

		OnConnect(context.Context, func()) error   // actions that will be performed after function call Connect()
		OnReconnect(context.Context, func()) error // actions that will be performed after function call Reconnect()
		OnStop(context.Context, func()) error      // actions that will be performed after function call Close()

		Master() Connector[C]
		Slaves() []Connector[C]

		Close() error
	}

	// Connector - entity for describe connection to specific DB instance
	Connector[C Config] interface {
		Config() C
		AddRepoNames(...Table)

		Logger() Logger

		Connection() PureSqlxConnection

		Repo(DTO) Repository
		// TODO when Go in next versions will support generics in methods
		// Repo[I ID, D DTO]() gRepository[I, D] // refactor to this NOW try use this ->
		// repository.NewGen[I, DTO]](connector) -> return GRepository

		AutoCreate(context.Context, DTO) (int64, error)
		AutoGet(context.Context, DTO) error
		AutoUpdate(context.Context, DTO) (int64, error)
		AutoDelete(context.Context, DTO) (int64, error)
	}

	// BaseRepositoryI - base method for all type Repo's
	BaseRepositoryI interface {
		Name() Table // Name of table/repo (repo obj)

		GetRowsByQuery(ctx context.Context, qb SelectBuilder) (*sql.Rows, error)
		CountByQuery(ctx context.Context, qb SelectBuilder) (uint64, error)

		Insert(context.Context, []Column, []Argument) (int64, error)
		UpdateCustom(context.Context, map[string]any, Condition) (int64, error)
	}

	// Repository - methods for classic Repo's (non generics)
	Repository interface {
		BaseRepositoryI

		Create(context.Context, any) (int64, error) // todo add one method for ID = string or move to generics API only
		Get(context.Context, ID, any) error
		Update(context.Context, ID, any) (int64, error)
		Delete(context.Context, ID) (int64, error)

		FindBy(context.Context, []Column, Condition, any) error
		FindOneBy(context.Context, []Column, Condition, any) error

		FindByWithInnerJoin(context.Context, []Column, Alias, Join, Condition, any) error
		FindOneByWithInnerJoin(context.Context, []Column, Alias, Join, Condition, any) error

		Select(context.Context, SelectBuilder, any) error
		SelectWithPagePagination(context.Context, SelectBuilder, PagePaginationParams, any) (PagePaginationResults, error)
		SelectWithCursorOnPKPagination(context.Context, SelectBuilder, CursorPaginationParams, any) error
	}

	// GRepository - methods for new (modern) approach generic based Repo's
	GRepository[I ID, D DTO] interface {
		BaseRepositoryI

		Create(context.Context, D) (I, error)
		Get(context.Context, I) (D, error)
		Update(context.Context, I, D) (int64, error)
		Delete(context.Context, I) (int64, error)

		FindBy(context.Context, []Column, Condition) ([]D, error)
		FindOneBy(context.Context, []Column, Condition) (D, error)

		FindByWithInnerJoin(context.Context, []Column, Alias, Join, Condition) ([]D, error)
		FindOneByWithInnerJoin(context.Context, []Column, Alias, Join, Condition) (D, error)

		Select(context.Context, SelectBuilder) ([]D, error)
		SelectWithPagePagination(context.Context, SelectBuilder, PagePaginationParams) ([]D, PagePaginationResults, error)
		SelectWithCursorOnPKPagination(context.Context, SelectBuilder, CursorPaginationParams) ([]D, error)
	}
)

type (
	PagePaginationParams struct {
		PageNumber uint64
		PageSize   uint64
	}

	PagePaginationResults struct {
		CurrentPageNumber uint64
		NextPageNumber    uint64
		CntPages          uint64
	}

	CursorPaginationParams struct {
		Limit     uint64
		Cursor    uint64
		DescOrder bool
	}
)

var (
	ErrInvalidRepoEmptyRepo = errors.New("invalid repo (empty repo). Not registered?" +
		" Check this usage connector.AddRepoNames(repos ...db.Table)")
	ErrMismatchRowsCnt = errors.New("mismatch rows counts")
	ErrZeroPageSize    = errors.New("zero value of params.PageSize")
	ErrZeroLimitSize   = errors.New("zero value of params.Limit")
)
