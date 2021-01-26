package cache

import (
	"context"
	"errors"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
)

var (
	ErrCacheMiss = errors.New("cache: cache miss")
	ErrNotStored = errors.New("cache: item not stored")
)

type Stats struct {
	Total int
	Hits  int
}

type Cache interface {
	Get(ctx context.Context, key string, m *mpb.Measurement) error
	Add(ctx context.Context, key string, m *mpb.Measurement) error
	Set(ctx context.Context, key string, m *mpb.Measurement) error
	Stats() Stats
}
