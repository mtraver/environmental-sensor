//go:build !go1.12

package main

import (
	"context"
	"net/http"

	"github.com/mtraver/environmental-sensor/web/cache"
	"google.golang.org/appengine"
)

func newContext(r *http.Request) context.Context {
	return appengine.NewContext(r)
}

func newCache() cache.Cache {
	return &cache.Memcache{}
}

func serve(mux http.Handler) {
	http.Handle("/", mux)
	appengine.Main()
}
