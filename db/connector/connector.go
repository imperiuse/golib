package connector

import (
	"context"
	"sync"

	"github.com/imperiuse/golib/db"
	"github.com/imperiuse/golib/db/repo"
	"github.com/imperiuse/golib/db/repo/empty"
)

type connector[C db.Config] struct {
	cfg    C
	logger db.Logger

	dbConn db.PureSqlxConnection
	phf    db.PlaceholderFormat

	// Special features for checking Repo names, caching and so on...
	mV                sync.RWMutex
	validationRepoMap map[db.Table]any
	mC                sync.Mutex
	cacheRepoMap      map[db.Table]db.Repository
}

func New[C db.Config](cfg C, logger db.Logger, dbConn db.PureSqlxConnection) db.Connector[C] {
	return &connector[C]{
		cfg:    cfg,
		logger: logger,
		dbConn: dbConn,
		phf:    cfg.PlaceholderFormat(),

		mV:                sync.RWMutex{},
		validationRepoMap: map[db.Table]any{},
		mC:                sync.Mutex{},
		cacheRepoMap:      map[db.Table]db.Repository{},
	}
}

// AddAllowsRepos - store information about available repo names
func (c *connector[C]) AddAllowsRepos(repos ...db.Table) {
	c.mV.Lock()
	defer c.mV.Unlock()

	for _, v := range repos {
		c.validationRepoMap[v] = new(any)
	}

	return
}

// GetAllowsRepos - get list of allow repos
func (c *connector[C]) GetAllowsRepos() []db.Table {
	c.mV.RLock()
	defer c.mV.RUnlock()

	var l = make([]db.Table, 0, len(c.validationRepoMap))
	for tableName := range c.validationRepoMap {
		l = append(l, tableName)
	}

	return l
}

// IsAllowRepo - is repo name in validationRepoMap (were previously add by connector.AddRepoNames)
func (c *connector[C]) IsAllowRepo(repo db.Table) bool {
	c.mV.RLock()
	defer c.mV.RUnlock()

	_, found := c.validationRepoMap[repo]

	return found
}

// Config - return config of connector
func (c *connector[C]) Config() C {
	return c.cfg
}

// Logger - return logger instance (*zap.Logger)
func (c *connector[C]) Logger() db.Logger {
	return c.logger
}

// Connection - return pure sqlx connection
func (c *connector[C]) Connection() db.PureSqlxConnection {
	return c.dbConn
}

// Repo - return db.Repository based on dto.Name() method
// if cfg.IsEnableValidationRepoNames() == true =>  do validation action too)
// if cfg.IsEnableReposCache() == true => use cache.
func (c *connector[C]) Repo(dto db.DTO) db.Repository {
	var repoName = dto.Repo()

	if c.cfg.IsEnableValidationRepoNames() {
		c.mV.RLock()
		defer c.mV.RUnlock()

		if _, found := c.validationRepoMap[repoName]; !found {
			return empty.Repo
		}
	}

	if c.cfg.IsEnableReposCache() {
		c.mC.Lock()
		defer c.mC.Unlock()

		r, found := c.cacheRepoMap[repoName]
		if found {
			return r
		}

		r = repo.New(c.logger, c.dbConn, repoName, c.phf)
		c.cacheRepoMap[repoName] = r

		return r
	}

	return repo.New(c.logger, c.dbConn, repoName, c.phf)
}

// AutoCreate - wrapper for c.Repo(dto).Create(ctx, dto)
func (c *connector[C]) AutoCreate(ctx context.Context, dto db.DTO) (int64, error) {
	return c.Repo(dto).Create(ctx, dto)
}

// AutoGet - wrapper for c.Repo(dto).Get(ctx, dto.Identity(), dto)
func (c *connector[C]) AutoGet(ctx context.Context, dto db.DTO) error {
	return c.Repo(dto).Get(ctx, dto.Identity(), dto)
}

// AutoUpdate - wrapper for c.Repo(dto).Update(ctx, dto.Identity(), dto)
func (c *connector[C]) AutoUpdate(ctx context.Context, dto db.DTO) (int64, error) {
	return c.Repo(dto).Update(ctx, dto.Identity(), dto)
}

// AutoDelete - wrapper for c.Repo(dto).Delete(ctx, dto.Identity())
func (c *connector[C]) AutoDelete(ctx context.Context, dto db.DTO) (int64, error) {
	return c.Repo(dto).Delete(ctx, dto.Identity())
}
