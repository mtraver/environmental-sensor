package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	mpbutil "github.com/mtraver/environmental-sensor/measurementpbutil"
	"github.com/mtraver/environmental-sensor/web/db"
	"github.com/mtraver/gaelog"
	"google.golang.org/api/idtoken"
	"google.golang.org/protobuf/proto"
)

const testDeviceTag = "test"

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

// pushHandler handles Pub/Sub push deliveries originating from Google Cloud IoT Core.
type pushHandler struct {
	PubSubToken    string
	PubSubAudience string
	Database       Database
	InfluxDB       *db.InfluxDB
}

// authenticate validates the JWT signed by Pub/Sub.
func (h pushHandler) authenticate(ctx context.Context, r *http.Request) error {
	// Verify the token provided as a param in the URL requested by Pub/Sub.
	if token, ok := r.URL.Query()["token"]; !ok || len(token) != 1 || token[0] != h.PubSubToken {
		return errors.New("Bad token")
	}

	// Get the Pub/Sub-generated JWT from the "Authorization" header.
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(strings.Split(authHeader, " ")) != 2 {
		return errors.New("Missing Authorization header")
	}
	token := strings.Split(authHeader, " ")[1]

	// Decode and verify the JWT.
	payload, err := idtoken.Validate(ctx, token, h.PubSubAudience)
	if err != nil {
		return fmt.Errorf("Invalid JWT: %w", err)
	}
	if payload.Issuer != "accounts.google.com" && payload.Issuer != "https://accounts.google.com" {
		return errors.New("Wrong issuer")
	}

	return nil
}

func (h pushHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := newContext(r)

	if err := h.authenticate(ctx, r); err != nil {
		gaelog.Criticalf(ctx, "Authentication failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg := &pushRequest{}
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		gaelog.Criticalf(ctx, "Could not decode body: %v\n", err)
		http.Error(w, fmt.Sprintf("Could not decode body: %v", err), http.StatusBadRequest)
		return
	}

	m := &mpb.Measurement{}
	if err := proto.Unmarshal(msg.Message.Data, m); err != nil {
		gaelog.Criticalf(ctx, "Failed to unmarshal protobuf: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to unmarshal protobuf: %v", err), http.StatusBadRequest)
		return
	}

	if err := mpbutil.Validate(m); err != nil {
		gaelog.Errorf(ctx, "%v", err)

		// Pub/Sub will only stop re-trying the message if it receives a status 200.
		// The docs say that any of 200, 201, 202, 204, or 102 will have this effect
		// (https://cloud.google.com/pubsub/docs/push), but the local emulator
		// doesn't respect anything other than 200, so return 200 just to be safe.
		// TODO(mtraver) I'd rather return e.g. 202 (http.StatusAccepted) to
		// indicate that it was successfully received but not that all is ok.
		w.WriteHeader(http.StatusOK)
		return
	}

	// If the device ID contains the string "test" then we won't save this measurement to the database.
	if strings.Contains(m.GetDeviceId(), testDeviceTag) {
		gaelog.Infof(ctx, "Got measurement from test device with ID %q, so it will not be saved to the database. Measurement: %+v", m.GetDeviceId(), m)
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := h.Database.Save(ctx, m); err != nil {
		gaelog.Errorf(ctx, "Failed to save measurement: %v\n", err)
	}

	if h.InfluxDB != nil {
		if err := h.InfluxDB.Save(ctx, m); err != nil {
			gaelog.Errorf(ctx, "Failed to save measurement to InfluxDB: %v\n", err)
		}
	}

	w.WriteHeader(http.StatusOK)
}
