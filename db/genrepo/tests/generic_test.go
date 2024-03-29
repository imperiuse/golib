package tests

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"go.uber.org/zap"

	"github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"

	"github.com/imperiuse/golib/db"
	"github.com/imperiuse/golib/db/connector"
	"github.com/imperiuse/golib/db/example/simple/config"
	"github.com/imperiuse/golib/db/example/simple/dto"
	"github.com/imperiuse/golib/db/genrepo/emptygen"
	"github.com/imperiuse/golib/db/mocks"
	"github.com/imperiuse/golib/db/repo"
	"github.com/imperiuse/golib/reflect/orm"
)

type badConfig struct{}

func (_ badConfig) PlaceholderFormat() db.PlaceholderFormat {
	return nil
}

func (_ badConfig) IsEnableValidationRepoNames() bool {
	return false
}

func (_ badConfig) IsEnableReposCache() bool {
	return false
}

func Test_NewGenRepo(t *testing.T) {
	r := repo.NewGen[dto.ID, dto.User[dto.ID]](
		connector.New[config.SimpleTestConfig](config.SimpleTestConfig{}, zap.NewNop(), mocks.GoodMockDBConn),
	)
	assert.NotNil(t, r)
	assert.Equal(t, dto.User[dto.ID]{}.Repo(), r.Name())

	r = repo.NewGen[dto.ID, dto.User[dto.ID]](
		connector.New[badConfig](badConfig{}, zap.NewNop(), mocks.GoodMockDBConn),
	)
	assert.NotNil(t, r)
	assert.Equal(t, dto.User[dto.ID]{}.Repo(), r.Name())
}

func Test_NewGenRepo_NotValid(t *testing.T) {
	c := connector.New[config.SimpleTestConfig](config.New(nil, true, false),
		zap.NewNop(), mocks.GoodMockDBConn)

	r := repo.NewGen[dto.ID, dto.User[dto.ID]](c)
	assert.NotNil(t, r)
	assert.Equal(t, emptygen.NewGen[dto.ID, dto.User[dto.ID]](), r)

	c.AddAllowsRepos(dto.User[dto.ID]{}.Repo())

	r = repo.NewGen[dto.ID, dto.User[dto.ID]](c)
	assert.NotNil(t, r)
	assert.Equal(t, dto.User[dto.ID]{}.Repo(), r.Name())
}

func Test_NewGenMethods(t *testing.T) {
	ctx := context.Background()

	r := repo.NewGen[dto.ID, dto.User[dto.ID]](
		connector.New[config.SimpleTestConfig](config.SimpleTestConfig{}, zap.NewNop(), mocks.GoodMockDBConn),
	)
	assert.Equal(t, dto.User[dto.ID]{}.Repo(), r.Name())

	u, err := r.Get(ctx, 1)
	assert.NotNil(t, u)
	assert.Equal(t, sql.ErrNoRows, err)

	n, err := r.Update(ctx, 1, u)
	assert.Equal(t, int64(0), n)
	assert.Nil(t, err)

	n, err = r.Delete(ctx, 1)
	assert.Equal(t, int64(0), n)
	assert.Nil(t, err)

	id, err := r.Create(ctx, u)
	assert.Equal(t, 0, id)
	assert.Nil(t, err)

	cols, jc := orm.GetDataForSelect(u)
	al := orm.GetTableAlias(u)

	u, err = r.FindOneBy(ctx, cols, squirrel.Eq{"id": 1})
	assert.NotNil(t, u)
	assert.Equal(t, sql.ErrNoRows, err)

	_, _ = al, jc

	cnt, err := r.CountByQuery(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"))
	assert.EqualValues(t, int64(0), cnt)
	assert.Equal(t, sql.ErrNoRows, errors.Unwrap(err))

	rows, err := r.GetRowsByQuery(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"))
	assert.NotNil(t, rows)
	assert.Nil(t, err)

	n, err = r.Insert(ctx, []string{"name"}, []any{"test"})
	assert.Equal(t, int64(0), n)
	assert.Nil(t, err)

	n, err = r.UpdateCustom(ctx, map[string]any{"cols": 123}, squirrel.Eq{"id": 1})
	assert.Equal(t, int64(0), n)
	assert.Nil(t, err)

	// todo think how to test without panic issues (problem we need fake Rows structure.... it's too complex...
	//assert.Equal(t, sql.ErrNoRows, r.FindBy(ctx, cols, squirrel.Eq{"id": 1}, &users))
	//assert.Equal(t, sql.ErrNoRows, r.Select(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), &users))
	//assert.Equal(t, nil, r.SelectWithCursorOnPKPagination(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), db.CursorPaginationParams{Limit: 10, Cursor: 1}, &users))

	ul, err := r.SelectWithCursorOnPKPagination(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), db.CursorPaginationParams{})
	assert.NotNil(t, ul)
	assert.Equal(t, db.ErrZeroLimitSize, err)

	ul, pr, err := r.SelectWithPagePagination(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), db.PagePaginationParams{})
	assert.NotNil(t, pr)
	assert.NotNil(t, ul)
	assert.Equal(t, db.ErrZeroPageSize, err)

	ul, pr, err = r.SelectWithPagePagination(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), db.PagePaginationParams{PageSize: 10, PageNumber: 1})
	assert.NotNil(t, pr)
	assert.NotNil(t, ul)
	assert.Equal(t, sql.ErrNoRows, errors.Unwrap(errors.Unwrap(err)))
}

func Test_NewGenMethods_Negative(t *testing.T) {
	ctx := context.Background()

	r := repo.NewGen[dto.ID, dto.User[dto.ID]](
		connector.New[config.SimpleTestConfig](config.SimpleTestConfig{}, zap.NewNop(), mocks.BadMockDBConn),
	)
	assert.Equal(t, dto.User[dto.ID]{}.Repo(), r.Name())

	u, err := r.Get(ctx, 1)
	assert.NotNil(t, u)
	assert.Equal(t, sql.ErrNoRows, err)

	n, err := r.Update(ctx, 1, u)
	assert.Equal(t, int64(0), n)
	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, errors.Unwrap(err))

	n, err = r.Delete(ctx, 1)
	assert.Equal(t, int64(0), n)
	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, errors.Unwrap(err))

	id, err := r.Create(ctx, u)
	assert.Equal(t, 0, id)
	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, errors.Unwrap(err))

	cols, jc := orm.GetDataForSelect(u)
	al := orm.GetTableAlias(u)

	u, err = r.FindOneBy(ctx, cols, squirrel.Eq{"id": 1})
	assert.NotNil(t, u)
	assert.Equal(t, sql.ErrNoRows, err)

	_, _ = al, jc

	cnt, err := r.CountByQuery(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"))
	assert.EqualValues(t, int64(0), cnt)
	assert.Equal(t, sql.ErrNoRows, errors.Unwrap(err))

	rows, err := r.GetRowsByQuery(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"))
	assert.NotNil(t, rows)
	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, err)

	n, err = r.Insert(ctx, []string{"name"}, []any{"test"})
	assert.Equal(t, int64(0), n)
	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, errors.Unwrap(err))

	n, err = r.UpdateCustom(ctx, map[string]any{"cols": 123}, squirrel.Eq{"id": 1})
	assert.Equal(t, int64(0), n)
	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, errors.Unwrap(err))

	// todo think how to test without panic issues (problem we need fake Rows structure.... it's too complex...
	//assert.Equal(t, sql.ErrNoRows, r.FindBy(ctx, cols, squirrel.Eq{"id": 1}, &users))
	//assert.Equal(t, sql.ErrNoRows, r.Select(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), &users))
	//assert.Equal(t, nil, r.SelectWithCursorOnPKPagination(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), db.CursorPaginationParams{Limit: 10, Cursor: 1}, &users))

	ul, err := r.SelectWithCursorOnPKPagination(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), db.CursorPaginationParams{})
	assert.NotNil(t, ul)
	assert.Equal(t, db.ErrZeroLimitSize, err)

	ul, pr, err := r.SelectWithPagePagination(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), db.PagePaginationParams{})
	assert.NotNil(t, pr)
	assert.NotNil(t, ul)
	assert.Equal(t, db.ErrZeroPageSize, err)

	ul, pr, err = r.SelectWithPagePagination(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), db.PagePaginationParams{PageSize: 10, PageNumber: 1})
	assert.NotNil(t, pr)
	assert.NotNil(t, ul)
	assert.Equal(t, sql.ErrNoRows, errors.Unwrap(errors.Unwrap(err)))
}
