package cache

import "errors"

var (
	ErrCacheMiss = errors.New("cache: cache miss")
	ErrNotStored = errors.New("cache: item not stored")
)
