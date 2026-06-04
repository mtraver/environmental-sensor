package main

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConfigSensors(t *testing.T) {
	cases := []struct {
		name   string
		config *Config
		want   []string
	}{
		{
			name:   "nil jobs",
			config: &Config{},
			want:   []string{},
		},
		{
			name: "single job single sensor",
			config: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "@every 2m",
						Operation: JobTypeSense,
						Sensors:   []string{"mcp9808"},
					},
				},
			},
			want: []string{"mcp9808"},
		},
		{
			name: "single job multiple sensors",
			config: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "@every 2m",
						Operation: JobTypeSense,
						Sensors:   []string{"mcp9808", "sds011"},
					},
				},
			},
			want: []string{"mcp9808", "sds011"},
		},
		{
			name: "multiple jobs distinct sensors",
			config: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "@every 2m",
						Operation: JobTypeSense,
						Sensors:   []string{"mcp9808"},
					},
					{
						Cronspec:  "@every 2m",
						Operation: JobTypeSense,
						Sensors:   []string{"sds011"},
					},
				},
			},
			want: []string{"mcp9808", "sds011"},
		},
		{
			name: "multiple jobs overlapping sensors",
			config: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "35 1-59/2 * * * *",
						Operation: JobTypeSetup,
						Sensors:   []string{"sds011"},
					},
					{
						Cronspec:  "0 0-59/2 * * * *",
						Operation: JobTypeSense,
						Sensors:   []string{"mcp9808", "sds011"},
					},
					{
						Cronspec:  "8 0-59/2 * * * *",
						Operation: JobTypeShutdown,
						Sensors:   []string{"sds011"},
					},
				},
			},
			want: []string{"mcp9808", "sds011"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.config.sensors()
			slices.Sort(got)
			slices.Sort(tc.want)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: false,
		},
		{
			name:    "empty jobs",
			config:  &Config{},
			wantErr: false,
		},
		{
			name: "single valid job",
			config: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "@every 2m",
						Operation: JobTypeSense,
						Sensors:   []string{"mcp9808"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple valid jobs",
			config: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "@every 2m",
						Operation: JobTypeSetup,
						Sensors:   []string{"mcp9808"},
					},
					{
						Cronspec:  "@every 5m",
						Operation: JobTypeSense,
						Sensors:   []string{"mcp9808", "sds011"},
					},
					{
						Cronspec:  "@every 6m",
						Operation: JobTypeShutdown,
						Sensors:   []string{"mcp9808"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing cronspec",
			config: &Config{
				Jobs: []JobSpec{
					{
						Operation: JobTypeSense,
						Sensors:   []string{"mcp9808"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid operation",
			config: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "@every 2m",
						Operation: "INVALID",
						Sensors:   []string{"mcp9808"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty operation",
			config: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "@every 2m",
						Operation: "",
						Sensors:   []string{"mcp9808"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "no sensors",
			config: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "@every 2m",
						Operation: JobTypeSense,
						Sensors:   []string{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate sensors",
			config: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "@every 2m",
						Operation: JobTypeSense,
						Sensors:   []string{"mcp9808", "mcp9808"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "second job invalid cronspec",
			config: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "@every 2m",
						Operation: JobTypeSense,
						Sensors:   []string{"mcp9808"},
					},
					{
						Operation: JobTypeSense,
						Sensors:   []string{"sds011"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "second job duplicate sensors",
			config: &Config{
				Jobs: []JobSpec{
					{
						Cronspec:  "@every 2m",
						Operation: JobTypeSense,
						Sensors:   []string{"mcp9808"},
					},
					{
						Cronspec:  "@every 5m",
						Operation: JobTypeSense,
						Sensors:   []string{"sds011", "sds011"},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.validate()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			} else if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
