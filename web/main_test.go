package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStringListToSet(t *testing.T) {
	cases := []struct {
		name string
		s    string
		want map[string]struct{}
	}{
		{
			name: "empty",
			s:    "",
			want: map[string]struct{}{},
		},
		{
			name: "only comma",
			s:    ",",
			want: map[string]struct{}{},
		},
		{
			name: "leading comma",
			s:    ",one",
			want: map[string]struct{}{"one": struct{}{}},
		},
		{
			name: "trailing comma",
			s:    "one,",
			want: map[string]struct{}{"one": struct{}{}},
		},
		{
			name: "space",
			s:    " ",
			want: map[string]struct{}{" ": struct{}{}},
		},
		{
			name: "multiple",
			s:    "one,two,three",
			want: map[string]struct{}{"one": struct{}{}, "two": struct{}{}, "three": struct{}{}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := stringListToSet(tc.s)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("mismatch (-got, +want):\n%s", diff)
			}
		})
	}
}
