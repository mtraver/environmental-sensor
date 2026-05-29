package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	mpbutil "github.com/mtraver/environmental-sensor/measurementpbutil"
	"github.com/mtraver/environmental-sensor/sensor"
	"google.golang.org/protobuf/proto"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	publishWaitDuration = 10 * time.Second
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
		s, exists := sensor.Get(name)
		if !exists {
			log.Printf("Sensor not registered: %q", name)
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
	Conn    Connection
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
		Timestamp: timepb,
	}

	count := 0
	for _, name := range j.Sensors {
		s, exists := sensor.Get(name)
		if !exists {
			log.Printf("Sensor not registered: %q", name)
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
	// Set the measurement's device ID to the connection's device ID.
	m.DeviceId = j.Conn.Device.ID()

	// Marshal proto to bytes for publication.
	pbBytes, err := proto.Marshal(m)
	if err != nil {
		return err
	}

	// AWS expects Lambda function payloads to be JSON-encoded.
	jsonB, err := json.Marshal(pbBytes)
	if err != nil {
		return err
	}

	token := j.Conn.Client.Publish(j.Conn.Device.TelemetryTopic(), 1, false, jsonB)
	if ok := token.WaitTimeout(publishWaitDuration); !ok {
		// Timed out.
		return fmt.Errorf("publish timed out after %v", publishWaitDuration)
	} else if token.Error() != nil {
		// Finished before timeout but failed to publish.
		return fmt.Errorf("failed to publish: %w", token.Error())
	} else {
		log.Println("Successful publish")
	}

	return nil
}

type ShutdownJob struct {
	Sensors []string
}

func (j ShutdownJob) Run() {
	for _, name := range j.Sensors {
		s, exists := sensor.Get(name)
		if !exists {
			log.Printf("Sensor not registered: %q", name)
			continue
		}

		if err := s.Shutdown(); err != nil {
			log.Printf("Failed to shut down %q: %v", name, err)
			continue
		}
	}
}
