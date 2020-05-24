package main

import (
	"fmt"
	"net/http"

	"github.com/mtraver/iotcore"
)

type indexHandler struct {
	device iotcore.Device
}

func (h indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, h.device.DeviceID)
}
