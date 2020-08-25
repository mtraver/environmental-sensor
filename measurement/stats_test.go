package measurement

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var cmpFloats = cmpopts.EquateApprox(0, 0.0001)

func TestMean(t *testing.T) {
	cases := []struct {
		name string
		sms  []StorableMeasurement
		want map[string]float32
	}{
		{
			name: "empty",
			sms:  []StorableMeasurement{},
			want: map[string]float32{},
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
			want: map[string]float32{
				"temp":  18.3,
				"PM2.5": 12.1,
				"PM10":  20.7,
				"RH":    55.0,
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
			want: map[string]float32{
				"temp":  18.65,
				"PM2.5": 10.25,
				"PM10":  21.3,
				"RH":    44.0,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Mean(tc.sms)
			if diff := cmp.Diff(got, tc.want, cmpFloats); diff != "" {
				t.Errorf("Unexpected result (-got +want):\n%s", diff)
			}
		})
	}
}

func TestStdDev(t *testing.T) {
	cases := []struct {
		name string
		sms  []StorableMeasurement
		want map[string]float32
	}{
		{
			name: "empty",
			sms:  []StorableMeasurement{},
			want: map[string]float32{},
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
			want: map[string]float32{
				"temp":  0.0,
				"PM2.5": 0.0,
				"PM10":  0.0,
				"RH":    0.0,
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
			want: map[string]float32{
				"temp":  4.83598,
				"PM2.5": 4.39261,
				"PM10":  2.99917,
				"RH":    20.6232,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StdDev(tc.sms)
			if diff := cmp.Diff(got, tc.want, cmpFloats); diff != "" {
				t.Errorf("Unexpected result (-got +want):\n%s", diff)
			}
		})
	}
}

func TestMin(t *testing.T) {
	cases := []struct {
		name string
		sms  []StorableMeasurement
		want map[string]float32
	}{
		{
			name: "empty",
			sms:  []StorableMeasurement{},
			want: map[string]float32{},
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
			want: map[string]float32{
				"temp":  18.3,
				"PM2.5": 12.1,
				"PM10":  20.7,
				"RH":    55.0,
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
			want: map[string]float32{
				"temp":  18.3,
				"PM2.5": 8.4,
				"PM10":  20.7,
				"RH":    33.0,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Min(tc.sms)
			if diff := cmp.Diff(got, tc.want, cmpFloats); diff != "" {
				t.Errorf("Unexpected result (-got +want):\n%s", diff)
			}
		})
	}
}

func TestMax(t *testing.T) {
	cases := []struct {
		name string
		sms  []StorableMeasurement
		want map[string]float32
	}{
		{
			name: "empty",
			sms:  []StorableMeasurement{},
			want: map[string]float32{},
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
			want: map[string]float32{
				"temp":  18.3,
				"PM2.5": 12.1,
				"PM10":  20.7,
				"RH":    55.0,
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
			want: map[string]float32{
				"temp":  19.0,
				"PM2.5": 12.1,
				"PM10":  21.9,
				"RH":    55.0,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Max(tc.sms)
			if diff := cmp.Diff(got, tc.want, cmpFloats); diff != "" {
				t.Errorf("Unexpected result (-got +want):\n%s", diff)
			}
		})
	}
}
