package tests

import (
	"context"

	"go.uber.org/zap"

	"github.com/stretchr/testify/suite"
)

type (
	Suite struct {
		suite.Suite

		Ctx    context.Context
		Cancel context.CancelFunc

		Log *zap.Logger
	}
)

func (s *Suite) Setup() {
	s.Ctx, s.Cancel = context.WithCancel(context.Background())
	s.Log = zap.NewNop()
}
