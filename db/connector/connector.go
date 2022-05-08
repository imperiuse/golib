package connector

import (
	"context"
	"github.com/imperiuse/golib/db"
	"github.com/imperiuse/golib/db/repository"
	"github.com/imperiuse/golib/db/repository/empty"
	"sync"
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

// AddRepoNames - store information about available repo names
func (c *connector[C]) AddRepoNames(repos ...db.Table) {
	c.mV.Lock()
	defer c.mV.Unlock()

	for _, v := range repos {
		c.validationRepoMap[v] = new(any)
	}

	return
}

func (c *connector[C]) Config() C {
	return c.cfg
}

func (c *connector[C]) Logger() db.Logger {
	return c.logger
}

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

		repo, found := c.cacheRepoMap[repoName]
		if found {
			return repo
		}

		repo = repository.New(c.logger, c.dbConn, repoName, c.phf)
		c.cacheRepoMap[repoName] = repo

		return repo
	}

	return repository.New(c.logger, c.dbConn, repoName, c.phf)
}

func (c *connector[C]) AutoCreate(ctx context.Context, dto db.DTO) (int64, error) {
	return c.Repo(dto).Create(ctx, dto)
}

func (c *connector[C]) AutoGet(ctx context.Context, dto db.DTO) error {
	return c.Repo(dto).Get(ctx, dto.Identity(), dto)
}

func (c *connector[C]) AutoUpdate(ctx context.Context, dto db.DTO) (int64, error) {
	return c.Repo(dto).Update(ctx, dto.Identity(), dto)
}

func (c *connector[C]) AutoDelete(ctx context.Context, dto db.DTO) (int64, error) {
	return c.Repo(dto).Delete(ctx, dto)
}
