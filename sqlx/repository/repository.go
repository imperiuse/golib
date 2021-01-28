package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/imperiuse/golib/sqlx/helper"

	"go.uber.org/zap"

	"github.com/Masterminds/squirrel"
	"github.com/imperiuse/golib/reflect/orm"
	"github.com/jmoiron/sqlx"
)

const (
	RowsAffectedUnknown = int64(0)
	SerialUnknown       = int64(0)
)

type (
	Query = string
	Alias = string
	Join  = string

	Table = string

	ID = interface{}
)

type (
	ZapLogger = *zap.Logger

	repository struct {
		logger ZapLogger
		db     SqlxDBConnectorI
		name   Repo
	}
)

func newRepository(logger ZapLogger, db SqlxDBConnectorI, tableName Table) *repository {
	return &repository{
		logger: logger,
		db:     db,
		name:   tableName,
	}
}

func (r *repository) zapFieldRepo() zap.Field {
	return zap.String("repo", r.name)
}

func zapFieldObj(obj DTO) zap.Field {
	return zap.Any("obj", obj)
}

func zapFieldID(id ID) zap.Field {
	return zap.Any("id", id)
}

func ConvertIDToInt64(id interface{}) int64 {
	if id == nil {
		return int64(0)
	}

	temp, err := strconv.Atoi(fmt.Sprint(id))
	if err != nil {
		return int64(0)
	}

	return int64(temp)
}

func ConvertIDToString(id interface{}) string {
	if id == nil {
		return "0"
	}
	return fmt.Sprint(id)
}

func (r *repository) SqlxDBConnectorI() SqlxDBConnectorI {
	return r.db
}

func (r *repository) Create(ctx context.Context, obj DTO) (ID, error) {
	r.logger.Info("[repo.Create]", r.zapFieldRepo(), zapFieldObj(obj))

	cols, vals := orm.GetDataForCreate(obj)

	query, args, err := squirrel.Insert(r.name).
		Columns(cols...).
		Values(vals...).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return SerialUnknown, err
	}

	var lastInsertID ID = int64(0)
	return lastInsertID, r.create(ctx, query, &lastInsertID, args...)
}

func (r *repository) create(ctx context.Context, query Query, lastInsertID *ID, args ...interface{}) error {
	return helper.WithTransaction(ctx, nil, r.db, helper.InsertAndGetLastID(ctx, lastInsertID, query, args...))
}

func (r *repository) Get(ctx context.Context, id ID, dest DTO) error {
	r.logger.Info("[repo.Get]", r.zapFieldRepo(), zapFieldID(id))
	query, args, err := squirrel.Select("*").
		From(r.name).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	return sqlx.GetContext(ctx, r.db, dest, query, args...)
}

func (r *repository) Update(ctx context.Context, id ID, obj DTO) (int64, error) {
	r.logger.Info("[repo.Update]", r.zapFieldRepo(), zapFieldID(id), zapFieldObj(obj))

	sm := orm.GetDataForUpdate(obj)

	query, args, err := squirrel.Update(r.name).
		SetMap(sm).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return RowsAffectedUnknown, err
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return RowsAffectedUnknown, err
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return RowsAffectedUnknown, err
	}

	return ra, nil
}

func (r *repository) Delete(ctx context.Context, id ID) (int64, error) {
	r.logger.Info("[repo.Delete]", r.zapFieldRepo(), zapFieldID(id))

	query, args, err := squirrel.Delete(r.name).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return RowsAffectedUnknown, err
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return RowsAffectedUnknown, err
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return RowsAffectedUnknown, err
	}

	return ra, nil
}

func (r *repository) Insert(ctx context.Context, columns []string, values []interface{}) (int64, error) {
	r.logger.Info("[repo.Insert]", r.zapFieldRepo(), zap.Any("columns", columns), zap.Any("values", values))

	query, args, err := squirrel.Insert(r.name).
		Columns(columns...).
		Values(values...).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return 0, err
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return RowsAffectedUnknown, err
	}

	return res.RowsAffected()
}

func (r *repository) UpdateCustom(ctx context.Context, set map[string]interface{}, condition squirrel.Eq) (int64, error) {
	r.logger.Info("[repo.UpdateCustom]", r.zapFieldRepo(), zap.Any("set_map", set), zap.Any("condition", condition))

	query, args, err := squirrel.Update(r.name).
		SetMap(set).
		Where(condition).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return RowsAffectedUnknown, err
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return RowsAffectedUnknown, err
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return RowsAffectedUnknown, err
	}

	return ra, nil
}

func (r *repository) FindBy(ctx context.Context, columns []string, condition squirrel.Eq, target interface{}) error {
	r.logger.Info("[repo.FindBy]", r.zapFieldRepo(), zap.Any("columns", columns), zap.Any("condition", condition))

	query, args, err := squirrel.Select(columns...).
		From(r.name).
		Where(condition).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	return sqlx.SelectContext(ctx, r.db, target, query, args...)
}

func (r *repository) FindByWithInnerJoin(ctx context.Context, columns []string, alias string, join string, condition squirrel.Eq, target interface{}) error {
	r.logger.Info("[repo.FindByWithInnerJoin]", r.zapFieldRepo(),
		zap.Any("columns", columns),
		zap.Any("join", join),
		zap.Any("condition", condition))

	query, args, err := squirrel.Select(columns...).
		From(r.name + " as " + alias).
		InnerJoin(join).
		Where(condition).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	return sqlx.GetContext(ctx, r.db, target, query, args...)
}

func (r *repository) GetRowsByQuery(ctx context.Context, qb squirrel.SelectBuilder) (*sql.Rows, error) {
	r.logger.Info("[repo.GetRowsByQuery]", r.zapFieldRepo(), zap.Any("qb", qb))

	query, args, err := qb.
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	return r.db.QueryContext(ctx, query, args...)
}

func (r *repository) CountByQuery(ctx context.Context, qb squirrel.SelectBuilder) (uint64, error) {
	r.logger.Info("[repo.CountByQuery]", r.zapFieldRepo(), zap.Any("qb", qb))

	query, args, err := qb.
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return 0, err
	}

	counter := uint64(0)
	err = r.db.QueryRowxContext(ctx, query, args...).Scan(&counter)

	return counter, err
}
