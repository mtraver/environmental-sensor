package cache

import (
	"context"
	"sync"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
)

type Local struct {
	cache map[string]*mpb.Measurement
	mu    sync.RWMutex
	total int
	hits  int
}

func NewLocal() Local {
	return Local{
		cache: make(map[string]*mpb.Measurement),
	}
}

func (l *Local) Get(ctx context.Context, key string, m *mpb.Measurement) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.total += 1
	got, ok := l.cache[key]

	if !ok {
		return ErrCacheMiss
	}

	l.hits += 1
	*m = *got
	return nil
}

func (l *Local) Add(ctx context.Context, key string, m *mpb.Measurement) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.cache[key]; ok {
		return ErrNotStored
	}

	l.cache[key] = m
	return nil
}

func (l *Local) Set(ctx context.Context, key string, m *mpb.Measurement) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.cache[key] = m
	return nil
}

func (l *Local) Stats() Stats {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return Stats{
		Total: l.total,
		Hits:  l.hits,
	}
}
