package util

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

var filterCases = []struct {
	name      string
	slice     []string
	predicate func(s string) bool
	want      []string
}{
	{
		name:      "nil",
		slice:     nil,
		predicate: filterOutEmpty,
		want:      nil,
	},
	{
		name:      "empty",
		slice:     []string{},
		predicate: filterOutEmpty,
		want:      []string{},
	},
	{
		name:      "none filtered",
		slice:     []string{"one", "two", "three"},
		predicate: filterOutEmpty,
		want:      []string{"one", "two", "three"},
	},
	{
		name:      "some filtered",
		slice:     []string{"one", "two", "", "three", "", ""},
		predicate: filterOutEmpty,
		want:      []string{"one", "two", "three"},
	},
	{
		name:      "all filtered",
		slice:     []string{"", "", "", "", ""},
		predicate: filterOutEmpty,
		want:      []string{},
	},
}

func filterOutEmpty(s string) bool {
	return s != ""
}

func TestFilter(t *testing.T) {
	for _, tc := range filterCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Filter(tc.slice, tc.predicate)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("mismatch (-got, +want):\n%s", diff)
			}
		})
	}
}

func TestFilterInPlace(t *testing.T) {
	for _, tc := range filterCases {
		t.Run(tc.name, func(t *testing.T) {
			got := FilterInPlace(tc.slice, tc.predicate)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("mismatch (-got, +want):\n%s", diff)
			}
		})
	}
}

func TestSliceToSet(t *testing.T) {
	cases := []struct {
		name  string
		slice []string
		want  map[string]struct{}
	}{
		{
			name:  "nil",
			slice: nil,
			want:  nil,
		},
		{
			name:  "empty",
			slice: []string{},
			want:  map[string]struct{}{},
		},
		{
			name:  "multiple",
			slice: []string{"", "one", "two", "two"},
			want: map[string]struct{}{
				"":    struct{}{},
				"one": struct{}{},
				"two": struct{}{},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SliceToSet(tc.slice)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("mismatch (-got, +want):\n%s", diff)
			}
		})
	}
}
