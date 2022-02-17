package repo

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"go.uber.org/zap"

	"github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"

	"github.com/imperiuse/golib/db"
	"github.com/imperiuse/golib/db/example/simple/dto"
	"github.com/imperiuse/golib/db/mocks"
	"github.com/imperiuse/golib/reflect/orm"
)

func Test_NewRepo(t *testing.T) {
	logger := zap.NewNop()

	r := New(logger, mocks.GoodMockDBConn, dto.User[dto.ID]{}.Repo(), nil)
	assert.NotNil(t, r)
	assert.Equal(t, logger, r.logger)
	assert.Equal(t, mocks.GoodMockDBConn, r.dbConn)
	assert.Equal(t, squirrel.Dollar, r.phf)

	r = New(logger, mocks.GoodMockDBConn, dto.User[dto.ID]{}.Repo(), squirrel.Dollar)
	assert.NotNil(t, r)
	assert.Equal(t, logger, r.logger)
	assert.Equal(t, mocks.GoodMockDBConn, r.dbConn)
	assert.Equal(t, squirrel.Dollar, r.phf)
}

func Test_RepoMethods(t *testing.T) {
	ctx := context.Background()

	r := New(zap.NewNop(), mocks.GoodMockDBConn, dto.User[dto.ID]{}.Repo(), squirrel.Dollar)
	assert.NotNil(t, r)

	assert.NotNil(t, dto.User[dto.ID]{}.Repo(), r.Name())

	user := dto.User[dto.ID]{BaseDTO: dto.BaseDTO[dto.ID]{Id: 1}}
	assert.Equal(t, sql.ErrNoRows, r.Get(ctx, 1, user))

	n, err := r.Update(ctx, 1, user)
	assert.Equal(t, int64(0), n)
	assert.Nil(t, err)

	n, err = r.Delete(ctx, user)
	assert.Equal(t, int64(0), n)
	assert.Nil(t, err)

	id, err := r.Create(ctx, user)
	assert.Equal(t, int64(0), id)
	assert.Nil(t, err)

	users := []dto.User[dto.ID]{}
	cols, jc := orm.GetDataForSelect(user)
	al := orm.GetTableAlias(user)
	assert.Equal(t, sql.ErrNoRows, r.FindOneBy(ctx, cols, squirrel.Eq{"id": 1}, &user))
	assert.Equal(t, sql.ErrNoRows, r.FindOneByWithInnerJoin(ctx, cols, al, jc, squirrel.Gt{"id": 1}, &users))

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
	//assert.Equal(t, sql.ErrNoRows, r.FindByWithInnerJoin(ctx, cols, al, jc, squirrel.Gt{"id": 1}, &users))
	//assert.Equal(t, sql.ErrNoRows, r.Select(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), &users))
	//assert.Equal(t, nil, r.SelectWithCursorOnPKPagination(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), db.CursorPaginationParams{Limit: 10, Cursor: 1}, &users))

	assert.Equal(t, db.ErrZeroLimitSize, r.SelectWithCursorOnPKPagination(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), db.CursorPaginationParams{}, &users))

	pr, err := r.SelectWithPagePagination(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), db.PagePaginationParams{}, &users)
	assert.NotNil(t, pr)
	assert.Equal(t, db.ErrZeroPageSize, err)

	pr, err = r.SelectWithPagePagination(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), db.PagePaginationParams{PageSize: 10, PageNumber: 1}, &users)
	assert.NotNil(t, pr)
	assert.Equal(t, sql.ErrNoRows, errors.Unwrap(errors.Unwrap(err)))
}

func Test_RepoMethods_Negative(t *testing.T) {
	ctx := context.Background()

	r := New(zap.NewNop(), mocks.BadMockDBConn, dto.User[dto.ID]{}.Repo(), squirrel.Dollar)
	assert.NotNil(t, r)

	assert.NotNil(t, dto.User[dto.ID]{}.Repo(), r.Name())

	user := dto.User[dto.ID]{BaseDTO: dto.BaseDTO[dto.ID]{Id: 1}}
	assert.Equal(t, sql.ErrNoRows, r.Get(ctx, 1, user))

	n, err := r.Update(ctx, 1, user)
	assert.Equal(t, int64(0), n)
	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, errors.Unwrap(err))

	n, err = r.Delete(ctx, user)
	assert.Equal(t, int64(0), n)
	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, errors.Unwrap(err))

	id, err := r.Create(ctx, user)
	assert.Equal(t, int64(0), id)
	assert.Error(t, db.ErrInvalidRepoEmptyRepo, errors.Unwrap(err))

	users := []dto.User[dto.ID]{}
	cols, jc := orm.GetDataForSelect(user)
	al := orm.GetTableAlias(user)
	assert.Equal(t, sql.ErrNoRows, r.FindOneBy(ctx, cols, squirrel.Eq{"id": 1}, &user))
	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, r.FindByWithInnerJoin(ctx, cols, al, jc, squirrel.Gt{"id": 1}, &users))
	assert.Equal(t, sql.ErrNoRows, r.FindOneByWithInnerJoin(ctx, cols, al, jc, squirrel.Gt{"id": 1}, &users))

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

	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, r.FindBy(ctx, cols, squirrel.Eq{"id": 1}, &users))
	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, r.Select(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), &users))
	assert.Equal(t, db.ErrInvalidRepoEmptyRepo, errors.Unwrap(r.SelectWithCursorOnPKPagination(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), db.CursorPaginationParams{Limit: 10, Cursor: 1}, &users)))

	assert.Equal(t, db.ErrZeroLimitSize, r.SelectWithCursorOnPKPagination(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), db.CursorPaginationParams{}, &users))

	pr, err := r.SelectWithPagePagination(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), db.PagePaginationParams{}, &users)
	assert.NotNil(t, pr)
	assert.Equal(t, db.ErrZeroPageSize, err)

	pr, err = r.SelectWithPagePagination(ctx, squirrel.SelectBuilder{}.Columns(cols...).Where("1=1"), db.PagePaginationParams{PageSize: 10, PageNumber: 1}, &users)
	assert.NotNil(t, pr)
	assert.Equal(t, sql.ErrNoRows, errors.Unwrap(errors.Unwrap(err)))

}

func Test_ConvertFunc(t *testing.T) {
	tests := []struct {
		Input  interface{}
		OutInt int64
		OutStr string
	}{
		{
			Input:  0,
			OutInt: 0,
			OutStr: "0",
		},
		{
			Input:  123,
			OutInt: 123,
			OutStr: "123",
		},
		{
			Input:  "123",
			OutInt: 123,
			OutStr: "123",
		},
		{
			Input:  1.23,
			OutInt: 0,
			OutStr: "1.23",
		},
		{
			Input:  nil,
			OutInt: 0,
			OutStr: "0",
		},
	}

	for _, v := range tests {
		assert.Equal(t, v.OutInt, ConvertIDToInt64(v.Input))
		assert.Equal(t, v.OutStr, ConvertIDToString(v.Input))
	}
}
