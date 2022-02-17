//nolint dupl // todo
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/imperiuse/golib/reflect/orm"
	"github.com/imperiuse/golib/sqlx/helper"
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

var (
	ErrZeroPageSize  = errors.New("zero value of params.PageSize")
	ErrZeroLimitSize = errors.New("zero value of params.Limit")
)

type (
	//ZapLogger ZapLogger
	ZapLogger = *zap.Logger

	// Condition - squirrel.Sqlizer
	Condition = squirrel.Sqlizer // squirrel.Eq or squirrel.Gt or squirrel.And and etc

	// SelectBuilder - squirrel.SelectBuilder
	SelectBuilder = squirrel.SelectBuilder

	// PlaceholderFormat - squirrel.PlaceholderFormat
	PlaceholderFormat = squirrel.PlaceholderFormat

	repository struct {
		logger ZapLogger
		db     SqlxDBConnectorI
		name   Repo
		phf    PlaceholderFormat
	}
)

func newRepository(logger ZapLogger, db SqlxDBConnectorI, tableName Table, phf PlaceholderFormat) *repository {
	if phf == nil {
		phf = squirrel.Dollar
	}

	return &repository{
		logger: logger,
		db:     db,
		name:   tableName,
		phf:    phf,
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

func (r *repository) Name() Table {
	return r.name
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
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return SerialUnknown, fmt.Errorf("[repo.Create] squirrel: %w", err)
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
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return fmt.Errorf("[repo.Get] squirrel: %w", err)
	}

	return sqlx.GetContext(ctx, r.db, dest, query, args...)
}

func (r *repository) Update(ctx context.Context, id ID, obj DTO) (int64, error) {
	r.logger.Info("[repo.Update]", r.zapFieldRepo(), zapFieldID(id), zapFieldObj(obj))

	sm := orm.GetDataForUpdate(obj)

	query, args, err := squirrel.Update(r.name).
		SetMap(sm).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.Update] squirrel: %w", err)
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.Update] db.ExecContext: %w", err)
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.Update] res.RowsAffected: %w", err)
	}

	return ra, nil
}

func (r *repository) Delete(ctx context.Context, id ID) (int64, error) {
	r.logger.Info("[repo.Delete]", r.zapFieldRepo(), zapFieldID(id))

	query, args, err := squirrel.Delete(r.name).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.Delete] squirrel: %w", err)
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.Delete] db.ExecContext: %w", err)
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.Delete] res.RowsAffected: %w", err)
	}

	return ra, nil
}

func (r *repository) Insert(ctx context.Context, columns []string, values []interface{}) (int64, error) {
	r.logger.Info("[repo.Insert]", r.zapFieldRepo(), zap.Any("columns", columns), zap.Any("values", values))

	query, args, err := squirrel.Insert(r.name).
		Columns(columns...).
		Values(values...).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("[repo.Insert] squirrel: %w", err)
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.Insert] db.ExecContext: %w", err)
	}

	return res.RowsAffected()
}

func (r *repository) UpdateCustom(ctx context.Context, set map[string]interface{}, cond Condition) (int64, error) {
	r.logger.Info("[repo.UpdateCustom]", r.zapFieldRepo(),
		zap.Any("set_map", set), zap.Any("condition", cond))

	query, args, err := squirrel.Update(r.name).
		SetMap(set).
		Where(cond).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.UpdateCustom] squirrel: %w", err)
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.ExecContext] squirrel: %w", err)
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.ExecContext] RowsAffected: %w", err)
	}

	return ra, nil
}

func (r *repository) FindBy(ctx context.Context, columns []string, condition Condition, target DTO) error {
	r.logger.Info("[repo.FindBy]", r.zapFieldRepo(),
		zap.Any("columns", columns), zap.Any("condition", condition))

	query, args, err := squirrel.Select(columns...).
		From(r.name).
		Where(condition).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return fmt.Errorf("[repo.FindBy] squirrel: %w", err)
	}

	return sqlx.SelectContext(ctx, r.db, target, query, args...)
}

func (r *repository) FindOneBy(ctx context.Context, columns []string, condition Condition, target DTO) error {
	r.logger.Info("[repo.FindOneBy]", r.zapFieldRepo(),
		zap.Any("columns", columns), zap.Any("condition", condition))

	query, args, err := squirrel.Select(columns...).
		From(r.name).
		Where(condition).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return fmt.Errorf("[repo.FindOneBy] squirrel: %w", err)
	}

	return sqlx.GetContext(ctx, r.db, target, query, args...)
}

func (r *repository) FindByWithInnerJoin(
	ctx context.Context,
	columns []string,
	fromWithAlias string,
	join string,
	condition Condition,
	target DTO,
) error {
	r.logger.Info("[repo.FindByWithInnerJoin]", r.zapFieldRepo(),
		zap.Any("columns", columns),
		zap.Any("join", join),
		zap.Any("condition", condition))

	query, args, err := squirrel.Select(columns...).
		From(fromWithAlias).
		InnerJoin(join).
		Where(condition).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return fmt.Errorf("[repo.FindByWithInnerJoin] squirrel: %w", err)
	}

	return sqlx.SelectContext(ctx, r.db, target, query, args...)
}

func (r *repository) FindOneByWithInnerJoin(
	ctx context.Context,
	columns []string,
	fromWithAlias string,
	join string,
	condition Condition,
	target DTO,
) error {
	r.logger.Info("[repo.FindOneByWithInnerJoin]", r.zapFieldRepo(),
		zap.Any("columns", columns),
		zap.Any("join", join),
		zap.Any("condition", condition))

	query, args, err := squirrel.Select(columns...).
		From(fromWithAlias).
		InnerJoin(join).
		Where(condition).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return fmt.Errorf("[repo.FindOneByWithInnerJoin] squirrel: %w", err)
	}

	return sqlx.GetContext(ctx, r.db, target, query, args...)
}

func (r *repository) GetRowsByQuery(ctx context.Context, qb squirrel.SelectBuilder) (*sql.Rows, error) {
	r.logger.Info("[repo.GetRowsByQuery]", r.zapFieldRepo(), zap.Any("qb", qb))

	query, args, err := qb.
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("[repo.GetRowsByQuery] squirrel: %w", err)
	}

	return r.db.QueryContext(ctx, query, args...)
}

func (r *repository) CountByQuery(ctx context.Context, qb squirrel.SelectBuilder) (uint64, error) {
	r.logger.Info("[repo.CountByQuery]", r.zapFieldRepo(), zap.Any("qb", qb))

	query, args, err := qb.
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("[repo.CountByQuery] squirrel: %w", err)
	}

	counter := uint64(0)

	err = r.db.QueryRowxContext(ctx, query, args...).Scan(&counter)
	if err != nil {
		return counter, fmt.Errorf("[repo.CountByQuery] db.QueryRowxContext: %w", err)
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

	CursorPaginationParams struct {
		Limit     uint64
		Cursor    uint64
		DescOrder bool
	}
)

func (r *repository) Select(ctx context.Context, sb SelectBuilder, target DTO) error {
	r.logger.Info("[repo.Select]", r.zapFieldRepo(), zap.Any("sb", sb))

	query, args, err := sb.
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return fmt.Errorf("[repo.Select] squirrel: %w", err)
	}

	return sqlx.SelectContext(ctx, r.db, target, query, args...)
}

func (r *repository) SelectWithPagePagination(
	ctx context.Context,
	selectBuilder squirrel.SelectBuilder,
	params PagePaginationParams,
	target DTO,
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
		return paginationResult, ErrZeroPageSize
	}

	totalCount, err := r.CountByQuery(ctx, squirrel.Select("count(1)").From(r.name))
	if err != nil {
		return paginationResult, fmt.Errorf("SelectWithPagePagination: r.CountByQuery: %w", err)
	}

	if paginationResult.CntPages = totalCount / params.PageSize; totalCount%params.PageSize != 0 {
		paginationResult.CntPages++
	}

	selectBuilder = selectBuilder.From(r.name).Limit(params.PageSize)
	if params.PageNumber > pageNumberPresent {
		selectBuilder = selectBuilder.Offset((params.PageNumber - 1) * params.PageSize)
	}

	query, args, err := selectBuilder.PlaceholderFormat(r.phf).ToSql()
	if err != nil {
		return paginationResult, fmt.Errorf("SelectWithPagePagination: selectBuilder.ToSql(): %w", err)
	}

	if err = sqlx.SelectContext(ctx, r.db, target, query, args...); err != nil {
		return paginationResult, fmt.Errorf("SelectWithPagePagination: sqlx.SelectContext(): %w", err)
	}

	// todo think here, now this block doesn't execute because var `ok` is always false
	//if psl, ok := target.(*[]interface{}); ok && len(*psl) > 0 {
	//	paginationResult.NextPageNumber = paginationResult.CurrentPageNumber + 1
	//}

	return paginationResult, nil

}

func (r *repository) SelectWithCursorOnPKPagination(
	ctx context.Context,
	selectBuilder squirrel.SelectBuilder,
	params CursorPaginationParams,
	target DTO,
) error {
	r.logger.Info("[repo.SelectWithCursorOnPKPagination]", r.zapFieldRepo(), zap.Any("params", params))

	if params.Limit == 0 {
		return ErrZeroLimitSize
	}

	var (
		wh      squirrel.Sqlizer = squirrel.Gt{"id": params.Cursor}
		orderBy                  = "id ASC"
	)

	if params.DescOrder {
		wh = squirrel.Lt{"id": params.Cursor}
		orderBy = "id DESC"
	}

	query, args, err := selectBuilder.From(r.name).Where(wh).OrderBy(orderBy).Limit(params.Limit).
		PlaceholderFormat(r.phf).ToSql()
	if err != nil {
		return fmt.Errorf("SelectWithCursorOnPKPagination: selectBuilder.ToSql(): %w", err)
	}

	if err = sqlx.SelectContext(ctx, r.db, target, query, args...); err != nil {
		return fmt.Errorf("SelectWithCursorOnPKPagination: sqlx.SelectContext(): %w", err)
	}

	return nil

}
