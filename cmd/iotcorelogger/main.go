// Program iotcorelogger reads the temperature from an MCP9808 sensor and publishes
// it to Google Cloud IoT Core over MQTT.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/mtraver/environmental-sensor/cmd/iotcorelogger/gcpiotcore"
	"github.com/mtraver/environmental-sensor/configpb"
	"github.com/mtraver/environmental-sensor/sensor"
	"github.com/mtraver/environmental-sensor/sensor/dummy"
	"github.com/mtraver/environmental-sensor/sensor/mcp9808"
	"github.com/mtraver/environmental-sensor/sensor/sds011"
	cron "github.com/robfig/cron/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// Flags.
var (
	configFilePath    string
	gcpDeviceFilePath string
	port              int
	dryrun            bool
)

var (
	// This directory is where we'll store anything the program needs to persist, like JWTs and
	// measurements that are pending upload. This is joined with the user's home directory in init.
	dotDir = ".iotcorelogger"

	// The directory in which to store measurements that failed to publish, e.g. because
	// the network went down. It's used to configure an mqtt.NewFileStore. This is joined
	// with the user's home directory in init.
	mqttStoreDir = path.Join(dotDir, "mqtt_store")

	// This is joined with the user's home directory in init.
	jwtPath = path.Join(dotDir, "iotcorelogger.jwt")
)

func init() {
	flag.StringVar(&configFilePath, "config", "", "path to a file containing a JSON-encoded config proto")
	flag.StringVar(&gcpDeviceFilePath, "gcp-device", "", "path to a JSON file describing a GCP IoT Core device. See github.com/mtraver/iotcore.")
	flag.IntVar(&port, "port", 8080, "port on which the device's web server should listen")
	flag.BoolVar(&dryrun, "dryrun", false, "set to true to print rather than publish measurements")

	// Update directory and file paths by joining them to the user's home directory.
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("Failed to get home dir: %v", err)
	}
	dotDir = path.Join(home, dotDir)
	mqttStoreDir = path.Join(home, mqttStoreDir)
	jwtPath = path.Join(home, jwtPath)

	// Make all directories required by the program.
	dirs := []string{dotDir, mqttStoreDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			log.Fatalf("Failed to make dir %s: %v", dir, err)
		}
	}
}

func parseFlags() error {
	flag.Parse()

	if configFilePath == "" {
		return fmt.Errorf("config flag must be given")
	}

	if gcpDeviceFilePath == "" {
		return fmt.Errorf("gcp-device flag must be given")
	}

	return nil
}

func validateConfig(c *configpb.Config) error {
	if len(c.SupportedSensors) == 0 {
		return fmt.Errorf("supported_sensors must contain at least one sensor")
	}

	if len(c.Jobs) == 0 {
		return fmt.Errorf("at least one job must be given")
	}

	for _, jpb := range c.Jobs {
		if jpb.Cronspec == "" {
			return fmt.Errorf("all jobs must set cronspec")
		}

		if jpb.Operation == configpb.Job_INVALID {
			return fmt.Errorf("all jobs must set operation")
		}

		if len(jpb.Sensors) == 0 {
			return fmt.Errorf("all jobs must have at least one sensor")
		}
	}

	return nil
}

func main() {
	if err := parseFlags(); err != nil {
		fmt.Printf("argument error: %v\n", err)
		os.Exit(2)
	}

	// Parse and validate config file.
	b, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatal(err)
	}
	var config configpb.Config
	if err := protojson.Unmarshal(b, &config); err != nil {
		log.Fatal(err)
	}
	if err := validateConfig(&config); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	// Parse device file.
	device, err := gcpiotcore.ParseDeviceFile(gcpDeviceFilePath)
	if err != nil {
		log.Fatalf("Failed to parse device file: %v", err)
	}

	// Connect to IoT Core over MQTT. Make a dummy client if we're not actually
	// going to connect.
	client := mqtt.NewClient(mqtt.NewClientOptions())
	if !dryrun {
		client, err = gcpiotcore.MQTTConnect(device, jwtPath, mqttStoreDir)
		if err != nil {
			log.Fatalf("[GCP] %v", err)
		}

		// If the program is killed, disconnect from the MQTT server.
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			log.Println("Cleaning up...")
			client.Disconnect(250)
			time.Sleep(500 * time.Millisecond)
			os.Exit(1)
		}()
	}

	// Initialize periph.
	if _, err := host.Init(); err != nil {
		log.Fatalf("Failed to initialize periph: %v", err)
	}

	// Register sensors.
	for _, name := range config.SupportedSensors {
		switch name {
		case "mcp9808":
			// Open default I²C bus.
			bus, err := i2creg.Open("")
			if err != nil {
				log.Fatalf("Failed to open I²C bus: %v", err)
			}
			defer bus.Close()

			s, err := mcp9808.New(bus)
			if err != nil {
				log.Fatalf("Failed to initialize MCP9808: %v", err)
			}
			sensor.Register("mcp9808", s)
		case "sds011":
			s, err := sds011.New("/dev/ttyUSB0")
			if err != nil {
				log.Fatalf("Failed to initialize SDS011: %v", err)
			}
			sensor.Register("sds011", s)
		case "dummy":
			sensor.Register("dummy", dummy.Dummy{})
		default:
			log.Fatalf("Unknown sensor %q", name)
		}
	}

	// Schedule jobs defined in the config.
	cr := cron.New(cron.WithSeconds())
	for _, jpb := range config.Jobs {
		switch jpb.Operation {
		case configpb.Job_INVALID:
			log.Fatalf("All jobs must set operation, got %v", configpb.Job_Operation_name[int32(jpb.Operation)])
		case configpb.Job_SETUP:
			log.Printf("Adding %s job with cronspec %q", configpb.Job_Operation_name[int32(jpb.Operation)], jpb.Cronspec)
			cr.AddJob(jpb.Cronspec, SetupJob{
				Sensors: jpb.Sensors,
			})
		case configpb.Job_SENSE:
			log.Printf("Adding %s job with cronspec %q", configpb.Job_Operation_name[int32(jpb.Operation)], jpb.Cronspec)
			cr.AddJob(jpb.Cronspec, SenseJob{
				Sensors: jpb.Sensors,
				Client:  client,
				Device:  &device,
				Dryrun:  dryrun,
			})
		case configpb.Job_SHUTDOWN:
			log.Printf("Adding %s job with cronspec %q", configpb.Job_Operation_name[int32(jpb.Operation)], jpb.Cronspec)
			cr.AddJob(jpb.Cronspec, ShutdownJob{
				Sensors: jpb.Sensors,
			})
		default:
			log.Fatalf("Unknown job type %v", jpb.Operation)
		}
	}
	cr.Start()

	// Start up a web server that provides basic info about the device.
	http.Handle("/", indexHandler{
		device: device,
	})
	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
		log.Fatal(err)
	}
}
