package main

import (
	"encoding/json"
	"log"

	"github.com/mtraver/awsiotcore/shadow"
)

func (mon *Monitor) HandleShadowUpdateDelta(delta *shadow.DeltaResponse[*Config]) {
	b, _ := json.MarshalIndent(delta, "", "  ")
	log.Printf("Received update delta:\n%s", b)

	merged := mergeConfig(mon.config, delta)
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

	return merged
}
