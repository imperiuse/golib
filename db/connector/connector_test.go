package connector

import (
	"context"
	"database/sql"
	"testing"

	"go.uber.org/zap"

	"github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"

	"github.com/imperiuse/golib/db"
	"github.com/imperiuse/golib/db/example/simple/config"
	"github.com/imperiuse/golib/db/example/simple/dto"
	"github.com/imperiuse/golib/db/mocks"
	"github.com/imperiuse/golib/db/repo/empty"
)

func TestConnector_New(t *testing.T) {
	cfg := config.SimpleTestConfig{}
	logger := zap.NewNop()
	c := New[config.SimpleTestConfig](cfg, logger, nil)

	assert.NotNil(t, c)
	assert.Equal(t, cfg, c.Config())
	assert.Equal(t, logger, c.Logger())
	assert.Nil(t, c.Connection())
}

func TestConnector_All(t *testing.T) {
	ctx := context.Background()

	// Without any options
	cfg := config.New(squirrel.Dollar, false, false)
	c := New[config.SimpleTestConfig](cfg, zap.NewNop(), mocks.GoodMockDBConn)
	assert.Equal(t, mocks.GoodMockDBConn, c.Connection())

	assert.NotNil(t, c.Repo(dto.User[dto.ID]{}))

	// With validation
	cfg = config.New(squirrel.Dollar, true, false)
	c = New[config.SimpleTestConfig](cfg, zap.NewNop(), mocks.GoodMockDBConn)

	assert.Equal(t, empty.Repo, c.Repo(dto.User[dto.ID]{}))

	c.AddAllowsRepos(dto.User[dto.ID]{}.Repo())
	assert.NotNil(t, c.Repo(dto.User[dto.ID]{}))
	assert.NotEqual(t, empty.Repo, c.Repo(dto.User[dto.ID]{}))

	n, err := c.AutoCreate(ctx, dto.User[dto.ID]{})
	assert.Equal(t, int64(0), n)
	assert.Nil(t, err)

	assert.Equal(t, sql.ErrNoRows, c.AutoGet(ctx, dto.User[dto.ID]{}))

	n, err = c.AutoUpdate(ctx, dto.User[dto.ID]{})
	assert.Equal(t, int64(0), n)
	assert.Nil(t, err)

	n, err = c.AutoDelete(ctx, dto.User[dto.ID]{})
	assert.Equal(t, int64(0), n)
	assert.Nil(t, err)

	// With cache
	cfg = config.New(squirrel.Dollar, true, true)
	c = New[config.SimpleTestConfig](cfg, zap.NewNop(), mocks.GoodMockDBConn)

	assert.Equal(t, empty.Repo, c.Repo(dto.User[dto.ID]{}))

	assert.False(t, c.IsAllowRepo(dto.User[dto.ID]{}.Repo()))

	assert.Equal(t, []db.Table{}, c.GetAllowsRepos())

	c.AddAllowsRepos(dto.User[dto.ID]{}.Repo())
	assert.NotNil(t, c.Repo(dto.User[dto.ID]{}))
	assert.NotEqual(t, empty.Repo, c.Repo(dto.User[dto.ID]{}))

	assert.True(t, c.IsAllowRepo(dto.User[dto.ID]{}.Repo()))

	assert.Equal(t, []db.Table{dto.User[dto.ID]{}.Repo()}, c.GetAllowsRepos())

	n, err = c.AutoCreate(ctx, dto.User[dto.ID]{})
	assert.Equal(t, int64(0), n)
	assert.Nil(t, err)

	assert.Equal(t, sql.ErrNoRows, c.AutoGet(ctx, dto.User[dto.ID]{}))

	n, err = c.AutoUpdate(ctx, dto.User[dto.ID]{})
	assert.Equal(t, int64(0), n)
	assert.Nil(t, err)

	n, err = c.AutoDelete(ctx, dto.User[dto.ID]{})
	assert.Equal(t, int64(0), n)
	assert.Nil(t, err)

}
