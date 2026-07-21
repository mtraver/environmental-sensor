package main

import (
	"context"
	"log"
	"time"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	mpbutil "github.com/mtraver/environmental-sensor/measurementpbutil"
	"github.com/mtraver/environmental-sensor/sensor"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

type JobType string

const (
	JobTypeSetup    JobType = "SETUP"
	JobTypeSense            = "SENSE"
	JobTypeShutdown         = "SHUTDOWN"
)

var (
	allJobTypes = map[JobType]struct{}{
		JobTypeSetup:    struct{}{},
		JobTypeSense:    struct{}{},
		JobTypeShutdown: struct{}{},
	}
)

type JobSpec struct {
	Cronspec  string   `json:"cronspec"`
	Operation JobType  `json:"operation"`
	Sensors   []string `json:"sensors"`
}

type SetupJob struct {
	Sensors []string
}

func (j SetupJob) Run() {
	for _, name := range j.Sensors {
		s := sensor.Get(name)
		if s == nil {
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
	Publish func(context.Context, *mpb.Measurement) error
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
		s := sensor.Get(name)
		if s == nil {
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
		log.Print(mpbutil.String(&m))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := j.Publish(ctx, &m); err != nil {
		log.Printf("Failed to publish measurement: %v", err)
	} else {
		log.Println("Successful publish")
	}
}

type ShutdownJob struct {
	Sensors []string
}

func (j ShutdownJob) Run() {
	for _, name := range j.Sensors {
		s := sensor.Get(name)
		if s == nil {
			log.Printf("Sensor not registered: %q", name)
			continue
		}

		if err := s.Shutdown(); err != nil {
			log.Printf("Failed to shut down %q: %v", name, err)
			continue
		}
	}
}
