package example

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/imperiuse/golib/db/repository"

	"github.com/stretchr/testify/assert"
)

func TestStorage_New(t *testing.T) {
	ctx := context.Background()

	s := New[MyStorageConfig]()
	assert.NotNil(t, s)

	user, err := repository.NewGen[UserDTO](s.Master()).Get(ctx, 1) // here can be cache for Repos under the hood
	assert.NotNil(t, user)
	assert.Nil(t, err)

	var user2 = UserDTO{}
	err = s.Master().Repo(&user2).Get(ctx, 2, &user2) // here can be cache for Repos under the hood
	assert.Nil(t, err)

	var user3 = UserDTO{id: 3}
	err = s.Master().AutoGet(ctx, &user3) // // here can be cache for Repos under the hood
	assert.Nil(t, err)

	var user4 = UserDTO{id: 4} // no, any cache
	rows, err := s.Master().Connection().QueryContext(
		ctx, fmt.Sprintf("SELECT * FROM %s WHERE id = ? LIMIT 1;", user4.Repo()), user.Identity(),
	)
	_ = rows // parsing rows manually or
	assert.Nil(t, err)
	assert.NotNil(t, rows)

	var user5 = UserDTO{id: 5} // no, any cache
	err = sqlx.GetContext(ctx, s.Master().Connection(), &user5, fmt.Sprintf("SELECT * FROM %s WHERE id = ? LIMIT 1;", user4.Repo()), user.Identity())
	assert.Nil(t, err)
}
