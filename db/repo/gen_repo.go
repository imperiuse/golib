package repo

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"

	"github.com/imperiuse/golib/db"
	"github.com/imperiuse/golib/db/genrepo/empty"
	"github.com/imperiuse/golib/reflect/orm"
)

type (
	gRepository[I db.ID, D db.DTO] struct {
		repository
	}
)

func NewGen[I db.ID, D db.GDTO[I], C db.Config](connector db.Connector[C]) db.GRepository[I, D] {
	var dto D

	cfg := connector.Config()

	if cfg.IsEnableValidationRepoNames() && !connector.IsAllowRepo(dto.Repo()) {
		return empty.NewGen[I, D]()
	}

	phf := cfg.PlaceholderFormat()
	if phf == nil {
		phf = squirrel.Dollar
	}

	return &gRepository[I, D]{
		repository{
			logger: connector.Logger(),
			dbConn: connector.Connection(),
			phf:    phf,
			name:   dto.Repo(),
		},
	}
}

func (g *gRepository[I, D]) Create(ctx context.Context, d D) (I, error) {
	var lastInsertID I

	g.logger.Info("[repo.Create]", g.loggerFieldRepo(), loggerFieldObj(d))

	cols, vals := orm.GetDataForCreate(d)

	query, args, err := squirrel.Insert(g.name).
		Columns(cols...).
		Values(vals...).
		Suffix("RETURNING id").
		PlaceholderFormat(g.phf).
		ToSql()
	if err != nil {
		return lastInsertID, fmt.Errorf("[repo.Create] squirrel: %w", err)
	}

	return lastInsertID, g.create(ctx, query, &lastInsertID, args...)
}

func (g *gRepository[I, D]) Get(ctx context.Context, id I) (D, error) {
	var dto D
	err := g.repository.Get(ctx, id, &dto)

	return dto, err
}

func (g *gRepository[I, D]) Update(ctx context.Context, id I, d D) (int64, error) {
	return g.repository.Update(ctx, id, d)
}

func (g *gRepository[I, D]) Delete(ctx context.Context, id I) (int64, error) {
	return g.repository.Delete(ctx, id)
}

func (g *gRepository[I, D]) FindBy(ctx context.Context, columns []db.Column, condition db.Condition) ([]D, error) {
	var dtos = make([]D, 0)
	err := g.repository.FindBy(ctx, columns, condition, &dtos)

	return dtos, err
}

func (g *gRepository[I, D]) FindOneBy(ctx context.Context, columns []db.Column, condition db.Condition) (D, error) {
	var dto D
	err := g.repository.FindOneBy(ctx, columns, condition, &dto)

	return dto, err
}

func (g *gRepository[I, D]) FindByWithInnerJoin(
	ctx context.Context, columns []db.Column, alias db.Alias, join db.Join, condition db.Condition,
) ([]D, error) {
	var dtos = make([]D, 0)
	err := g.repository.FindByWithInnerJoin(ctx, columns, alias, join, condition, &dtos)

	return dtos, err
}

func (g *gRepository[I, D]) FindOneByWithInnerJoin(
	ctx context.Context, columns []db.Column, alias db.Alias, join db.Join, condition db.Condition,
) (D, error) {
	var dtos = make([]D, 0, 1)
	err := g.repository.FindByWithInnerJoin(ctx, columns, alias, join, condition, &dtos)

	if len(dtos) == 1 {
		return dtos[0], err
	}

	return *new(D), fmt.Errorf("%v, %w", db.ErrMismatchRowsCnt, err)

}

func (g *gRepository[I, D]) Select(ctx context.Context, builder db.SelectBuilder) ([]D, error) {
	var dtos = make([]D, 0)
	err := g.repository.Select(ctx, builder, &dtos)

	return dtos, err
}

func (g *gRepository[I, D]) SelectWithCursorOnPKPagination(
	ctx context.Context, builder db.SelectBuilder, params db.CursorPaginationParams,
) ([]D, error) {
	var dtos = make([]D, 0)
	err := g.repository.SelectWithCursorOnPKPagination(ctx, builder, params, &dtos)

	return dtos, err
}

func (g *gRepository[I, D]) SelectWithPagePagination(
	ctx context.Context, builder db.SelectBuilder, params db.PagePaginationParams,
) ([]D, db.PagePaginationResults, error) {
	var dtos = make([]D, 0)
	pr, err := g.repository.SelectWithPagePagination(ctx, builder, params, &dtos)

	if len(dtos) > 0 {
		pr.NextPageNumber = pr.CurrentPageNumber + 1
	}

	return dtos, pr, err
}
