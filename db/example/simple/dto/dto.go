package dto

import (
	"time"

	"github.com/imperiuse/golib/db"
)

// Various example of DTO's
type (
	ID = int

	NotDTO struct{}

	BaseDTO[I db.ID] struct {
		Id        I         `db:"id"          orm_use_in:"select"`
		CreatedAt time.Time `db:"created_at"  orm_use_in:"select"`
		UpdatedAt time.Time `db:"updated_at"  orm_use_in:"select,update"`
	}

	User[I db.ID] struct {
		BaseDTO[I]
		Name     string `db:"name"     orm_use_in:"select,create,update"`
		Email    string `db:"email"    orm_use_in:"select,create,update"`
		Password string `db:"password" orm_use_in:"select,create,update"`
		RoleID   I      `db:"role_id" orm_use_in:"select,create,update"`
		_        any    `orm_table_name:"Users" orm_alias:"u"`
	}

	Role[I db.ID] struct {
		BaseDTO[I]
		Name   string `db:"name"       orm_use_in:"select,create,update"`
		Rights int    `db:"rights"     orm_use_in:"select,create,update"`
		_      any    `orm_table_name:"Roles" orm_alias:"r"`
	}

	UsersRole[I db.ID] struct {
		User[I] `db:"u" orm_alias:"u"`
		Role[I] `db:"r" orm_alias:"r"`
		_       any `orm_join:" ON u.role_id = r.id "`
	}

	// Paginator - test table for pagination methods
	Paginator[I db.ID] struct {
		BaseDTO[I]
		Name string `db:"name"    orm_use_in:"select,create,update"`
		N    int    `db:"n"       orm_use_in:"select,create,update"`

		_ any `orm_table_name:"Paginators"  orm_alias:"p"`
	}
)

func (b BaseDTO[I]) Identity() db.ID {
	return b.Id
}

func (b BaseDTO[I]) ID() I {
	return b.Id
}

func (b UsersRole[I]) ID() I {
	return *new(I)
}

func (b UsersRole[I]) Identity() any {
	return *new(I)
}

func (b UsersRole[I]) Repo() string {
	return "UsersRole"
}

func (_ Paginator[I]) Repo() db.Table {
	return "Paginators"
}

func (_ Role[I]) Repo() db.Table {
	return "Roles"
}

func (_ User[I]) Repo() db.Table {
	return "Users"
}

var DSL = map[string]string{
	"Roles": `CREATE TABLE IF NOT EXISTS Roles
(
id           INTEGER     PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
created_at   TIMESTAMP   NOT NULL DEFAULT NOW(),
updated_at   TIMESTAMP   NOT NULL DEFAULT NOW(),
name         TEXT        NOT NULL,
rights      INTEGER     NOT NULL
);`,
	"Users": `CREATE TABLE IF NOT EXISTS Users
(
id           INTEGER     PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
created_at   TIMESTAMP   NOT NULL DEFAULT NOW(),
updated_at   TIMESTAMP   NOT NULL DEFAULT NOW(),
name         TEXT        NOT NULL,
email        TEXT        NOT NULL,
password     TEXT        NOT NULL,
role_id      INTEGER     NOT NULL,
CONSTRAINT fkey__r FOREIGN KEY (role_id) REFERENCES roles (id) MATCH SIMPLE	ON UPDATE NO ACTION ON DELETE CASCADE
);`,
	"Paginators": `CREATE TABLE IF NOT EXISTS Paginators
(
id           INTEGER     PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
created_at   TIMESTAMP   NOT NULL DEFAULT NOW(),
updated_at   TIMESTAMP   NOT NULL DEFAULT NOW(),
name         TEXT        NOT NULL,
n            INTEGER     NOT NULL
);`,
}
