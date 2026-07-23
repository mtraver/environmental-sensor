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

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	aic "github.com/mtraver/awsiotcore"
	"github.com/mtraver/awsiotcore/shadow"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"github.com/mtraver/environmental-sensor/sensor"
	"github.com/mtraver/environmental-sensor/sensor/dummy"
	"github.com/mtraver/environmental-sensor/sensor/mcp9808"
	"github.com/mtraver/environmental-sensor/sensor/sds011"
	"github.com/mtraver/environmental-sensor/sensor/sen6x"
	cron "github.com/netresearch/go-cron"
	"google.golang.org/protobuf/proto"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	periphsen6x "periph.io/x/devices/v3/sen6x"
)

const (
	timeout = 10 * time.Second
)

// Monitor manages the MQTT connection and the cron jobs.
type Monitor struct {
	device *aic.Device

	// Job config.
	configMu      sync.Mutex
	config        *Config
	configVersion int

	// Cron.
	cron *cron.Cron

	// MQTT connection.
	connMan      *autopaho.ConnectionManager
	shadowClient *shadow.Client[*Config]

	// System resources.
	i2cBus i2c.BusCloser

	// Sensors that use I2C must hold this lock for the duration of any I2C operations
	// to ensure that multiple sensors don't use the bus simultaneously.
	i2cBusMu sync.Mutex

	// Connection metrics.
	connectionMetricsMu sync.RWMutex
	firstConnectTime    *time.Time
	lastConnectTime     *time.Time
	connectionCount     int

	// Publish metrics.
	publishMetricsMu       sync.RWMutex
	lastPublishTime        *time.Time
	successfulPublishCount int
	publishFailureCount    int
}

// NewMonitor creates a new Monitor, connecting to the MQTT broker and starting the cron job runner.
func NewMonitor(ctx context.Context, device *aic.Device) (*Monitor, error) {
	cr := cron.New(cron.WithSeconds())
	cr.Start()

	monitor := &Monitor{
		device: device,
		cron:   cr,
	}

	if flagDryrun {
		return monitor, nil
	}

	log.Println("Connecting to MQTT broker...")
	clientConfig := device.ClientConfig()
	clientConfig.KeepAlive = 20
	clientConfig.SessionExpiryInterval = 5 * 60
	clientConfig.OnConnectionUp = monitor.OnConnectionUp
	clientConfig.OnConnectionDown = monitor.OnConnectionDown
	clientConfig.OnConnectError = monitor.OnConnectError
	clientConfig.ClientConfig.OnServerDisconnect = monitor.OnServerDisconnect
	clientConfig.ClientConfig.OnClientError = monitor.OnClientError

	// Connect to broker and reconnect until the context is cancelled.
	connMan, err := autopaho.NewConnection(ctx, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create MQTT connection: %w", err)
	}
	monitor.connMan = connMan

	// Wait for the connection to come up.
	if err := connMan.AwaitConnection(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", err)
	}

	return monitor, nil
}

func (mon *Monitor) Publish(ctx context.Context, m *mpb.Measurement) error {
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

	if _, err := mon.connMan.Publish(ctx, &paho.Publish{
		Topic:   mon.device.TelemetryTopic(),
		Payload: jsonBytes,
		QoS:     1,
	}); err != nil {
		mon.publishMetricsMu.Lock()
		defer mon.publishMetricsMu.Unlock()
		mon.publishFailureCount += 1

		return fmt.Errorf("failed to publish: %w", err)
	}

	mon.publishMetricsMu.Lock()
	defer mon.publishMetricsMu.Unlock()
	now := time.Now().UTC()
	mon.lastPublishTime = &now
	mon.successfulPublishCount += 1

	return nil
}

func (mon *Monitor) OnConnectionUp(client *autopaho.ConnectionManager, connAck *paho.Connack) {
	log.Println("Connected to MQTT broker")

	now := time.Now().UTC()

	mon.connectionMetricsMu.Lock()
	if mon.firstConnectTime == nil {
		mon.firstConnectTime = &now
	}
	mon.lastConnectTime = &now
	mon.connectionCount += 1
	mon.connectionMetricsMu.Unlock()

	// Create a shadow client if we don't have one already.
	if mon.shadowClient == nil {
		mon.shadowClient = shadow.NewClient[*Config](client, mon.device.ID(), "", mon)
	}

	log.Println("Subscribing to device shadow topics...")
	ctxShadow, cancelCtxShadow := context.WithTimeout(context.Background(), timeout)
	defer cancelCtxShadow()
	if err := mon.shadowClient.OnConnectionUp(ctxShadow); err != nil {
		log.Printf("Failed to subscribe to shadow topics: %v", err)
		return
	}

	log.Println("Getting shadow state...")
	ctxGet, cancelCtxGet := context.WithTimeout(context.Background(), timeout)
	resp, err := mon.shadowClient.Get(ctxGet)
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
	_, err = mon.shadowClient.ReportState(ctxReport, mon.config)
	if err != nil {
		log.Printf("Failed to update shadow: %v", err)
	}
}

func (mon *Monitor) OnConnectionDown() bool {
	log.Printf("Connection to MQTT broker lost")

	mon.connectionMetricsMu.Lock()
	defer mon.connectionMetricsMu.Unlock()
	mon.lastConnectTime = nil

	// Return true so that a reconnect is attempted.
	return true
}

func (mon *Monitor) OnConnectError(err error) {
	log.Printf("Error while attempting to connect to MQTT broker: %v", err)
}

func (mon *Monitor) OnServerDisconnect(d *paho.Disconnect) {
	if d.Properties != nil {
		log.Printf("Server requested disconnect: %s", d.Properties.ReasonString)
	} else {
		log.Printf("Server requested disconnect: reason code %d", d.ReasonCode)
	}

	mon.connectionMetricsMu.Lock()
	defer mon.connectionMetricsMu.Unlock()
	mon.lastConnectTime = nil
}

func (mon *Monitor) OnClientError(err error) {
	log.Printf("Client error: %v", err)
}

func (mon *Monitor) Close(ctx context.Context) error {
	mon.cron.StopAndWait()

	// TODO(mtraver) Make this unconditional if/when I remove -dryrun.
	if mon.connMan != nil {
		mon.connMan.Disconnect(ctx)

		select {
		case <-mon.connMan.Done():
		case <-ctx.Done():
		}
	}

	if err := sensor.RemoveAll(); err != nil {
		return err
	}

	if err := mon.closeI2CBus(); err != nil {
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
	for _, name := range sensor.Names() {
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
			s, err := mcp9808.New(mon.i2cBus, &mon.i2cBusMu)
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

		case "sen6x":
			// TODO(mtraver) Provide model via config rather than hard-coding SEN66.
			s, err := sen6x.New(periphsen6x.SEN66, mon.i2cBus, &mon.i2cBusMu)
			if err != nil {
				return fmt.Errorf("failed to initialize SEN6x: %w", err)
			}
			log.Printf("Current SEN6x config:\n%s", s.CurrentConfig())

			sensor.Register(name, s)

		case "dummy":
			sensor.Register(name, dummy.Dummy{})

		default:
			return fmt.Errorf("unknown sensor %q", name)
		}

		log.Printf("Registered sensor %q", name)
	}

	// Apply sensor-specific config.
	for name := range desired {
		if err := sensor.Get(name).Configure(config.SensorConfig[name]); err != nil {
			return fmt.Errorf("failed to configure %q: %w", name, err)
		}
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

func (mon *Monitor) closeI2CBus() error {
	if mon.i2cBus != nil {
		err := mon.i2cBus.Close()
		mon.i2cBus = nil

		return err
	}

	return nil
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
