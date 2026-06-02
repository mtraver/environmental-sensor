package main

import (
	"fmt"
	"net/http"
)

func (mon *Monitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "%s\n", mon.device.ID())
}
