package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestShouldIgnoreID(t *testing.T) {
	cases := []struct {
		name     string
		ignored  map[string]struct{}
		deviceID string
		want     bool
	}{
		{
			name:     "empty",
			ignored:  map[string]struct{}{},
			deviceID: "orange",
			want:     false,
		},
		{
			name:     "empty_str",
			ignored:  map[string]struct{}{"": struct{}{}},
			deviceID: "orange",
			want:     false,
		},
		{
			name:     "allow",
			ignored:  map[string]struct{}{"kiwi": struct{}{}},
			deviceID: "orange",
			want:     false,
		},
		{
			name:     "allow_multiple",
			ignored:  map[string]struct{}{"strawberry": struct{}{}, "blueberry": struct{}{}},
			deviceID: "orange",
			want:     false,
		},
		{
			name:     "ignore",
			ignored:  map[string]struct{}{"orange": struct{}{}},
			deviceID: "orange",
			want:     true,
		},
		{
			name:     "ignore_multiple",
			ignored:  map[string]struct{}{"kiwi": struct{}{}, "strawberry": struct{}{}, "orange": struct{}{}},
			deviceID: "orange",
			want:     true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := pushHandler{IgnoredDevices: c.ignored}
			got := h.shouldIgnoreID(c.deviceID)
			if got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestShouldIgnoreSource(t *testing.T) {
	cases := []struct {
		name    string
		ignored map[string]struct{}
		source  string
		want    bool
	}{
		{
			name:    "empty",
			ignored: map[string]struct{}{},
			source:  "AWS",
			want:    false,
		},
		{
			name:    "empty_str",
			ignored: map[string]struct{}{"": struct{}{}},
			source:  "AWS",
			want:    false,
		},
		{
			name:    "allow",
			ignored: map[string]struct{}{"my_source": struct{}{}},
			source:  "AWS",
			want:    false,
		},
		{
			name:    "allow_multiple",
			ignored: map[string]struct{}{"my_source": struct{}{}, "my_src_2": struct{}{}},
			source:  "AWS",
			want:    false,
		},
		{
			name:    "ignore",
			ignored: map[string]struct{}{"AWS": struct{}{}},
			source:  "AWS",
			want:    true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := pushHandler{IgnoredSources: c.ignored}
			got := h.shouldIgnoreSource(c.source)
			if got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestUnmarshalPushRequest(t *testing.T) {
	// This example request comes from https://cloud.google.com/pubsub/docs/push#receive_push
	req := `{
  "message": {
    "attributes": {
      "key": "value"
    },
    "data": "SGVsbG8gQ2xvdWQgUHViL1N1YiEgSGVyZSBpcyBteSBtZXNzYWdlIQ==",
    "messageId": "2070443601311540",
    "message_id": "2070443601311540",
    "publishTime": "2021-02-26T19:13:55.749Z",
    "publish_time": "2021-02-26T19:13:55.749Z"
  },
  "subscription": "projects/myproject/subscriptions/mysubscription"
}`

	want := pushRequest{
		Message: pubSubMessage{
			Attributes: map[string]string{
				"key": "value",
			},
			Data:        []byte("Hello Cloud Pub/Sub! Here is my message!"),
			ID:          "2070443601311540",
			PublishTime: time.Date(2021, 2, 26, 19, 13, 55, int((749 * time.Millisecond).Nanoseconds()), time.UTC),
		},
		Subscription: "projects/myproject/subscriptions/mysubscription",
	}

	var got pushRequest
	if err := json.Unmarshal([]byte(req), &got); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("mismatch (-got +want):\n%s", diff)
	}
}
