package main

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	mpbutil "github.com/mtraver/environmental-sensor/measurementpbutil"
	"github.com/mtraver/environmental-sensor/sensor"
	"google.golang.org/protobuf/proto"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

type Device interface {
	ID() string
	TelemetryTopic() string
}

type SetupJob struct {
	Sensors []string
}

func (j SetupJob) Run() {
	for _, name := range j.Sensors {
		s, err := sensor.Get(name)
		if err != nil {
			log.Printf("Error getting sensor %q: %v", name, err)
			continue
		}
		if err := s.Init(); err != nil {
			log.Printf("Failed to init %q: %v", name, err)
			continue
		}
	}
}

type SenseJob struct {
	Sensors []string
	Client  mqtt.Client
	Device  Device
	Dryrun  bool
}

func (j SenseJob) Run() {
	// Create a Measurement that we'll pass along to each sensor.
	timepb := tspb.New(time.Now().UTC())
	if err := timepb.CheckValid(); err != nil {
		log.Printf("Invalid timestamp: %v", err)
		return
	}
	m := mpb.Measurement{
		DeviceId:  j.Device.ID(),
		Timestamp: timepb,
	}

	count := 0
	for _, name := range j.Sensors {
		s, err := sensor.Get(name)
		if err != nil {
			log.Printf("Error getting sensor %q: %v", name, err)
			continue
		}
		if err := s.Sense(&m); err != nil {
			log.Printf("Failed to take measurement from %q: %v", name, err)
			continue
		}
		count++
	}

	if count <= 0 {
		log.Print("Took no measurements, will not publish")
		return
	}

	if j.Dryrun {
		log.Print(mpbutil.String(m))
	} else if err := j.publish(&m); err != nil {
		log.Printf("Failed to publish measurement: %v", err)
	}
}

func (j SenseJob) publish(m *mpb.Measurement) error {
	// Marshal to bytes for publication.
	pbBytes, err := proto.Marshal(m)
	if err != nil {
		return err
	}

	waitDur := 10 * time.Second
	token := j.Client.Publish(j.Device.TelemetryTopic(), 1, false, pbBytes)
	if ok := token.WaitTimeout(waitDur); !ok {
		// Timed out.
		return fmt.Errorf("publish timed out after %v", waitDur)
	} else if token.Error() != nil {
		// Finished before timeout but failed to publish.
		return fmt.Errorf("failed to publish: %v", token.Error())
	}

	return nil
}

type ShutdownJob struct {
	Sensors []string
}

func (j ShutdownJob) Run() {
	for _, name := range j.Sensors {
		s, err := sensor.Get(name)
		if err != nil {
			log.Printf("Error getting sensor %q: %v", name, err)
			continue
		}
		if err := s.Shutdown(); err != nil {
			log.Printf("Failed to shut down %q: %v", name, err)
			continue
		}
	}
}
