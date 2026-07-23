package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mtraver/awsiotcore/shadow"
	"github.com/mtraver/environmental-sensor/sensor/mcp9808"
)

func TestMergeConfig(t *testing.T) {
	cases := []struct {
		name    string
		current *Config
		delta   *shadow.DeltaResponse[*Config]
		want    *Config
	}{
		{
			name: "jobs in delta",
			current: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "* * * * *",
						Operation: JobTypeSense,
						Sensors:   []string{mcp9808.Name},
					},
				},
			},
			delta: &shadow.DeltaResponse[*Config]{
				State: &Config{
					Jobs: []JobSpec{
						{
							Cronspec:  "@every 5m",
							Operation: JobTypeSense,
							Sensors:   []string{mcp9808.Name},
						},
					},
				},
			},
			want: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "@every 5m",
						Operation: JobTypeSense,
						Sensors:   []string{mcp9808.Name},
					},
				},
			},
		},
		{
			name: "jobs empty in delta",
			current: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "* * * * *",
						Operation: JobTypeSense,
						Sensors:   []string{mcp9808.Name},
					},
				},
			},
			delta: &shadow.DeltaResponse[*Config]{
				State: &Config{
					Jobs: []JobSpec{},
				},
			},
			want: &Config{
				Jobs: []JobSpec{},
			},
		},
		{
			name: "jobs nil in delta",
			current: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "* * * * *",
						Operation: JobTypeSense,
						Sensors:   []string{mcp9808.Name},
					},
				},
			},
			delta: &shadow.DeltaResponse[*Config]{
				State: &Config{},
			},
			want: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "* * * * *",
						Operation: JobTypeSense,
						Sensors:   []string{mcp9808.Name},
					},
				},
			},
		},
		{
			name: "state nil in delta",
			current: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "* * * * *",
						Operation: JobTypeSense,
						Sensors:   []string{mcp9808.Name},
					},
				},
			},
			delta: &shadow.DeltaResponse[*Config]{
				State: nil,
			},
			want: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "* * * * *",
						Operation: JobTypeSense,
						Sensors:   []string{mcp9808.Name},
					},
				},
			},
		},
		{
			name: "nil delta",
			current: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "* * * * *",
						Operation: JobTypeSense,
						Sensors:   []string{mcp9808.Name},
					},
				},
			},
			delta: nil,
			want: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "* * * * *",
						Operation: JobTypeSense,
						Sensors:   []string{mcp9808.Name},
					},
				},
			},
		},
		{
			name:    "current config empty",
			current: &Config{},
			delta: &shadow.DeltaResponse[*Config]{
				State: &Config{
					Jobs: []JobSpec{
						{
							Cronspec:  "@every 2m",
							Operation: JobTypeSense,
							Sensors:   []string{mcp9808.Name},
						},
					},
				},
			},
			want: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "@every 2m",
						Operation: JobTypeSense,
						Sensors:   []string{mcp9808.Name},
					},
				},
			},
		},
		{
			name:    "current config nil",
			current: nil,
			delta: &shadow.DeltaResponse[*Config]{
				State: &Config{
					Jobs: []JobSpec{
						{
							Cronspec:  "@every 2m",
							Operation: JobTypeSense,
							Sensors:   []string{mcp9808.Name},
						},
					},
				},
			},
			want: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "@every 2m",
						Operation: JobTypeSense,
						Sensors:   []string{mcp9808.Name},
					},
				},
			},
		},
		{
			name:    "current config and delta nil",
			current: nil,
			delta:   nil,
			want:    &Config{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeConfig(tc.current, tc.delta)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
