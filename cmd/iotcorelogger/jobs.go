package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
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
	Sensors     []string
	Connections map[ConnectionType]Connection
	Dryrun      bool
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
	var wg sync.WaitGroup

	errs := make(chan error, len(j.Connections))

	for name, conn := range j.Connections {
		// Set the measurement's device ID to that of the current connection's device.
		m.DeviceId = conn.Device.ID()

		// Marshal to bytes for publication.
		pbBytes, err := proto.Marshal(m)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func(b []byte, name ConnectionType, client mqtt.Client, topic string) {
			defer wg.Done()

			// Different connection types require different payload marshalling.
			var token mqtt.Token
			switch name {
			case GCP:
				token = client.Publish(topic, 1, false, pbBytes)
			case AWS:
				// AWS expects Lambda function payloads to be JSON-encoded.
				jsonB, err := json.Marshal(pbBytes)
				if err != nil {
					errs <- err
					return
				}

				token = client.Publish(topic, 1, false, jsonB)
			default:
				errs <- fmt.Errorf("unknown connection type %q", name)
				return
			}

			waitDur := 10 * time.Second
			if ok := token.WaitTimeout(waitDur); !ok {
				// Timed out.
				errs <- fmt.Errorf("[%s] publish timed out after %v", name, waitDur)
			} else if token.Error() != nil {
				// Finished before timeout but failed to publish.
				errs <- fmt.Errorf("[%s] failed to publish: %v", name, token.Error())
			} else {
				log.Printf("[%s] successful publish\n", name)
			}
		}(pbBytes, name, conn.Client, conn.Device.TelemetryTopic())
	}

	wg.Wait()
	close(errs)

	errSlice := []error{}
	for e := range errs {
		errSlice = append(errSlice, e)
	}

	return errors.Join(errSlice...)
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
