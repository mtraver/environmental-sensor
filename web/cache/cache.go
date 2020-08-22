package cache

import (
	"context"
	"errors"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"google.golang.org/appengine/memcache"
	"google.golang.org/protobuf/proto"
)

var ErrCacheMiss = errors.New("cache: cache miss")

// memcacheWriteFunc is the signature of functions in google.golang.org/appengine/memcache that write to the cache.
type memcacheWriteFunc func(context.Context, *memcache.Item) error

func Get(ctx context.Context, key string, m *mpb.Measurement) error {
	item, err := memcache.Get(ctx, key)

	switch err {
	case nil:
		return proto.Unmarshal(item.Value, m)
	case memcache.ErrCacheMiss:
		return ErrCacheMiss
	default:
		return err
	}
}

func doWrite(ctx context.Context, key string, m *mpb.Measurement, f memcacheWriteFunc) error {
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

func Add(ctx context.Context, key string, m *mpb.Measurement) error {
	return doWrite(ctx, key, m, memcache.Add)
}

func Set(ctx context.Context, key string, m *mpb.Measurement) error {
	return doWrite(ctx, key, m, memcache.Set)
}
