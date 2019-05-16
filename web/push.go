package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/mtraver/gaelog"
	"google.golang.org/appengine"

	"github.com/mtraver/environmental-sensor/measurement"
)

// This is the structure of the JSON payload pushed to the endpoint by Cloud Pub/Sub.
// See https://cloud.google.com/pubsub/docs/push.
type pushRequest struct {
	Message struct {
		Attributes map[string]string
		Data       []byte
		ID         string `json:"message_id"`
	}
	Subscription string
}

func pushHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	lg, err := gaelog.New(r)
	if err != nil {
		lg.Errorf("%v", err)
	}
	defer lg.Close()

	msg := &pushRequest{}
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		lg.Criticalf("Could not decode body: %v\n", err)
		http.Error(w, fmt.Sprintf("Could not decode body: %v", err), http.StatusBadRequest)
		return
	}

	m := &measurement.Measurement{}
	if err := proto.Unmarshal(msg.Message.Data, m); err != nil {
		lg.Criticalf("Failed to unmarshal protobuf: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to unmarshal protobuf: %v", err), http.StatusBadRequest)
		return
	}

	if err := m.Validate(); err != nil {
		lg.Errorf("%v", err)

		// Pub/Sub will only stop re-trying the message if it receives a status 200.
		// The docs say that any of 200, 201, 202, 204, or 102 will have this effect
		// (https://cloud.google.com/pubsub/docs/push), but the local emulator
		// doesn't respect anything other than 200, so return 200 just to be safe.
		// TODO(mtraver) I'd rather return e.g. 202 (http.StatusAccepted) to
		// indicate that it was successfully received but not that all is ok.
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := database.Save(ctx, m); err != nil {
		lg.Errorf("Failed to save measurement: %v\n", err)
	}

	w.WriteHeader(http.StatusOK)
}
