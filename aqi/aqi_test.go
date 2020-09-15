package aqi

import (
	"fmt"
	"testing"
)

func TestPM25(t *testing.T) {
	cases := []struct {
		pm   float32
		want int
	}{
		{-10, 0},
		{0, 0},
		{12, 50},
		{35.4, 100},
		{55.4, 150},
		{150.4, 200},
		{250.4, 300},
		{350.4, 400},
		{500.4, 500},
		{650, 500},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.pm), func(t *testing.T) {
			if got := PM25(c.pm); got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestPM10(t *testing.T) {
	cases := []struct {
		pm   float32
		want int
	}{
		{-10, 0},
		{0, 0},
		{54, 50},
		{154, 100},
		{254, 150},
		{354, 200},
		{424, 300},
		{504, 400},
		{604, 500},
		{650, 500},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.pm), func(t *testing.T) {
			if got := PM10(c.pm); got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	cases := []struct {
		aqi  int
		want string
	}{
		{-10, "Good"},
		{0, "Good"},
		{27, "Good"},
		{50, "Good"},
		{55, "Moderate"},
		{100, "Moderate"},
		{150, "Unhealthy for Sensitive Groups"},
		{200, "Unhealthy"},
		{300, "Very Unhealthy"},
		{400, "Hazardous"},
		{500, "Hazardous"},
		{650, "Hazardous"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.aqi), func(t *testing.T) {
			if got := String(c.aqi); got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestAbbrv(t *testing.T) {
	cases := []struct {
		aqi  int
		want string
	}{
		{-10, "G"},
		{0, "G"},
		{27, "G"},
		{50, "G"},
		{55, "M"},
		{100, "M"},
		{150, "USG"},
		{200, "U"},
		{300, "VU"},
		{400, "H"},
		{500, "H"},
		{650, "H"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.aqi), func(t *testing.T) {
			if got := Abbrv(c.aqi); got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}
