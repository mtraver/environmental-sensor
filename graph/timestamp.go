package graph

import (
	"time"
)

const (
	gqlTimeLayout = "2006-01-02T15:04:05.999Z"
)

func gqlTimestampToTime(timestamp string) (time.Time, error) {
	return time.Parse(gqlTimeLayout, timestamp)
}

func timeToGQLTimestamp(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.Format(gqlTimeLayout)
}
