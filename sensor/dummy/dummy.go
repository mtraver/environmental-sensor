package dummy

import (
	"encoding/json"
	"log"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
)

const Name = "dummy"

type Dummy struct{}

func (d Dummy) OnRegister() error {
	log.Println("DUMMY SENSOR OnRegister")
	return nil
}

func (d Dummy) OnRemove() error {
	log.Println("DUMMY SENSOR OnRemove")
	return nil
}

func (s Dummy) Configure(raw json.RawMessage) error {
	log.Println("DUMMY SENSOR Configure")
	return nil
}

func (d Dummy) RunSetupJob() error {
	log.Println("DUMMY SENSOR RunSetupJob")
	return nil
}

func (d Dummy) RunSenseJob(m *mpb.Measurement) error {
	log.Println("DUMMY SENSOR RunSenseJob")
	m.Temp = wpb.Float(20)
	return nil
}

func (d Dummy) RunShutdownJob() error {
	log.Println("DUMMY SENSOR RunShutdownJob")
	return nil
}
