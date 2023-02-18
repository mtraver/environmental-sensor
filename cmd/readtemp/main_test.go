package main

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"periph.io/x/conn/v3/physic"
)

func TestToJSONProto(t *testing.T) {
	temp := 18*physic.Celsius + physic.ZeroCelsius
	want := map[string]any{
		"deviceId": "none",
		"temp":     18.0,
	}

	gotStr, err := toJSONProto(temp)
	if err != nil {
		t.Errorf("Got error, expected nil: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(gotStr), &got); err != nil {
		t.Errorf("Got error, expected nil: %v", err)
	}

	if diff := cmp.Diff(got, want, cmpopts.IgnoreMapEntries(func(k string, v any) bool { return k == "timestamp" })); diff != "" {
		t.Errorf("Unexpected result (-got +want):\n%s", diff)
	}
}
