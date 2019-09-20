package storage

import (
	"sync"
	"time"
)

const (
	TTLCacheResetDefault = 24 * time.Hour
	TTLCacheDefault      = 5 * time.Minute
)

type (
	Key   = string
	Value = struct {
		time time.Time
		data interface{}
	}
	StoreMap = map[Key]Value

	Store struct {
		TTL time.Duration
		mu  sync.Mutex
		m   StoreMap
	}

	StoreI interface {
		Get(Key) (interface{}, bool)
		Set(Key, interface{})
	}
)

func New(ttl time.Duration, autoCleanTimeout time.Duration) *Store {

	// "защита от дурака"
	if autoCleanTimeout == 0 {
		autoCleanTimeout = TTLCacheResetDefault
	}

	if ttl == 0 {
		ttl = TTLCacheDefault
	}

	storage := &Store{
		TTL: ttl,
		mu:  sync.Mutex{},
		m:   StoreMap{},
	}

	// запуск горутины которая раз в autoCleanTimeout полностью сбрасывае кэш
	go func() {
		for {
			time.Sleep(autoCleanTimeout)
			storage.RemoveAll()
		}
	}()

	return storage
}

func (c *Store) RemoveAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.m = StoreMap{}
}

func (c *Store) Get(k Key) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.get(k, c.TTL)

}

func (c *Store) GetTTL(k Key, ttl time.Duration) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.get(k, ttl)
}

func (c *Store) get(k Key, ttl time.Duration) (interface{}, bool) {
	v, ok := c.m[k]
	if ok && time.Since(v.time) > c.TTL { // auto remove too old data from cache
		delete(c.m, k)
		ok = false
		v.data = ""
	}

	return v.data, ok
}

func (c *Store) Set(k Key, v interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.m[k] = Value{time.Now(), v}
}

func (c *Store) Delete(k Key) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.m, k)
}
