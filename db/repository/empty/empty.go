// Package empty used for realization absent Optional[T] pattern
package empty

import (
	"context"
	"database/sql"

	"github.com/imperiuse/golib/db"
)

const name = "emptyRepo"

type repo struct{}

func (r *repo) Name() db.Table {
	return name
}

func (r *repo) GetRowsByQuery(context.Context, db.SelectBuilder) (*sql.Rows, error) {
	return nil, db.ErrInvalidRepo
}

func (r *repo) CountByQuery(context.Context, db.SelectBuilder) (uint64, error) {
	return 0, db.ErrInvalidRepo
}

func (r *repo) Insert(context.Context, []db.Column, []db.Argument) (int64, error) {
	return 0, db.ErrInvalidRepo
}

func (r *repo) UpdateCustom(context.Context, map[string]any, db.Condition) (int64, error) {
	return 0, db.ErrInvalidRepo
}

func (r *repo) Create(context.Context, any) (db.ID, error) {
	return nil, db.ErrInvalidRepo
}

func (r *repo) Get(context.Context, db.ID, any) error {
	return db.ErrInvalidRepo
}

func (r *repo) Update(context.Context, db.ID, any) (int64, error) {
	return 0, db.ErrInvalidRepo
}

func (r *repo) Delete(context.Context, db.ID) (int64, error) {
	return 0, db.ErrInvalidRepo
}

func (r *repo) FindBy(context.Context, []db.Column, db.Condition, any) error {
	return db.ErrInvalidRepo
}

func (r *repo) FindOneBy(context.Context, []db.Column, db.Condition, any) error {
	return db.ErrInvalidRepo
}

func (r *repo) FindByWithInnerJoin(context.Context, []db.Column, db.Alias, db.Join, db.Condition, any) error {
	return db.ErrInvalidRepo
}

func (r *repo) FindOneByWithInnerJoin(context.Context, []db.Column, db.Alias, db.Join, db.Condition, any) error {
	return db.ErrInvalidRepo
}

func (r *repo) Select(context.Context, db.SelectBuilder, any) error {
	return db.ErrInvalidRepo
}

func (r *repo) SelectWithPagePagination(context.Context, db.SelectBuilder, db.PagePaginationParams, any) (db.PagePaginationResults, error) {
	return db.PagePaginationResults{}, db.ErrInvalidRepo
}

func (r *repo) SelectWithCursorOnPKPagination(context.Context, db.SelectBuilder, db.CursorPaginationParams, any) error {
	return db.ErrInvalidRepo
}

func New() db.Repository {
	return &repo{}
}
