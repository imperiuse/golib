package cache

import (
	"fmt"
	"sync"
	"time"
)

//go:generate moq  -out mock_cache.go -skip-ensure -pkg mocks . Cache
type (
	// Config - for cache.
	Config struct {
		DefaultTTL string `yaml:"default_ttl"`

		CustomCacheConfig map[string]any `yaml:"customCacheConfig"`
	}

	// TTL - struct describe TTL use cases for cache.
	TTL = struct {
		TTL      time.Duration // PREFER(more important) THAN ExpireAt
		ExpireAt time.Time
	}

	// Cache - interface of cache.
	Cache[K comparable, V any] interface {
		// Get - get data(Value) for Key (return Value if key doesn't expire)
		Get(K) (V, bool)

		// Set - set data(Value) for Key (without TTL, not expired)
		Set(K, V)

		// SetWithTTL - set data(Value) for Key (with TTL or when expired)
		SetWithTTL(K, V, TTL)

		// SetTTL - set TTL for Key
		SetTTL(K, TTL)

		// TryGetOrInvokeLambda - try Get(Key) if not found, exec Lambda and if success store Value by Set func
		TryGetOrInvokeLambda(K, lambda[K, V]) (V, error)

		// Delete - delete Key
		Delete(K)

		// CleanAll - delete all data in storage (all Key)
		CleanAll()
	}

	lambda[K comparable, V any] func(K) (V, TTL, error)
)

// Specific realization of Cache interface
type (
	storage[K comparable, V any] struct {
		defaultTTL time.Duration // default TTL for data obj in md

		sync.RWMutex
		md map[K]data[V] // map of data
	}

	data[V any] struct {
		data  V
		creAt time.Time
		expAt time.Time
	}
)

// New "Constructor" of Cache
func New[K comparable, V any](config Config) (Cache[K, V], error) {
	defaultTTL, err := time.ParseDuration(config.DefaultTTL)
	if err != nil {
		return nil, fmt.Errorf("parse time.ParseDuration(config.DefaultTTL): %w", err)
	}

	return &storage[K, V]{
		defaultTTL: defaultTTL,
		RWMutex:    sync.RWMutex{},
		md:         map[K]data[V]{},
	}, nil
}

// Get - get data by key
func (s *storage[K, V]) Get(key K) (V, bool) {
	s.RLock()
	d, found := s.md[key]
	s.RUnlock()

	if !found {
		return *new(V), false
	}

	if !d.expAt.IsZero() && time.Now().UTC().After(d.expAt) {
		s.Lock()
		delete(s.md, key)
		s.Unlock()

		return *new(V), false
	}

	return d.data, true
}

// Set - store data to specific key
func (s *storage[K, V]) Set(key K, value V) {
	s.Lock()
	defer s.Unlock()

	s.md[key] = data[V]{
		data:  value,
		creAt: time.Now().UTC(),
		expAt: time.Now().UTC().Add(s.defaultTTL),
	}
}

// SetWithTTL - set with specific TTL preferences
func (s *storage[K, V]) SetWithTTL(key K, value V, ttl TTL) {
	expAt := time.Now().UTC().Add(s.defaultTTL)

	if ttl.TTL != 0 {
		expAt = time.Now().UTC().Add(ttl.TTL)
	} else if !ttl.ExpireAt.IsZero() {
		expAt = ttl.ExpireAt.UTC()
	}

	s.Lock()
	defer s.Unlock()

	s.md[key] = data[V]{
		data:  value,
		creAt: time.Now().UTC(),
		expAt: expAt,
	}
}

// SetTTL - update TTL for specific key (if key exist)
func (s *storage[K, V]) SetTTL(key K, ttl TTL) {
	s.RLock()
	d, found := s.md[key]
	s.RUnlock()

	if !found {
		return
	}

	if ttl.TTL != 0 {
		s.Lock()
		defer s.Unlock()

		s.md[key] = data[V]{
			data:  d.data,
			creAt: d.creAt,
			expAt: time.Now().UTC().Add(ttl.TTL),
		}

		return
	}

	if !ttl.ExpireAt.IsZero() {
		s.Lock()
		defer s.Unlock()

		s.md[key] = data[V]{
			data:  d.data,
			creAt: d.creAt,
			expAt: ttl.ExpireAt.UTC(),
		}

		return
	}
}

// TryGetOrInvokeLambda  - try to get data if exit return him, if not exist - run lambda func and store result to key
func (s *storage[K, V]) TryGetOrInvokeLambda(key K, f lambda[K, V]) (V, error) {
	if d, found := s.Get(key); found {
		return d, nil
	}

	v, ttl, err := f(key)
	if err != nil {
		return *new(V), fmt.Errorf("cache: can't exec lambda for get key value: %w", err)
	}

	s.SetWithTTL(key, v, ttl)

	return v, nil
}

// Delete - delete specific data by key
func (s *storage[K, V]) Delete(key K) {
	s.Lock()
	defer s.Unlock()

	delete(s.md, key)
}

// CleanAll - delete all keys
func (s *storage[K, V]) CleanAll() {
	s.Lock()
	defer s.Unlock()

	s.md = map[K]data[V]{}
}
