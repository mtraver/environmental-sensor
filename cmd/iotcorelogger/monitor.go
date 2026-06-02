package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	aic "github.com/mtraver/awsiotcore"
	"github.com/mtraver/environmental-sensor/configpb"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"github.com/mtraver/environmental-sensor/sensor"
	"github.com/mtraver/environmental-sensor/sensor/dummy"
	"github.com/mtraver/environmental-sensor/sensor/mcp9808"
	"github.com/mtraver/environmental-sensor/sensor/sds011"
	cron "github.com/netresearch/go-cron"
	"google.golang.org/protobuf/proto"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
)

const (
	connectTimeout     = 10 * time.Second
	connectWaitTimeout = 15 * time.Second
	publishTimeout     = 10 * time.Second
)

// Monitor manages the MQTT connection and the cron jobs.
type Monitor struct {
	device *aic.Device
	client mqtt.Client
	i2cBus i2c.BusCloser
	cron   *cron.Cron
}

// NewMonitor creates a new Monitor, connecting to the MQTT broker and starting the cron job runner.
func NewMonitor(device *aic.Device, config *configpb.Config) (*Monitor, error) {
	cr := cron.New(cron.WithSeconds())
	cr.Start()

	monitor := &Monitor{
		device: device,
		// Client is initialized to a dummy client that doesn't connect to anything.
		client: mqtt.NewClient(mqtt.NewClientOptions()),
		cron:   cr,
	}

	if err := monitor.applyConfig(config); err != nil {
		return nil, err
	}

	if flagDryrun {
		return monitor, nil
	}

	log.Println("Connecting to MQTT broker...")
	client, err := device.NewClient(
		fileStoreOpt(mqttStoreDir),
		connectTimeoutOpt(connectTimeout),
		onConnectOpt(monitor.OnConnect),
		onConnectionLostOpt(monitor.OnConnectionLost),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make MQTT client: %w", err)
	}

	// Connect to the MQTT broker.
	token := client.Connect()
	if ok := token.WaitTimeout(connectWaitTimeout); !ok {
		return nil, fmt.Errorf("MQTT connection attempt timed out after %v", connectWaitTimeout)
	}
	if token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}
	monitor.client = client

	return monitor, nil
}

func (mon *Monitor) reconcileI2CBus(config *configpb.Config) error {
	var needI2C bool
	for _, name := range config.SupportedSensors {
		if sensor.UsesI2C(name) {
			needI2C = true
			break
		}
	}

	if needI2C && mon.i2cBus == nil {
		// Open default I²C bus.
		bus, err := i2creg.Open("")
		if err != nil {
			return fmt.Errorf("failed to open I²C bus: %w", err)
		}

		mon.i2cBus = bus
	} else if !needI2C && mon.i2cBus != nil {
		mon.i2cBus.Close()
		mon.i2cBus = nil
	}

	return nil
}

func (mon *Monitor) applyConfig(config *configpb.Config) error {
	if err := mon.reconcileI2CBus(config); err != nil {
		return err
	}

	// Register sensors.
	for _, name := range config.SupportedSensors {
		switch name {
		case "mcp9808":
			s, err := mcp9808.New(mon.i2cBus)
			if err != nil {
				return fmt.Errorf("failed to initialize MCP9808: %w", err)
			}
			sensor.Register("mcp9808", s)

		case "sds011":
			s, err := sds011.New("/dev/ttyUSB0")
			if err != nil {
				return fmt.Errorf("Failed to initialize SDS011: %w", err)
			}
			sensor.Register("sds011", s)

		case "dummy":
			sensor.Register("dummy", dummy.Dummy{})

		default:
			return fmt.Errorf("unknown sensor %q", name)
		}
	}

	// Schedule jobs defined in the config.
	for _, jpb := range config.Jobs {
		job, err := mon.jobFromConfig(jpb)
		if err != nil {
			return fmt.Errorf("failed to make job: %w", err)
		}

		log.Printf("Adding %s job for sensors %v with cronspec %q",
			configpb.Job_Operation_name[int32(jpb.Operation)], jpb.Sensors, jpb.Cronspec)
		mon.cron.AddJob(jpb.Cronspec, job)
	}

	return nil
}

func (mon *Monitor) Publish(m *mpb.Measurement) error {
	// Set the measurement's device ID to the monitor's device ID.
	m.DeviceId = mon.device.ID()

	// Marshal proto to bytes for publication.
	pbBytes, err := proto.Marshal(m)
	if err != nil {
		return err
	}

	// AWS expects Lambda function payloads to be JSON-encoded.
	jsonBytes, err := json.Marshal(pbBytes)
	if err != nil {
		return err
	}

	token := mon.client.Publish(mon.device.TelemetryTopic(), 1, false, jsonBytes)
	if ok := token.WaitTimeout(publishTimeout); !ok {
		return fmt.Errorf("publish timed out after %v", publishTimeout)
	}
	if token.Error() != nil {
		return fmt.Errorf("failed to publish: %w", token.Error())
	}

	return nil
}

func (mon *Monitor) OnConnect(client mqtt.Client) {
	log.Println("Connected to MQTT broker")
}

func (mon *Monitor) OnConnectionLost(client mqtt.Client, err error) {
	log.Printf("Connection to MQTT broker lost: %v", err)
}

func (mon *Monitor) Close() error {
	mon.cron.StopAndWait()

	mon.client.Disconnect(1000)

	if mon.i2cBus != nil {
		err := mon.i2cBus.Close()
		mon.i2cBus = nil
		return err
	}

	return nil
}

func (mon *Monitor) jobFromConfig(jpb *configpb.Job) (cron.Job, error) {
	switch jpb.Operation {
	case configpb.Job_SETUP:
		return SetupJob{
			Sensors: jpb.Sensors,
		}, nil

	case configpb.Job_SENSE:
		return SenseJob{
			Sensors: jpb.Sensors,
			Publish: mon.Publish,
			Dryrun:  flagDryrun,
		}, nil

	case configpb.Job_SHUTDOWN:
		return ShutdownJob{
			Sensors: jpb.Sensors,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported operation: %v", jpb.Operation)
	}
}
