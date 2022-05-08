package example

import (
	"context"
	"fmt"
	"testing"

	"github.com/Masterminds/squirrel"

	"github.com/imperiuse/golib/db/example/simple/config"
	"github.com/imperiuse/golib/db/example/simple/dto"
	"github.com/imperiuse/golib/db/example/simple/storage"

	"github.com/jmoiron/sqlx"

	"github.com/imperiuse/golib/db/repository"

	"github.com/stretchr/testify/assert"
)

func TestStorage_New(t *testing.T) {
	ctx := context.Background()

	cfg := config.New(squirrel.Dollar, false, false)
	s := storage.New[config.SimpleTestConfig](cfg)
	assert.NotNil(t, s)

	user, err := repository.NewGen[dto.User](s.Master()).Get(ctx, 1) // here can be cache for Repos under the hood
	assert.NotNil(t, user)
	assert.Nil(t, err)

	var user2 = dto.User{}
	err = s.Master().Repo(&user2).Get(ctx, 2, &user2) // here can be cache for Repos under the hood
	assert.Nil(t, err)

	var user3 = dto.User{BaseDTO: dto.BaseDTO{ID: 3}}
	err = s.Master().AutoGet(ctx, &user3) // // here can be cache for Repos under the hood
	assert.Nil(t, err)

	var user4 = dto.User{BaseDTO: dto.BaseDTO{ID: 3}} // no, any cache
	rows, err := s.Master().Connection().QueryContext(
		ctx, fmt.Sprintf("SELECT * FROM %s WHERE id = ? LIMIT 1;", user4.Repo()), user.Identity(),
	)
	_ = rows // parsing rows manually or
	assert.Nil(t, err)
	assert.NotNil(t, rows)

	var user5 = dto.User{BaseDTO: dto.BaseDTO{ID: 3}} // no, any cache
	err = sqlx.GetContext(ctx, s.Master().Connection(), &user5,
		fmt.Sprintf("SELECT * FROM %s WHERE id = ? LIMIT 1;", user4.Repo()), user.Identity())
	assert.Nil(t, err)
}
