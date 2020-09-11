// +build go1.12

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mtraver/environmental-sensor/web/cache"
)

func newContext(r *http.Request) context.Context {
	return r.Context()
}

func newCache() cache.Cache {
	return cache.NewLocal()
}

func serve(mux http.Handler) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), mux))
}
