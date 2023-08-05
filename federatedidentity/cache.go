package federatedidentity

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

type credentialCache struct {
	mu      sync.Mutex
	entries map[string]entry
}

type entry struct {
	cred *credentials.Credentials
	exp  time.Time
}

func newCache() credentialCache {
	return credentialCache{
		entries: make(map[string]entry),
	}
}

func (c credentialCache) set(key string, cred *credentials.Credentials, ttl time.Duration) {
	e := entry{
		cred: cred,
		exp:  time.Now().Add(ttl),
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = e
}

func (c credentialCache) get(key string) *credentials.Credentials {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.entries[key]
	if !ok {
		return nil
	}

	// Present and unexpired
	if time.Now().Before(e.exp) {
		return e.cred
	}

	// Expired
	delete(c.entries, key)
	return nil
}

func (c credentialCache) clean() {
	c.mu.Lock()
	defer c.mu.Unlock()

	toRemove := []string{}

	for k, e := range c.entries {
		if !time.Now().Before(e.exp) {
			toRemove = append(toRemove, k)
		}
	}

	for _, k := range toRemove {
		delete(c.entries, k)
	}
}
