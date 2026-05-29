package main

import (
	"fmt"
	"net/http"
)

type indexHandler struct {
	conn Connection
}

func (h indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n", h.conn.Device.ID())
}
