package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	aic "github.com/mtraver/awsiotcore"
	"github.com/mtraver/awsiotcore/shadow"
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
	timeout = 10 * time.Second
)

// Monitor manages the MQTT connection and the cron jobs.
type Monitor struct {
	device        *aic.Device
	config        *Config
	configVersion int
	configMu      sync.Mutex
	client        mqtt.Client
	shadowClient  *shadow.Client[*Config]
	i2cBus        i2c.BusCloser
	cron          *cron.Cron
}

// NewMonitor creates a new Monitor, connecting to the MQTT broker and starting the cron job runner.
func NewMonitor(device *aic.Device) (*Monitor, error) {
	cr := cron.New(cron.WithSeconds())
	cr.Start()

	monitor := &Monitor{
		device: device,
		// Client is initialized to a dummy client that doesn't connect to anything.
		client: mqtt.NewClient(mqtt.NewClientOptions()),
		cron:   cr,
	}

	if flagDryrun {
		return monitor, nil
	}

	log.Println("Connecting to MQTT broker...")
	client, err := device.NewClient(
		fileStoreOpt(mqttStoreDir),
		connectTimeoutOpt(timeout),
		onConnectOpt(monitor.OnConnect),
		onConnectionLostOpt(monitor.OnConnectionLost),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make MQTT client: %w", err)
	}

	// Connect to the MQTT broker.
	token := client.Connect()
	if ok := token.WaitTimeout(timeout); !ok {
		return nil, fmt.Errorf("MQTT connection attempt timed out after %v", timeout)
	}
	if token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}
	monitor.client = client

	return monitor, nil
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
	if ok := token.WaitTimeout(timeout); !ok {
		return fmt.Errorf("publish timed out after %v", timeout)
	}
	if token.Error() != nil {
		return fmt.Errorf("failed to publish: %w", token.Error())
	}

	return nil
}

func (mon *Monitor) OnConnect(client mqtt.Client) {
	log.Println("Connected to MQTT broker")

	log.Println("Subscribing to device shadow topics...")
	shadowClient, err := shadow.NewClient[*Config](client, mon.device.ID(), "", mon)
	if err != nil {
		log.Printf("Failed to make shadow client: %v", err)
		return
	}
	mon.shadowClient = shadowClient

	log.Println("Getting shadow state...")
	ctxGet, cancelCtxGet := context.WithTimeout(context.Background(), timeout)
	resp, err := shadowClient.Get(ctxGet)
	if err == nil {
		cancelCtxGet()

		log.Printf("Got config version %d:\n%s", resp.Version, resp.State.Desired)
		if err := mon.applyConfigAndReport(resp.State.Desired, resp.Version); err != nil {
			log.Printf("Failed to apply config: %v", err)
		}

		return
	}
	cancelCtxGet()

	if !errors.Is(err, shadow.ErrNotFound) {
		log.Printf("Failed to get shadow state: %v", err)
		return
	}

	log.Println("Device shadow does not exist")
	log.Println("Reporting state to create shadow")

	ctxReport, cancelCtxReport := context.WithTimeout(context.Background(), timeout)
	defer cancelCtxReport()
	_, err = shadowClient.ReportState(ctxReport, mon.config)
	if err != nil {
		log.Printf("Failed to update shadow: %v", err)
	}
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

func (mon *Monitor) applyConfig(config *Config, version int) error {
	mon.configMu.Lock()
	defer mon.configMu.Unlock()

	if version < mon.configVersion {
		return nil
	}

	if err := config.validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Determine if any sensors in the new config require I2C.
	var needI2C bool
	for _, name := range config.sensors() {
		if sensor.UsesI2C(name) {
			needI2C = true
			break
		}
	}

	// Remove jobs not present in the new config, allowing any
	// sensors that are no longer required to be removed.
	if err := mon.pauseAndRemoveOldJobs(config); err != nil {
		return err
	}

	// Open the I2C bus if required by any sensors in the new config.
	if needI2C {
		if err := mon.openI2CBus(); err != nil {
			return err
		}
	}

	// Add and remove sensors. Sensors have no dependencies on
	// other sensors so this can be done in a single step.
	if err := mon.reconcileSensors(config); err != nil {
		return err
	}

	// Close the I2C bus if it's no longer required by any sensors.
	if !needI2C {
		mon.closeI2CBus()
	}

	// Finally, add new jobs. The required sensors and system
	// resources will now be present/initialized.
	if err := mon.addNewJobs(config); err != nil {
		return err
	}

	mon.config = config
	mon.configVersion = version

	return nil
}

func (mon *Monitor) applyConfigAndReport(config *Config, version int) error {
	if err := mon.applyConfig(config, version); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if _, err := mon.shadowClient.ReportState(ctx, config); err != nil {
		return fmt.Errorf("failed to report config: %w", err)
	}

	return nil
}

func (mon *Monitor) pauseAndRemoveOldJobs(config *Config) error {
	desired := make(map[string]JobSpec)
	for _, jobSpec := range config.Jobs {
		desired[jobName(jobSpec)] = jobSpec
	}

	// Pause and remove jobs that are not present in the new config.
	for _, entry := range mon.cron.Entries() {
		if _, ok := desired[entry.Name]; !ok {
			if err := mon.cron.PauseEntryByName(entry.Name); err != nil && !errors.Is(err, cron.ErrEntryNotFound) {
				return fmt.Errorf("failed to pause job %q: %w", entry.Name, err)
			}

			mon.cron.RemoveByName(entry.Name)
			log.Printf("Removed job %q", entry.Name)
		}
	}

	return nil
}

func (mon *Monitor) addNewJobs(config *Config) error {
	desired := make(map[string]JobSpec)
	for _, jobSpec := range config.Jobs {
		desired[jobName(jobSpec)] = jobSpec
	}

	// Upsert jobs in new config.
	for name, jobSpec := range desired {
		job, err := mon.jobFromSpec(&jobSpec)
		if err != nil {
			return fmt.Errorf("failed to make job %q: %w", name, err)
		}

		if _, err := mon.cron.UpsertJob(jobSpec.Cronspec, job, cron.WithName(name)); err != nil {
			return err
		}
		log.Printf("Upserted job %q: %s, sensors %v, cronspec %q",
			name, jobSpec.Operation, jobSpec.Sensors, jobSpec.Cronspec)
	}

	return nil
}

func (mon *Monitor) reconcileSensors(config *Config) error {
	desired := make(map[string]struct{})
	for _, name := range config.sensors() {
		desired[name] = struct{}{}
	}

	// Remove sensors not in the new config.
	for name := range sensor.Names() {
		if _, ok := desired[name]; !ok {
			sensor.Remove(name)
			log.Printf("Removed sensor %q", name)
		}
	}

	// Register sensors not in the current config.
	for name := range desired {
		if sensor.Get(name) != nil {
			continue
		}

		switch name {
		case "mcp9808":
			s, err := mcp9808.New(mon.i2cBus)
			if err != nil {
				return fmt.Errorf("failed to initialize MCP9808: %w", err)
			}
			sensor.Register(name, s)

		case "sds011":
			s, err := sds011.New("/dev/ttyUSB0")
			if err != nil {
				return fmt.Errorf("failed to initialize SDS011: %w", err)
			}
			sensor.Register(name, s)

		case "dummy":
			sensor.Register(name, dummy.Dummy{})

		default:
			return fmt.Errorf("unknown sensor %q", name)
		}

		log.Printf("Registered sensor %q", name)
	}

	return nil
}

func (mon *Monitor) openI2CBus() error {
	if mon.i2cBus == nil {
		// Open default I²C bus.
		bus, err := i2creg.Open("")
		if err != nil {
			return fmt.Errorf("failed to open I²C bus: %w", err)
		}

		mon.i2cBus = bus
	}

	return nil
}

func (mon *Monitor) closeI2CBus() {
	if mon.i2cBus != nil {
		mon.i2cBus.Close()
		mon.i2cBus = nil
	}
}

func (mon *Monitor) jobFromSpec(jobSpec *JobSpec) (cron.Job, error) {
	switch jobSpec.Operation {
	case JobTypeSetup:
		return SetupJob{
			Sensors: jobSpec.Sensors,
		}, nil

	case JobTypeSense:
		return SenseJob{
			Sensors: jobSpec.Sensors,
			Publish: mon.Publish,
			Dryrun:  flagDryrun,
		}, nil

	case JobTypeShutdown:
		return ShutdownJob{
			Sensors: jobSpec.Sensors,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported operation: %v", jobSpec.Operation)
	}
}

func jobName(jobSpec JobSpec) string {
	sensors := make([]string, len(jobSpec.Sensors))
	copy(sensors, jobSpec.Sensors)
	slices.Sort(sensors)
	return fmt.Sprintf("%s/%s", jobSpec.Operation, strings.Join(sensors, ","))
}
