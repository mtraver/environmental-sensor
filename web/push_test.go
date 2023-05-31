package main

import (
	"testing"
)

func TestShouldIgnore(t *testing.T) {
	cases := []struct {
		name     string
		ignored  []string
		deviceID string
		want     bool
	}{
		{
			name:     "empty",
			ignored:  []string{},
			deviceID: "orange",
			want:     false,
		},
		{
			name:     "empty_str",
			ignored:  []string{""},
			deviceID: "orange",
			want:     false,
		},
		{
			name:     "allow",
			ignored:  []string{"kiwi"},
			deviceID: "orange",
			want:     false,
		},
		{
			name:     "allow_multiple",
			ignored:  []string{"strawberry", "blueberry"},
			deviceID: "orange",
			want:     false,
		},
		{
			name:     "ignore",
			ignored:  []string{"orange"},
			deviceID: "orange",
			want:     true,
		},
		{
			name:     "ignore_substr",
			ignored:  []string{"ran"},
			deviceID: "orange",
			want:     true,
		},
		{
			name:     "ignore_multiple",
			ignored:  []string{"kiwi", "strawberry", "ran"},
			deviceID: "orange",
			want:     true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := pushHandler{IgnoredDevices: c.ignored}
			got := h.shouldIgnore(c.deviceID)
			if got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}
