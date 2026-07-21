package db

import (
	"testing"
)

func TestCacheKeyLatest(t *testing.T) {
	want := "foo#latest"
	got := cacheKeyLatest("foo")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
