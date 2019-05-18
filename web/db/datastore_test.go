package db

import (
	"testing"
)

func TestCacheKeyLatest(t *testing.T) {
	want := "foo#latest"
	got := cacheKeyLatest("foo")
	if got != want {
		t.Errorf("Want %q, got %q", want, got)
	}
}
