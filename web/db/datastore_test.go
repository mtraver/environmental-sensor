package db

import (
	"testing"
)

func TestCacheKeyLatest(t *testing.T) {
	expected := "foo#latest"
	key := cacheKeyLatest("foo")
	if key != expected {
		t.Errorf("Incorrect key. Expected %q, got %q", expected, key)
	}
}
