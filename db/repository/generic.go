package repository

import (
	"context"
	"github.com/imperiuse/golib/db"
)

type (
	gRepository[D db.DTO] struct {
		repository
	}
)

func NewGen[D db.DTO, C db.Config](connector db.Connector[C]) db.GRepository[D] {
	var dto D
	return &gRepository[D]{
		repository{
			logger: connector.Logger(),
			dbConn: connector.Connection(),
			phf:    connector.Config().PlaceholderFormat(),
			name:   dto.Repo(),
		},
	}
}

func (g *gRepository[D]) Create(ctx context.Context, d D) (db.ID, error) {
	return g.repository.Create(ctx, d)
}

func (g *gRepository[D]) Get(ctx context.Context, id db.ID) (D, error) {
	var dto D
	err := g.repository.Get(ctx, id, &dto)

	return dto, err
}

func (g *gRepository[D]) Update(ctx context.Context, id db.ID, d D) (int64, error) {
	return g.repository.Update(ctx, id, d)
}

func (g *gRepository[D]) FindBy(ctx context.Context, columns []db.Column, condition db.Condition) ([]D, error) {
	var dtos = make([]D, 0)
	err := g.repository.FindBy(ctx, columns, condition, &dtos)

	return dtos, err
}

func (g *gRepository[D]) FindOneBy(ctx context.Context, columns []db.Column, condition db.Condition) (D, error) {
	var dto D
	err := g.repository.FindOneBy(ctx, columns, condition, &dto)

	return dto, err
}

func (g *gRepository[D]) FindByWithInnerJoin(
	ctx context.Context, columns []db.Column, alias db.Alias, join db.Join, condition db.Condition,
) ([]D, error) {
	var dtos = make([]D, 0)
	err := g.repository.FindByWithInnerJoin(ctx, columns, alias, join, condition, &dtos)

	return dtos, err
}

func (g *gRepository[D]) FindOneByWithInnerJoin(
	ctx context.Context, columns []db.Column, alias db.Alias, join db.Join, condition db.Condition,
) (D, error) {
	var dto D
	err := g.repository.FindByWithInnerJoin(ctx, columns, alias, join, condition, &dto)

	return dto, err

}

func (g *gRepository[D]) Select(ctx context.Context, builder db.SelectBuilder) ([]D, error) {
	var dtos = make([]D, 0)
	err := g.repository.Select(ctx, builder, &dtos)

	return dtos, err
}

func (g *gRepository[D]) SelectWithCursorOnPKPagination(
	ctx context.Context, builder db.SelectBuilder, params db.CursorPaginationParams,
) ([]D, error) {
	var dtos = make([]D, 0)
	err := g.repository.SelectWithCursorOnPKPagination(ctx, builder, params, &dtos)

	return dtos, err
}

func (g *gRepository[D]) SelectWithPagePagination(
	ctx context.Context, builder db.SelectBuilder, params db.PagePaginationParams,
) ([]D, db.PagePaginationResults, error) {
	var dtos = make([]D, 0)
	pr, err := g.repository.SelectWithPagePagination(ctx, builder, params, &dtos)

	if len(dtos) > 0 {
		pr.NextPageNumber = pr.CurrentPageNumber + 1
	}

	return dtos, pr, err
}
