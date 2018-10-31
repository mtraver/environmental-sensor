package cache

import (
	"context"
	"encoding/json"
	"errors"

	"google.golang.org/appengine/memcache"

	"github.com/mtraver/environmental-sensor/receiver/measurement"
)

var ErrCacheMiss = errors.New("cache: cache miss")

// memcacheWriteFunc is the signature of functions in google.golang.org/appengine/memcache
// that write to the cache
type memcacheWriteFunc func(context.Context, *memcache.Item) error

func Get(ctx context.Context, key string, m *measurement.StorableMeasurement) error {
	item, err := memcache.Get(ctx, key)

	switch err {
	case nil:
		return json.Unmarshal(item.Value, m)
	case memcache.ErrCacheMiss:
		return ErrCacheMiss
	default:
		return err
	}
}

func doWrite(ctx context.Context, key string, m *measurement.StorableMeasurement, f memcacheWriteFunc) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	item := &memcache.Item{
		Key:   key,
		Value: data,
	}

	return f(ctx, item)
}

func Add(ctx context.Context, key string, m *measurement.StorableMeasurement) error {
	return doWrite(ctx, key, m, memcache.Add)
}

func Set(ctx context.Context, key string, m *measurement.StorableMeasurement) error {
	return doWrite(ctx, key, m, memcache.Set)
}
