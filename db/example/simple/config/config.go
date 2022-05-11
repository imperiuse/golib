package config

import (
	"github.com/imperiuse/golib/db"
)

// SimpleTestConfig - simple test config which implement db.Config
type SimpleTestConfig struct {
	phf                 db.PlaceholderFormat
	isEnabledValidation bool
	isEnabledRepoCache  bool
}

func New(phf db.PlaceholderFormat, isValidationEnable, isCacheEnable bool) SimpleTestConfig {
	return SimpleTestConfig{
		phf:                 phf,
		isEnabledValidation: isValidationEnable,
		isEnabledRepoCache:  isCacheEnable,
	}
}

func (c SimpleTestConfig) PlaceholderFormat() db.PlaceholderFormat {
	return c.phf
}

func (c SimpleTestConfig) IsEnableValidationRepoNames() bool {
	return c.isEnabledValidation
}

func (c SimpleTestConfig) IsEnableReposCache() bool {
	return c.isEnabledRepoCache
}
