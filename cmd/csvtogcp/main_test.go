package main

import (
	"math"
	"testing"
)

const epsilon = 0.00001

func floatsEqual(a float32, b float32) bool {
	return math.Abs(float64(a-b)) < epsilon
}

/*
 * strsToFloats
 */

func TestStrsToFloatsEmpty(t *testing.T) {
	floats, err := strsToFloats([]string{})
	if err != nil {
		t.Errorf("Error on empty list: %v", err)
	}

	if len(floats) != 0 {
		t.Errorf("Result list has len %v, expected it to be empty", len(floats))
	}
}

func TestStrsToFloatsValid(t *testing.T) {
	input := []string{"10.0", "-3.9", "0.03"}
	output := []float32{10.0, -3.9, 0.03}

	floats, err := strsToFloats(input)
	if err != nil {
		t.Errorf("Error on valid input: %v", err)
	}

	for i := range input {
		if !floatsEqual(floats[i], output[i]) {
			t.Errorf("Incorrect for input %q: Expected %v, got %v",
				input[i], output[i], floats[i])
		}
	}
}

func TestStrsToFloatsInvalid(t *testing.T) {
	_, err := strsToFloats([]string{"5.0", "spam", "spam", "spam",
		"baked beans", "spam"})
	if err == nil {
		t.Error("Expected error on invalid input, but error is nil")
	}
}

/*
 * mean
 */

func TestMeanEmpty(t *testing.T) {
	m := mean([]float32{})
	if !math.IsNaN(float64(m)) {
		t.Errorf("Expected NaN, got %v", m)
	}
}

func TestMeanSingle(t *testing.T) {
	m := mean([]float32{12.7})
	if !floatsEqual(m, 12.7) {
		t.Errorf("Expected 12.7, got %v", m)
	}
}

func TestMeanMultiple(t *testing.T) {
	m := mean([]float32{10.0, 20.0, 42.6})
	if !floatsEqual(m, 24.2) {
		t.Errorf("Expected 15.0, got %v", m)
	}
}

/*
 * lineToProto
 */

func TestLineToProtoEmpty(t *testing.T) {
	_, err := lineToProto([]string{}, "foo")
	if err == nil {
		t.Error("Expected error on invalid input, but error is nil")
	}
}

func TestLineToProtoNoMeasurements(t *testing.T) {
	_, err := lineToProto([]string{"2006-01-02T15:04:05.999999"}, "foo")
	if err == nil {
		t.Error("Expected error on invalid input, but error is nil")
	}
}

func TestLineToProtoValid(t *testing.T) {
	deviceID := "foo"
	m, err := lineToProto([]string{
		"2006-01-02T15:04:05.999999", "18.5", "18.0", "18.6"}, deviceID)
	if err != nil {
		t.Errorf("Failed to convert line: %v", err)
	}

	if m.GetDeviceId() != deviceID {
		t.Errorf("Device ID Expected to be %q, got %q", deviceID, m.GetDeviceId())
	}

	if !floatsEqual(m.GetTemp().GetValue(), 18.366667) {
		t.Errorf("Expected 18.366667, got %v", m.GetTemp())
	}
}
