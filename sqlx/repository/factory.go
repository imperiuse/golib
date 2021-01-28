package repository

import (
	"context"
	"database/sql"

	"github.com/imperiuse/golib/reflect/orm"

	"github.com/Masterminds/squirrel"
	"github.com/imperiuse/golib/sqlx/helper"
	"github.com/jmoiron/sqlx"
)

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

	DTO   = interface{}
	DTOID = interface {
		Id() ID
	}

	RepositoriesI interface {
		Repo(Repo) Repository
		AutoRepo(DTO) Repository

		AutoCreate(context.Context, DTOID) (ID, error)
		AutoGet(context.Context, DTOID) error
		AutoUpdate(context.Context, DTOID) (int64, error)
		AutoDelete(context.Context, DTOID) (int64, error)
	}

	Repository interface {
		// Pure Sqlx Db Connector which we pass to NewSqlxMapRepo
		SqlxDBConnectorI() SqlxDBConnectorI

		// usual "CRUD"
		Create(context.Context, DTO) (ID, error)
		Get(context.Context, ID, DTO) error
		Update(context.Context, ID, DTO) (int64, error)
		Delete(context.Context, ID) (int64, error)

		Insert(context.Context, []Column, []Argument) (int64, error)
		UpdateCustom(context.Context, map[string]interface{}, squirrel.Eq) (int64, error)
		FindBy(context.Context, []Column, squirrel.Eq, DTO) error
		FindByWithInnerJoin(context.Context, []Column, Alias, Join, squirrel.Eq, DTO) error

		GetRowsByQuery(ctx context.Context, qb squirrel.SelectBuilder) (*sql.Rows, error)
		CountByQuery(ctx context.Context, qb squirrel.SelectBuilder) (uint64, error)
	}

	Column = string

	Argument = interface{}
)

func NewSqlxMapRepo(logger ZapLogger, db SqlxDBConnectorI, tables ...Table) Repositories {
	mapRepo := make(Repositories, len(tables))
	for _, name := range tables {
		mapRepo[name] = newRepository(logger, db, name)
	}
	return mapRepo
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

func (r Repositories) AutoCreate(ctx context.Context, obj DTOID) (ID, error) {
	return r.AutoRepo(obj).Create(ctx, obj)
}

func (r Repositories) AutoUpdate(ctx context.Context, obj DTOID) (int64, error) {
	return r.AutoRepo(obj).Update(ctx, obj.Id(), obj)
}

func (r Repositories) AutoGet(ctx context.Context, obj DTOID) error {
	return r.AutoRepo(obj).Get(ctx, obj.Id(), obj)
}

func (r Repositories) AutoDelete(ctx context.Context, obj DTOID) (int64, error) {
	return r.AutoRepo(obj).Delete(ctx, obj.Id())
}
