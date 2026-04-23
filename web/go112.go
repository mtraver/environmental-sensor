//go:build go1.12

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/mtraver/environmental-sensor/web/cache"
	"github.com/mtraver/gaelog"
)

func newContext(r *http.Request) context.Context {
	return r.Context()
}

func newCache() cache.Cache {
	c := cache.NewLocal()
	return &c
}

func serve(mux http.Handler) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), gaelog.Wrap(stripTrailingSlash(mux))))
}

func stripTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
		next.ServeHTTP(w, r)
	})
}
