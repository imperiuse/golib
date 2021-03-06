//nolint dupl // todo
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/pkg/errors"

	"github.com/imperiuse/golib/sqlx/helper"

	"go.uber.org/zap"

	"github.com/Masterminds/squirrel"
	"github.com/imperiuse/golib/reflect/orm"
	"github.com/jmoiron/sqlx"
)

const (
	RowsAffectedUnknown = int64(0) // RowsAffectedUnknown - 0
	SerialUnknown       = int64(0) // SerialUnknown  - 0
)

type (
	Query = string // Query - sql query
	Alias = string // Alias - alias of table
	Join  = string // Join  - sql join part

	Table = string // Table - table name

	ID = interface{} // ID - uniq ID
)

type (
	//ZapLogger ZapLogger
	ZapLogger = *zap.Logger

	Condition = squirrel.Sqlizer // squirrel.Eq or squirrel.Gt or squirrel.And and etc

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

// ConvertIDToInt64 convert ID to int64.
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

// ConvertIDToString convert ID to string.
func ConvertIDToString(id interface{}) string {
	if id == nil {
		return "0"
	}

	return fmt.Sprint(id)
}

func (r *repository) PureConnector() SqlxDBConnectorI {
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
		return SerialUnknown, errors.Wrap(err, "[repo.Create] squirrel")
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
		return errors.Wrap(err, "[repo.Get] squirrel")
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
		return RowsAffectedUnknown, errors.Wrap(err, "[repo.Update] squirrel")
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return RowsAffectedUnknown, errors.Wrap(err, "[repo.Update] db.ExecContext")
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return RowsAffectedUnknown, errors.Wrap(err, "[repo.Update] res.RowsAffected")
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
		return RowsAffectedUnknown, errors.Wrap(err, "[repo.Delete] squirrel")
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return RowsAffectedUnknown, errors.Wrap(err, "[repo.Delete] db.ExecContext")
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return RowsAffectedUnknown, errors.Wrap(err, "[repo.Delete] res.RowsAffected")
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
		return 0, errors.Wrap(err, "[repo.Insert] squirrel")
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return RowsAffectedUnknown, errors.Wrap(err, "[repo.Insert] db.ExecContext")
	}

	return res.RowsAffected()
}

func (r *repository) UpdateCustom(ctx context.Context, set map[string]interface{}, cond Condition) (int64, error) {
	r.logger.Info("[repo.UpdateCustom]", r.zapFieldRepo(),
		zap.Any("set_map", set), zap.Any("condition", cond))

	query, args, err := squirrel.Update(r.name).
		SetMap(set).
		Where(cond).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return RowsAffectedUnknown, errors.Wrap(err, "[repo.UpdateCustom] squirrel")
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return RowsAffectedUnknown, errors.Wrap(err, "[repo.ExecContext] squirrel")
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return RowsAffectedUnknown, errors.Wrap(err, "[repo.ExecContext] RowsAffected")
	}

	return ra, nil
}

func (r *repository) FindBy(ctx context.Context, columns []string, condition Condition, target interface{}) error {
	r.logger.Info("[repo.FindBy]", r.zapFieldRepo(),
		zap.Any("columns", columns), zap.Any("condition", condition))

	query, args, err := squirrel.Select(columns...).
		From(r.name).
		Where(condition).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "[repo.FindBy] squirrel")
	}

	return sqlx.SelectContext(ctx, r.db, target, query, args...)
}

func (r *repository) FindOneBy(ctx context.Context, columns []string, condition Condition, target interface{}) error {
	r.logger.Info("[repo.FindOneBy]", r.zapFieldRepo(),
		zap.Any("columns", columns), zap.Any("condition", condition))

	query, args, err := squirrel.Select(columns...).
		From(r.name).
		Where(condition).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "[repo.FindOneBy] squirrel")
	}

	return sqlx.GetContext(ctx, r.db, target, query, args...)
}

func (r *repository) FindByWithInnerJoin(
	ctx context.Context,
	columns []string,
	fromWithAlias string,
	join string,
	condition Condition,
	target interface{},
) error {
	r.logger.Info("[repo.FindByWithInnerJoin]", r.zapFieldRepo(),
		zap.Any("columns", columns),
		zap.Any("join", join),
		zap.Any("condition", condition))

	query, args, err := squirrel.Select(columns...).
		From(fromWithAlias).
		InnerJoin(join).
		Where(condition).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "[repo.FindByWithInnerJoin] squirrel")
	}

	return sqlx.SelectContext(ctx, r.db, target, query, args...)
}

func (r *repository) FindOneByWithInnerJoin(
	ctx context.Context,
	columns []string,
	fromWithAlias string,
	join string,
	condition Condition,
	target interface{},
) error {
	r.logger.Info("[repo.FindOneByWithInnerJoin]", r.zapFieldRepo(),
		zap.Any("columns", columns),
		zap.Any("join", join),
		zap.Any("condition", condition))

	query, args, err := squirrel.Select(columns...).
		From(fromWithAlias).
		InnerJoin(join).
		Where(condition).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "[repo.FindOneByWithInnerJoin] squirrel")
	}

	return sqlx.GetContext(ctx, r.db, target, query, args...)
}

func (r *repository) GetRowsByQuery(ctx context.Context, qb squirrel.SelectBuilder) (*sql.Rows, error) {
	r.logger.Info("[repo.GetRowsByQuery]", r.zapFieldRepo(), zap.Any("qb", qb))

	query, args, err := qb.
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[repo.GetRowsByQuery] squirrel")
	}

	return r.db.QueryContext(ctx, query, args...)
}

func (r *repository) CountByQuery(ctx context.Context, qb squirrel.SelectBuilder) (uint64, error) {
	r.logger.Info("[repo.CountByQuery]", r.zapFieldRepo(), zap.Any("qb", qb))

	query, args, err := qb.
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "[repo.CountByQuery] squirrel")
	}

	counter := uint64(0)

	err = r.db.QueryRowxContext(ctx, query, args...).Scan(&counter)
	if err != nil {
		return counter, errors.Wrap(err, "[repo.CountByQuery] db.QueryRowxContext")
	}

	return counter, nil
}

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
)

func (r *repository) SelectWithPagePagination(
	ctx context.Context,
	selectBuilder squirrel.SelectBuilder,
	params PagePaginationParams,
	target interface{},
) (
	PagePaginationResults,
	error,
) {
	r.logger.Info("[repo.SelectWithPagePagination]", r.zapFieldRepo(), zap.Any("params", params))

	const pageNumberPresent = 1

	paginationResult := PagePaginationResults{
		CurrentPageNumber: params.PageNumber,
		NextPageNumber:    0,
		CntPages:          0,
	}

	if params.PageSize == 0 {
		return paginationResult, errors.New("zero value of params.PageSize")
	}

	totalCount, err := r.CountByQuery(ctx, squirrel.Select("count(1)").From(r.name))
	if err != nil {
		return paginationResult, errors.Wrap(err, "SelectWithPagePagination: r.CountByQuery")
	}

	if paginationResult.CntPages = totalCount / params.PageSize; totalCount%params.PageSize != 0 {
		paginationResult.CntPages++
	}

	selectBuilder = selectBuilder.Limit(params.PageSize)
	if params.PageNumber > pageNumberPresent {
		selectBuilder = selectBuilder.Offset((params.PageNumber - 1) * params.PageSize)
	}

	query, args, err := selectBuilder.ToSql()
	if err != nil {
		return paginationResult, errors.Wrap(err, "SelectWithPagePagination: selectBuilder.ToSql()")
	}

	if err = sqlx.SelectContext(ctx, r.db, target, query, args...); err != nil {
		return paginationResult, errors.Wrap(err, "SelectWithPagePagination: sqlx.SelectContext()")
	}

	// todo think here
	//if psl, ok := target.(*[]interface{}); ok && len(*psl) > 0 {
	//	paginationResult.NextPageNumber = paginationResult.CurrentPageNumber + 1
	//}

	return paginationResult, nil

}
