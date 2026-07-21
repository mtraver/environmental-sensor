package measurement

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mtraver/environmental-sensor/metric"
)

var cmpFloats = cmpopts.EquateApprox(0, 0.0001)

func TestMean(t *testing.T) {
	cases := []struct {
		name string
		sms  []StorableMeasurement
		want map[metric.Key]float32
	}{
		{
			name: "empty",
			sms:  []StorableMeasurement{},
			want: map[metric.Key]float32{},
		},
		{
			name: "single_measurement",
			sms: []StorableMeasurement{
				{
					Temp: floatPtr(18.3),
					PM25: floatPtr(12.1),
					PM10: floatPtr(20.7),
					RH:   floatPtr(55.0),
				},
			},
			want: map[metric.Key]float32{
				metric.Temp: 18.3,
				metric.PM25: 12.1,
				metric.PM10: 20.7,
				metric.RH:   55.0,
			},
		},
		{
			name: "multiple_measurements",
			sms: []StorableMeasurement{
				{
					Temp: floatPtr(18.3),
					PM25: floatPtr(12.1),
					PM10: floatPtr(20.7),
					RH:   floatPtr(55.0),
				},
				{
					Temp: floatPtr(19.0),
					PM25: floatPtr(8.4),
					PM10: floatPtr(21.9),
					RH:   floatPtr(33.0),
				},
			},
			want: map[metric.Key]float32{
				metric.Temp: 18.65,
				metric.PM25: 10.25,
				metric.PM10: 21.3,
				metric.RH:   44.0,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Mean(tc.sms)
			if diff := cmp.Diff(got, tc.want, cmpFloats); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestStdDev(t *testing.T) {
	cases := []struct {
		name string
		sms  []StorableMeasurement
		want map[metric.Key]float32
	}{
		{
			name: "empty",
			sms:  []StorableMeasurement{},
			want: map[metric.Key]float32{},
		},
		{
			name: "single_measurement",
			sms: []StorableMeasurement{
				{
					Temp: floatPtr(18.3),
					PM25: floatPtr(12.1),
					PM10: floatPtr(20.7),
					RH:   floatPtr(55.0),
				},
			},
			want: map[metric.Key]float32{
				metric.Temp: 0.0,
				metric.PM25: 0.0,
				metric.PM10: 0.0,
				metric.RH:   0.0,
			},
		},
		{
			name: "multiple_measurements",
			sms: []StorableMeasurement{
				{
					Temp: floatPtr(18.3),
					PM25: floatPtr(12.1),
					PM10: floatPtr(20.7),
					RH:   floatPtr(55.0),
				},
				{
					Temp: floatPtr(19.0),
					PM25: floatPtr(8.4),
					PM10: floatPtr(21.9),
					RH:   floatPtr(33.0),
				},
				{
					Temp: floatPtr(25.85),
					PM25: floatPtr(17.7),
					PM10: floatPtr(28.0),
					RH:   floatPtr(47.3),
				},
				{
					Temp: floatPtr(12.2),
					PM25: floatPtr(19.4),
					PM10: floatPtr(26.2),
					RH:   floatPtr(89.1),
				},
			},
			want: map[metric.Key]float32{
				metric.Temp: 4.83598,
				metric.PM25: 4.39261,
				metric.PM10: 2.99917,
				metric.RH:   20.6232,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StdDev(tc.sms)
			if diff := cmp.Diff(got, tc.want, cmpFloats); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestMin(t *testing.T) {
	cases := []struct {
		name string
		sms  []StorableMeasurement
		want map[metric.Key]float32
	}{
		{
			name: "empty",
			sms:  []StorableMeasurement{},
			want: map[metric.Key]float32{},
		},
		{
			name: "single_measurement",
			sms: []StorableMeasurement{
				{
					Temp: floatPtr(18.3),
					PM25: floatPtr(12.1),
					PM10: floatPtr(20.7),
					RH:   floatPtr(55.0),
				},
			},
			want: map[metric.Key]float32{
				metric.Temp: 18.3,
				metric.PM25: 12.1,
				metric.PM10: 20.7,
				metric.RH:   55.0,
			},
		},
		{
			name: "multiple_measurements",
			sms: []StorableMeasurement{
				{
					Temp: floatPtr(18.3),
					PM25: floatPtr(12.1),
					PM10: floatPtr(20.7),
					RH:   floatPtr(55.0),
				},
				{
					Temp: floatPtr(19.0),
					PM25: floatPtr(8.4),
					PM10: floatPtr(21.9),
					RH:   floatPtr(33.0),
				},
			},
			want: map[metric.Key]float32{
				metric.Temp: 18.3,
				metric.PM25: 8.4,
				metric.PM10: 20.7,
				metric.RH:   33.0,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Min(tc.sms)
			if diff := cmp.Diff(got, tc.want, cmpFloats); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestMax(t *testing.T) {
	cases := []struct {
		name string
		sms  []StorableMeasurement
		want map[metric.Key]float32
	}{
		{
			name: "empty",
			sms:  []StorableMeasurement{},
			want: map[metric.Key]float32{},
		},
		{
			name: "single_measurement",
			sms: []StorableMeasurement{
				{
					Temp: floatPtr(18.3),
					PM25: floatPtr(12.1),
					PM10: floatPtr(20.7),
					RH:   floatPtr(55.0),
				},
			},
			want: map[metric.Key]float32{
				metric.Temp: 18.3,
				metric.PM25: 12.1,
				metric.PM10: 20.7,
				metric.RH:   55.0,
			},
		},
		{
			name: "multiple_measurements",
			sms: []StorableMeasurement{
				{
					Temp: floatPtr(18.3),
					PM25: floatPtr(12.1),
					PM10: floatPtr(20.7),
					RH:   floatPtr(55.0),
				},
				{
					Temp: floatPtr(19.0),
					PM25: floatPtr(8.4),
					PM10: floatPtr(21.9),
					RH:   floatPtr(33.0),
				},
			},
			want: map[metric.Key]float32{
				metric.Temp: 19.0,
				metric.PM25: 12.1,
				metric.PM10: 21.9,
				metric.RH:   55.0,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Max(tc.sms)
			if diff := cmp.Diff(got, tc.want, cmpFloats); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
