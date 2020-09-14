package dummy

import (
	"log"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
)

type Dummy struct{}

func (d Dummy) Init() error {
	log.Printf("DUMMY SENSOR INIT")
	return nil
}

func (d Dummy) Sense(m *mpb.Measurement) error {
	log.Printf("DUMMY SENSOR SENSE")
	return nil
}

func (d Dummy) Shutdown() error {
	log.Printf("DUMMY SENSOR SHUTDOWN")
	return nil
}
