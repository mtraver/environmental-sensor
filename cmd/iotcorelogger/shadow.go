package main

import (
	"encoding/json"
	"log"

	"github.com/google/go-cmp/cmp"
	"github.com/mtraver/awsiotcore/shadow"
)

func (mon *Monitor) HandleShadowUpdateDelta(delta *shadow.DeltaResponse[*Config]) {
	b, _ := json.MarshalIndent(delta, "", "  ")
	log.Printf("Received config update:\n%s", b)

	merged := mergeConfig(mon.config, delta)
	log.Printf("Config diff (-old, +new):\n%s", cmp.Diff(mon.config, merged))

	log.Println("Applying new config")
	if err := mon.applyConfigAndReport(merged, delta.Version); err != nil {
		log.Printf("Failed to apply config from delta: %v", err)
	}
}

func (mon *Monitor) HandleShadowUpdateDocuments(documents *shadow.DocumentsResponse[*Config]) {
	// b, _ := json.MarshalIndent(documents, "", "  ")
	// log.Printf("Received update documents:\n%s", b)
}

func mergeConfig(current *Config, delta *shadow.DeltaResponse[*Config]) *Config {
	merged := &Config{}

	if current != nil {
		merged.Jobs = current.Jobs
	}

	if delta == nil {
		return merged
	}

	if delta.State == nil {
		return merged
	}

	if delta.State.Jobs != nil {
		merged.Jobs = delta.State.Jobs
	}

	if delta.State.SensorConfig != nil {
		merged.SensorConfig = delta.State.SensorConfig
	}

	return merged
}
