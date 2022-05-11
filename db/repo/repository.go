package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/imperiuse/golib/db"
	"github.com/imperiuse/golib/db/helper"
	"github.com/imperiuse/golib/db/transaction"
	"github.com/imperiuse/golib/reflect/orm"
)

const (
	RowsAffectedUnknown = 0 // RowsAffectedUnknown - 0
	SerialUnknown       = 0 // SerialUnknown  - 0
)

type (
	repository struct {
		logger db.Logger
		dbConn db.PureSqlxConnection
		phf    db.PlaceholderFormat
		name   db.Table
	}
)

func New(logger db.Logger, db db.PureSqlxConnection, tableName db.Table, phf db.PlaceholderFormat) *repository {
	if phf == nil {
		phf = squirrel.Dollar
	}

	return &repository{
		logger: logger,
		dbConn: db,
		name:   tableName,
		phf:    phf,
	}
}

func (r *repository) loggerFieldRepo() zap.Field {
	return zap.String("repo", r.name)
}

func loggerFieldObj(obj any) zap.Field {
	return zap.Any("obj", obj)
}

func loggerFieldID(id db.ID) zap.Field {
	return zap.Any("id", id)
}

// ConvertIDToInt64 convert ID to int64.
func ConvertIDToInt64(id any) int64 {
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
func ConvertIDToString(id any) string {
	if id == nil {
		return "0"
	}

	return fmt.Sprint(id)
}

func (r *repository) Name() db.Table {
	return r.name
}

func (r *repository) Create(ctx context.Context, obj any) (int64, error) {
	r.logger.Info("[repo.Create]", r.loggerFieldRepo(), loggerFieldObj(obj))

	cols, vals := orm.GetDataForCreate(obj)

	query, args, err := squirrel.
		Insert(r.name).
		Columns(cols...).
		Values(vals...).
		Suffix("RETURNING id"). // todo check this foe MySQL, Sqlite and other db.... I'm not sure ...
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return SerialUnknown, fmt.Errorf("[repo.Create] squirrel: %w", err)
	}

	var lastInsertID = int64(0)

	return lastInsertID, r.create(ctx, query, &lastInsertID, args...)
}

func (r *repository) create(ctx context.Context, query db.Query, lastInsertID any, args ...any) error {
	return transaction.WithTransaction(ctx, nil, r.dbConn, helper.InsertAndGetLastID(ctx, lastInsertID, query, args...))
}

func (r *repository) Get(ctx context.Context, id db.ID, dest any) error {
	r.logger.Info("[repo.Get]", r.loggerFieldRepo(), loggerFieldID(id))

	query, args, err := squirrel.
		Select("*").
		From(r.name).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return fmt.Errorf("[repo.Get] squirrel: %w", err)
	}

	return sqlx.GetContext(ctx, r.dbConn, dest, query, args...)
}

func (r *repository) Update(ctx context.Context, id db.ID, obj any) (int64, error) {
	r.logger.Info("[repo.Update]", r.loggerFieldRepo(), loggerFieldID(id), loggerFieldObj(obj))

	sm := orm.GetDataForUpdate(obj)

	query, args, err := squirrel.
		Update(r.name).
		SetMap(sm).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.Update] squirrel: %w", err)
	}

	res, err := r.dbConn.ExecContext(ctx, query, args...)
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.Update] dbConn.ExecContext: %w", err)
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.Update] res.RowsAffected: %w", err)
	}

	return ra, nil
}

func (r *repository) Delete(ctx context.Context, id db.ID) (int64, error) {
	r.logger.Info("[repo.Delete]", r.loggerFieldRepo(), loggerFieldID(id))

	query, args, err := squirrel.
		Delete(r.name).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.Delete] squirrel: %w", err)
	}

	res, err := r.dbConn.ExecContext(ctx, query, args...)
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.Delete] dbConn.ExecContext: %w", err)
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.Delete] res.RowsAffected: %w", err)
	}

	return ra, nil
}

func (r *repository) Insert(ctx context.Context, columns []string, values []any) (int64, error) {
	r.logger.Info("[repo.Insert]", r.loggerFieldRepo(), zap.Any("columns", columns), zap.Any("values", values))

	query, args, err := squirrel.
		Insert(r.name).
		Columns(columns...).
		Values(values...).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("[repo.Insert] squirrel: %w", err)
	}

	res, err := r.dbConn.ExecContext(ctx, query, args...)
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.Insert] dbConn.ExecContext: %w", err)
	}

	return res.RowsAffected()
}

func (r *repository) UpdateCustom(ctx context.Context, set map[string]any, cond db.Condition) (int64, error) {
	r.logger.Info("[repo.UpdateCustom]", r.loggerFieldRepo(),
		zap.Any("set_map", set), zap.Any("condition", cond))

	query, args, err := squirrel.
		Update(r.name).
		SetMap(set).
		Where(cond).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.UpdateCustom] squirrel: %w", err)
	}

	res, err := r.dbConn.ExecContext(ctx, query, args...)
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.ExecContext] squirrel: %w", err)
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return RowsAffectedUnknown, fmt.Errorf("[repo.ExecContext] RowsAffected: %w", err)
	}

	return ra, nil
}

func (r *repository) FindBy(ctx context.Context, columns []string, condition db.Condition, target any) error {
	r.logger.Info("[repo.FindBy]", r.loggerFieldRepo(),
		zap.Any("columns", columns), zap.Any("condition", condition))

	query, args, err := squirrel.
		Select(columns...).
		From(r.name).
		Where(condition).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return fmt.Errorf("[repo.FindBy] squirrel: %w", err)
	}

	return sqlx.SelectContext(ctx, r.dbConn, target, query, args...)
}

func (r *repository) FindOneBy(ctx context.Context, columns []string, condition db.Condition, target any) error {
	r.logger.Info("[repo.FindOneBy]", r.loggerFieldRepo(),
		zap.Any("columns", columns), zap.Any("condition", condition))

	query, args, err := squirrel.
		Select(columns...).
		From(r.name).
		Where(condition).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return fmt.Errorf("[repo.FindOneBy] squirrel: %w", err)
	}

	return sqlx.GetContext(ctx, r.dbConn, target, query, args...)
}

func (r *repository) FindByWithInnerJoin(
	ctx context.Context,
	columns []string,
	fromWithAlias string,
	join string,
	condition db.Condition,
	target any,
) error {
	r.logger.Info("[repo.FindByWithInnerJoin]", r.loggerFieldRepo(),
		zap.Any("columns", columns),
		zap.Any("join", join),
		zap.Any("condition", condition))

	query, args, err := squirrel.
		Select(columns...).
		From(fromWithAlias).
		InnerJoin(join).
		Where(condition).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return fmt.Errorf("[repo.FindByWithInnerJoin] squirrel: %w", err)
	}

	return sqlx.SelectContext(ctx, r.dbConn, target, query, args...)
}

func (r *repository) FindOneByWithInnerJoin(
	ctx context.Context,
	columns []string,
	fromWithAlias string,
	join string,
	condition db.Condition,
	target any,
) error {
	r.logger.Info("[repo.FindOneByWithInnerJoin]", r.loggerFieldRepo(),
		zap.Any("columns", columns),
		zap.Any("join", join),
		zap.Any("condition", condition))

	query, args, err := squirrel.
		Select(columns...).
		From(fromWithAlias).
		InnerJoin(join).
		Where(condition).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return fmt.Errorf("[repo.FindOneByWithInnerJoin] squirrel: %w", err)
	}

	return sqlx.GetContext(ctx, r.dbConn, target, query, args...)
}

func (r *repository) GetRowsByQuery(ctx context.Context, qb squirrel.SelectBuilder) (*sql.Rows, error) {
	r.logger.Info("[repo.GetRowsByQuery]", r.loggerFieldRepo(), zap.Any("qb", qb))

	query, args, err := qb.
		From(r.name).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("[repo.GetRowsByQuery] squirrel: %w", err)
	}

	return r.dbConn.QueryContext(ctx, query, args...)
}

func (r *repository) CountByQuery(ctx context.Context, qb squirrel.SelectBuilder) (uint64, error) {
	r.logger.Info("[repo.CountByQuery]", r.loggerFieldRepo(), zap.Any("qb", qb))

	query, args, err := qb.
		From(r.name).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("[repo.CountByQuery] squirrel: %w", err)
	}

	counter := uint64(0)

	err = r.dbConn.QueryRowxContext(ctx, query, args...).Scan(&counter)
	if err != nil {
		return counter, fmt.Errorf("[repo.CountByQuery] dbConn.QueryRowxContext: %w", err)
	}

	return counter, nil
}

func (r *repository) Select(ctx context.Context, sb db.SelectBuilder, target any) error {
	r.logger.Info("[repo.Select]", r.loggerFieldRepo(), zap.Any("sb", sb))

	query, args, err := sb.
		From(r.name).
		PlaceholderFormat(r.phf).
		ToSql()
	if err != nil {
		return fmt.Errorf("[repo.Select] squirrel: %w", err)
	}

	return sqlx.SelectContext(ctx, r.dbConn, target, query, args...)
}

func (r *repository) SelectWithPagePagination(
	ctx context.Context,
	selectBuilder squirrel.SelectBuilder,
	params db.PagePaginationParams,
	target any,
) (
	db.PagePaginationResults,
	error,
) {
	r.logger.Info("[repo.SelectWithPagePagination]", r.loggerFieldRepo(), zap.Any("params", params))

	const pageNumberPresent = 1

	paginationResult := db.PagePaginationResults{
		CurrentPageNumber: params.PageNumber,
		NextPageNumber:    0,
		CntPages:          0,
	}

	if params.PageSize == 0 {
		return paginationResult, db.ErrZeroPageSize
	}

	totalCount, err := r.CountByQuery(ctx, squirrel.Select("count(1)"))
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

	if err = sqlx.SelectContext(ctx, r.dbConn, target, query, args...); err != nil {
		return paginationResult, fmt.Errorf("SelectWithPagePagination: sqlx.SelectContext(): %w", err)
	}

	// upd I could do this with generics, see generics.go the same methods
	// todo think here, now this block doesn't execute because var `ok` is always false
	//if psl, ok := target.(*[]any); ok && len(*psl) > 0 {
	//	paginationResult.NextPageNumber = paginationResult.CurrentPageNumber + 1
	//}

	return paginationResult, nil
}

func (r *repository) SelectWithCursorOnPKPagination(
	ctx context.Context,
	selectBuilder squirrel.SelectBuilder,
	params db.CursorPaginationParams,
	target any,
) error {
	r.logger.Info("[repo.SelectWithCursorOnPKPagination]", r.loggerFieldRepo(), zap.Any("params", params))

	if params.Limit == 0 {
		return db.ErrZeroLimitSize
	}

	var (
		wh      squirrel.Sqlizer = squirrel.Gt{"id": params.Cursor}
		orderBy                  = "id ASC"
	)

	if params.DescOrder {
		wh = squirrel.Lt{"id": params.Cursor}
		orderBy = "id DESC"
	}

	query, args, err := selectBuilder.
		From(r.name).
		Where(wh).
		OrderBy(orderBy).
		Limit(params.Limit).
		PlaceholderFormat(r.phf).ToSql()
	if err != nil {
		return fmt.Errorf("SelectWithCursorOnPKPagination: selectBuilder.ToSql(): %w", err)
	}

	if err = sqlx.SelectContext(ctx, r.dbConn, target, query, args...); err != nil {
		return fmt.Errorf("SelectWithCursorOnPKPagination: sqlx.SelectContext(): %w", err)
	}

	return nil
}
