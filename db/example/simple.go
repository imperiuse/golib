package example

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/imperiuse/golib/db/connector"
	"github.com/imperiuse/golib/db/mocks"
	"go.uber.org/zap"

	"github.com/imperiuse/golib/db"
)

type MyStorageConfig struct{}

func (c MyStorageConfig) PlaceholderFormat() db.PlaceholderFormat {
	return squirrel.Dollar
}

func (c MyStorageConfig) IsEnableValidationRepoNames() bool {
	return false
}

func (c MyStorageConfig) IsEnableReposCache() bool {
	return false
}

type MyStorage[C db.Config] struct {
	cfg C
}

func New[C db.Config]() db.Storage[C] {
	return &MyStorage[C]{cfg: *new(C)}
}

type UserDTO struct {
	id db.ID
}

func (u UserDTO) Repo() db.Table {
	return "user"
}

func (u UserDTO) Identity() db.ID {
	return u.id
}

func (m *MyStorage[C]) Config() C {
	return m.cfg
}

func (m *MyStorage[C]) Connect() error {
	return nil
}

func (m *MyStorage[C]) Reconnect() error {
	return nil
}

func (m *MyStorage[C]) OnConnect(_ context.Context, _ func()) error {
	return nil
}

func (m *MyStorage[C]) OnReconnect(_ context.Context, _ func()) error {
	return nil
}

func (m *MyStorage[C]) OnStop(_ context.Context, _ func()) error {
	return nil
}

func (m *MyStorage[C]) Master() db.Connector[C] {
	return connector.New[C](m.cfg, zap.NewNop(), mocks.GoodMockDBConn, squirrel.Dollar)
}

func (m *MyStorage[C]) Slaves() []db.Connector[C] {
	return []db.Connector[C]{connector.New[C](m.cfg, zap.NewNop(), mocks.GoodMockDBConn, squirrel.Dollar)}
}

func (m *MyStorage[C]) Close() error {
	return nil
}
