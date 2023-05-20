package main

import (
	"fmt"
	"net/http"
)

type indexHandler struct {
	connections map[ConnectionType]Connection
}

func (h indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for name, conn := range h.connections {
		fmt.Fprintf(w, "%s: %s\n", name, conn.Device.ID())
	}
}
