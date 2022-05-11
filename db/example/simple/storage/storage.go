package storage

import (
	"context"

	"go.uber.org/zap"

	"github.com/imperiuse/golib/db"
	"github.com/imperiuse/golib/db/connector"
	"github.com/imperiuse/golib/db/mocks"
)

// MyStorage - simple test storage struct which implement db.Storage[C] interface
type MyStorage[C db.Config] struct {
	cfg C
}

// New - create simple test Storage which implement db.Storage[C] interface
func New[C db.Config](cfg C) db.Storage[C] {
	return &MyStorage[C]{cfg: cfg}
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
	return connector.New[C](m.cfg, zap.NewNop(), mocks.GoodMockDBConn)
}

func (m *MyStorage[C]) Slaves() []db.Connector[C] {
	return []db.Connector[C]{connector.New[C](m.cfg, zap.NewNop(), mocks.GoodMockDBConn)}
}

func (m *MyStorage[C]) Close() error {
	return nil
}
