package main

import (
	"testing"
	"time"
)

func TestTimeAgoString(t *testing.T) {
	cases := []struct {
		durAgo   time.Duration
		expected string
	}{
		{time.Second * 1, "just now"},
		{time.Second * 4, "just now"},
		{time.Second * 5, "5 s ago"},
		{time.Second * 6, "5 s ago"},
		{time.Second * 11, "10 s ago"},
		{time.Minute * 1, "1 min ago"},
		{time.Minute * 12, "12 min ago"},
		{time.Hour, "1 hr ago"},
		{time.Hour * 5, "5 hr ago"},
		{time.Hour + time.Minute*3, "1 hr 3 min ago"},
		{time.Hour * 24, "> 24 hr ago"},
	}

	for _, c := range cases {
		if s := timeAgoString(time.Now().UTC().Add(-c.durAgo)); s != c.expected {
			t.Errorf("Got %q, expected %q", s, c.expected)
		}
	}
}
