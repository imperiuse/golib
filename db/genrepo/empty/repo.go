package empty

import (
	"context"
	"database/sql"

	"github.com/imperiuse/golib/db"
)

const nameGen = "emptyGenRepo"

type gRepo[I db.ID, D db.GDTO[I]] struct{}

func NewGen[I db.ID, D db.GDTO[I]]() db.GRepository[I, D] {
	return &gRepo[I, D]{}
}

func (g *gRepo[I, D]) Name() db.Table {
	return nameGen
}

func (g *gRepo[I, D]) GetRowsByQuery(context.Context, db.SelectBuilder) (*sql.Rows, error) {
	return nil, db.ErrInvalidRepoEmptyRepo
}

func (g *gRepo[I, D]) CountByQuery(context.Context, db.SelectBuilder) (uint64, error) {
	return 0, db.ErrInvalidRepoEmptyRepo
}

func (g *gRepo[I, D]) Insert(context.Context, []db.Column, []db.Argument) (int64, error) {
	return 0, db.ErrInvalidRepoEmptyRepo
}

func (g *gRepo[I, D]) UpdateCustom(context.Context, map[string]any, db.Condition) (int64, error) {
	return 0, db.ErrInvalidRepoEmptyRepo
}

func (g *gRepo[I, D]) Create(context.Context, D) (I, error) {
	return *new(I), db.ErrInvalidRepoEmptyRepo
}

func (g *gRepo[I, D]) Get(context.Context, I) (D, error) {
	return *new(D), db.ErrInvalidRepoEmptyRepo
}

func (g *gRepo[I, D]) Update(context.Context, I, D) (int64, error) {
	return 0, db.ErrInvalidRepoEmptyRepo
}

func (g *gRepo[I, D]) Delete(context.Context, I) (int64, error) {
	return 0, db.ErrInvalidRepoEmptyRepo
}

func (g *gRepo[I, D]) FindBy(context.Context, []db.Column, db.Condition) ([]D, error) {
	return *new([]D), db.ErrInvalidRepoEmptyRepo
}

func (g *gRepo[I, D]) FindOneBy(context.Context, []db.Column, db.Condition) (D, error) {
	return *new(D), db.ErrInvalidRepoEmptyRepo
}

func (g *gRepo[I, D]) FindByWithInnerJoin(context.Context, []db.Column, db.Alias, db.Join, db.Condition) ([]D, error) {
	return *new([]D), db.ErrInvalidRepoEmptyRepo
}

func (g *gRepo[I, D]) FindOneByWithInnerJoin(context.Context, []db.Column, db.Alias, db.Join, db.Condition) (D, error) {
	return *new(D), db.ErrInvalidRepoEmptyRepo
}

func (g *gRepo[I, D]) Select(context.Context, db.SelectBuilder) ([]D, error) {
	return *new([]D), db.ErrInvalidRepoEmptyRepo
}

func (g *gRepo[I, D]) SelectWithPagePagination(context.Context, db.SelectBuilder, db.PagePaginationParams) ([]D, db.PagePaginationResults, error) {
	return *new([]D), db.PagePaginationResults{}, db.ErrInvalidRepoEmptyRepo
}

func (g *gRepo[I, D]) SelectWithCursorOnPKPagination(context.Context, db.SelectBuilder, db.CursorPaginationParams) ([]D, error) {
	return *new([]D), db.ErrInvalidRepoEmptyRepo
}
