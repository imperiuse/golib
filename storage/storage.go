package storage

import (
	"sync"
	"time"
)

type (
	WhoisDataStorage struct {
		TTL time.Duration
		mu  sync.Mutex
		m   map[FQDN]whoisData
	}

	FQDN = string

	whoisData struct {
		raw  string
		time time.Time
	}
)

func New(ttl time.Duration) *WhoisDataStorage {
	return &WhoisDataStorage{
		TTL: ttl,
		mu:  sync.Mutex{},
		m:   map[FQDN]whoisData{},
	}
}

func (c *WhoisDataStorage) Get(fqdn FQDN) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, ok := c.m[fqdn]
	if ok && time.Since(data.time) > c.TTL { // auto remove too old data from cache
		delete(c.m, fqdn)
		ok = false
		data.raw = ""
	}

	return data.raw, ok
}

func (c *WhoisDataStorage) Set(fqdn FQDN, whois string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.m[fqdn] = whoisData{whois, time.Now()}
}
