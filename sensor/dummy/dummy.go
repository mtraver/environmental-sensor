package dummy

import (
	"log"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
)

type Dummy struct{}

func (d Dummy) Init() error {
	log.Printf("DUMMY SENSOR INIT")
	return nil
}

func (d Dummy) Sense(m *mpb.Measurement) error {
	log.Printf("DUMMY SENSOR SENSE")
	m.Temp = wpb.Float(20)
	return nil
}

func (d Dummy) Shutdown() error {
	log.Printf("DUMMY SENSOR SHUTDOWN")
	return nil
}
