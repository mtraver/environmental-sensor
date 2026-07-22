package dummy

import (
	"log"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
)

type Dummy struct{}

func (d Dummy) OnRegister() error {
	log.Printf("DUMMY SENSOR OnRegister")
	return nil
}

func (d Dummy) OnRemove() error {
	log.Printf("DUMMY SENSOR OnRemove")
	return nil
}

func (d Dummy) RunSetupJob() error {
	log.Printf("DUMMY SENSOR RunSetupJob")
	return nil
}

func (d Dummy) RunSenseJob(m *mpb.Measurement) error {
	log.Printf("DUMMY SENSOR RunSenseJob")
	m.Temp = wpb.Float(20)
	return nil
}

func (d Dummy) RunShutdownJob() error {
	log.Printf("DUMMY SENSOR RunShutdownJob")
	return nil
}
