//nolint golint
package repository

import (
	"context"
	"database/sql"

	"github.com/imperiuse/golib/reflect/orm"
	"github.com/imperiuse/golib/sqlx/helper"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var ErrNotFoundAnyRepo = errors.New("not found any repo (connectors)")

//go:generate mockery --name=SqlxDBConnectorI
type (
	SqlxDBConnectorI interface {
		sqlx.QueryerContext
		sqlx.ExecerContext
		sqlx.ExtContext
		sqlx.PreparerContext
		helper.TxxI
	}

	Repo = Table

	Repositories map[Repo]*repository

	DTO             = interface{}
	DtoWithIdentity = interface {
		Identity() ID
	}

	RepositoriesI interface {
		PureConnector() SqlxDBConnectorI

		Repo(Repo) Repository
		AutoRepo(DTO) Repository

		AutoCreate(context.Context, DtoWithIdentity) (ID, error)
		AutoGet(context.Context, DtoWithIdentity) error
		AutoUpdate(context.Context, DtoWithIdentity) (int64, error)
		AutoDelete(context.Context, DtoWithIdentity) (int64, error)
	}

	Repository interface {
		// Name of main table (repo obj)
		Name() Table

		// Pure Sqlx Db Connector which we pass to NewSqlxMapRepo
		PureConnector() SqlxDBConnectorI

		// usual "CRUD"
		Create(context.Context, DTO) (ID, error)
		Get(context.Context, ID, DTO) error
		Update(context.Context, ID, DTO) (int64, error)
		Delete(context.Context, ID) (int64, error)

		Insert(context.Context, []Column, []Argument) (int64, error)
		UpdateCustom(context.Context, map[string]interface{}, Condition) (int64, error)

		FindBy(context.Context, []Column, Condition, DTO) error
		FindOneBy(context.Context, []Column, Condition, DTO) error

		FindByWithInnerJoin(context.Context, []Column, Alias, Join, Condition, DTO) error
		FindOneByWithInnerJoin(context.Context, []Column, Alias, Join, Condition, DTO) error

		Select(context.Context, SelectBuilder, DTO) error
		SelectWithPagePagination(context.Context, SelectBuilder, PagePaginationParams, DTO) (PagePaginationResults, error)
		SelectWithCursorOnPKPagination(context.Context, SelectBuilder, CursorPaginationParams, DTO) error

		GetRowsByQuery(ctx context.Context, qb SelectBuilder) (*sql.Rows, error)
		CountByQuery(ctx context.Context, qb SelectBuilder) (uint64, error)
	}

	Column = string

	Argument = interface{}
)

func NewSqlxMapRepo(logger ZapLogger, db SqlxDBConnectorI, phf PlaceholderFormat, tables []Table, objs []DTO) Repositories {
	mapRepo := make(Repositories, len(tables))
	for _, name := range tables {
		mapRepo[name] = newRepository(logger, db, name, phf)
	}

	for _, obj := range objs {
		tableName := orm.GetTableName(obj)
		if tableName == orm.Undefined {
			continue
		}

		if _, found := mapRepo[tableName]; found {
			continue
		}

		mapRepo[tableName] = newRepository(logger, db, tableName, phf)
	}

	return mapRepo
}

func (r Repositories) PureConnector() SqlxDBConnectorI {
	for _, repo := range r {
		return repo.PureConnector()
	}

	return emptyRepo.PureConnector()
}

func (r Repositories) Repo(name Repo) Repository {
	if rep, found := r[name]; found {
		return rep
	}

	return emptyRepo
}

func (r Repositories) AutoRepo(obj DTO) Repository {
	if rep, found := r[orm.GetMetaDTO(obj).TableName]; found {
		return rep
	}

	return emptyRepo
}

func (r Repositories) AutoCreate(ctx context.Context, obj DtoWithIdentity) (ID, error) {
	return r.AutoRepo(obj).Create(ctx, obj)
}

func (r Repositories) AutoUpdate(ctx context.Context, obj DtoWithIdentity) (int64, error) {
	return r.AutoRepo(obj).Update(ctx, obj.Identity(), obj)
}

func (r Repositories) AutoGet(ctx context.Context, obj DtoWithIdentity) error {
	return r.AutoRepo(obj).Get(ctx, obj.Identity(), obj)
}

func (r Repositories) AutoDelete(ctx context.Context, obj DtoWithIdentity) (int64, error) {
	return r.AutoRepo(obj).Delete(ctx, obj.Identity())
}
