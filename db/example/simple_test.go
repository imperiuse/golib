package example

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/imperiuse/golib/db/example/simple/config"
	"github.com/imperiuse/golib/db/example/simple/dto"
	"github.com/imperiuse/golib/db/example/simple/storage"
	"github.com/imperiuse/golib/db/repo"
)

func TestStorage_New(t *testing.T) {
	cfg := config.New(squirrel.Dollar, false, false)
	s := storage.New[config.SimpleTestConfig](cfg)
	assert.NotNil(t, s)
}

func TestStorage_Various_Get_Examples(t *testing.T) {
	ctx := context.Background()
	cfg := config.New(squirrel.Dollar, false, false)
	s := storage.New[config.SimpleTestConfig](cfg)

	user, err := repo.NewGen[dto.ID, dto.User[dto.ID]](s.Master()).Get(ctx, 1) // here can be cache for Repos under the hood
	assert.NotNil(t, user)
	assert.Equal(t, sql.ErrNoRows, err)

	var user2 = dto.User[dto.ID]{}
	err = s.Master().Repo(&user2).Get(ctx, 2, &user2) // here can be cache for Repos under the hood
	assert.Equal(t, sql.ErrNoRows, err)

	var user3 = dto.User[dto.ID]{BaseDTO: dto.BaseDTO[dto.ID]{Id: 3}}
	err = s.Master().AutoGet(ctx, &user3) // // here can be cache for Repos under the hood
	assert.Equal(t, sql.ErrNoRows, err)

	var user4 = dto.User[dto.ID]{BaseDTO: dto.BaseDTO[dto.ID]{Id: 3}} // no, any cache
	rows, err := s.Master().Connection().QueryContext(
		ctx, fmt.Sprintf("SELECT * FROM %s WHERE id = ? LIMIT 1;", user4.Repo()), user.Identity(),
	)
	_ = rows // parsing rows manually or
	assert.Nil(t, err)
	assert.Nil(t, err)

	var user5 = dto.User[dto.ID]{BaseDTO: dto.BaseDTO[dto.ID]{Id: 3}} // no, any cache
	err = sqlx.GetContext(ctx, s.Master().Connection(), &user5,
		fmt.Sprintf("SELECT * FROM %s WHERE id = ? LIMIT 1;", user4.Repo()), user.Identity())
	assert.Equal(t, sql.ErrNoRows, err)
}
