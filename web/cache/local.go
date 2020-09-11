package cache

import (
	"context"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
)

type Local struct {
	cache map[string]*mpb.Measurement
}

func NewLocal() Local {
	return Local{
		cache: make(map[string]*mpb.Measurement),
	}
}

func (l Local) Get(ctx context.Context, key string, m *mpb.Measurement) error {
	got, ok := l.cache[key]

	if !ok {
		return ErrCacheMiss
	}

	*m = *got
	return nil
}

func (l Local) Add(ctx context.Context, key string, m *mpb.Measurement) error {
	if _, ok := l.cache[key]; ok {
		return ErrNotStored
	}

	l.cache[key] = m
	return nil
}

func (l Local) Set(ctx context.Context, key string, m *mpb.Measurement) error {
	l.cache[key] = m
	return nil
}
