package sds011

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mtraver/sds011"
)

var cmpFloats = cmpopts.EquateApprox(0, 0.0001)

func TestMean(t *testing.T) {
	cases := []struct {
		name string
		m    []sds011.Measurement
		want sds011.Measurement
	}{
		{
			name: "empty",
			m:    []sds011.Measurement{},
			want: sds011.Measurement{},
		},
		{
			name: "single",
			m: []sds011.Measurement{
				{
					PM25: 12.126,
					PM10: 25.845,
				},
			},
			want: sds011.Measurement{
				PM25: 12.126,
				PM10: 25.845,
			},
		},
		{
			name: "multiple",
			m: []sds011.Measurement{
				{
					PM25: 12.126,
					PM10: 25.845,
				},
				{
					PM25: 8.34,
					PM10: 35.9,
				},
				{
					PM25: 10.05,
					PM10: 22.98,
				},
			},
			want: sds011.Measurement{
				PM25: 10.172,
				PM10: 28.24166,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := mean(c.m)
			if diff := cmp.Diff(got, c.want, cmpFloats); diff != "" {
				t.Errorf("Unexpected result (-got +want):\n%s", diff)
			}
		})
	}
}
