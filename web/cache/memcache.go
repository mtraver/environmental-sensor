package cache

import (
	"context"
	"sync"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"google.golang.org/appengine/memcache"
	"google.golang.org/protobuf/proto"
)

type Memcache struct {
	mu    sync.RWMutex
	total int
	hits  int
}

// memcacheWriteFunc is the signature of functions in google.golang.org/appengine/memcache that write to the cache.
type memcacheWriteFunc func(context.Context, *memcache.Item) error

func (mc *Memcache) Get(ctx context.Context, key string, m *mpb.Measurement) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.total += 1
	item, err := memcache.Get(ctx, key)

	switch err {
	case nil:
		mc.hits += 1
		return proto.Unmarshal(item.Value, m)
	case memcache.ErrCacheMiss:
		return ErrCacheMiss
	default:
		return err
	}
}

func (mc *Memcache) doWrite(ctx context.Context, key string, m *mpb.Measurement, f memcacheWriteFunc) error {
	data, err := proto.Marshal(m)
	if err != nil {
		return err
	}

	item := &memcache.Item{
		Key:   key,
		Value: data,
	}

	return f(ctx, item)
}

func (mc *Memcache) Add(ctx context.Context, key string, m *mpb.Measurement) error {
	return mc.doWrite(ctx, key, m, memcache.Add)
}

func (mc *Memcache) Set(ctx context.Context, key string, m *mpb.Measurement) error {
	return mc.doWrite(ctx, key, m, memcache.Set)
}

func (mc *Memcache) Stats() Stats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return Stats{
		Total: mc.total,
		Hits:  mc.hits,
	}
}
