// Program iotcorelogger reads from sensors and publishes the measurements to AWS IoT Core over MQTT.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	aic "github.com/mtraver/awsiotcore"
	"github.com/mtraver/environmental-sensor/configpb"
	"google.golang.org/protobuf/encoding/protojson"
	"periph.io/x/host/v3"
)

var (
	flagConfigFilePath    string
	flagAWSDeviceFilePath string
	flagPort              int
	flagDryrun            bool
)

var (
	// This directory is where we'll store anything the program needs to persist, like JWTs and
	// measurements that are pending upload. This is joined with the user's home directory in init.
	dotDir = ".iotcorelogger"

	// The directory in which to store measurements that failed to publish, e.g. because
	// the network went down. It's used to configure an mqtt.NewFileStore. This is joined
	// with the user's home directory in init.
	mqttStoreDir = path.Join(dotDir, "mqtt_store")
)

func init() {
	flag.StringVar(&flagConfigFilePath, "config", "", "path to a file containing a JSON-encoded config proto")
	flag.StringVar(&flagAWSDeviceFilePath, "aws-device", "", "path to a device config file describing an AWS IoT Core device")
	flag.IntVar(&flagPort, "port", 8080, "port on which the device's web server should listen")
	flag.BoolVar(&flagDryrun, "dryrun", false, "set to true to print rather than publish measurements")

	flag.Usage = func() {
		message := `usage: iotcorelogger [options]

Options:
`

		fmt.Fprint(flag.CommandLine.Output(), message)
		flag.PrintDefaults()
	}

	// Update directory and file paths by joining them to the user's home directory.
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("Failed to get home dir: %v", err)
	}
	dotDir = path.Join(home, dotDir)
	mqttStoreDir = path.Join(home, mqttStoreDir)

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

	if flagConfigFilePath == "" {
		return errors.New("-config must be given")
	}

	if flagAWSDeviceFilePath == "" {
		return errors.New("-aws-device must be given")
	}

	return nil
}

func parseDeviceFile(filepath string) (aic.Device, error) {
	b, err := os.ReadFile(filepath)
	if err != nil {
		return aic.Device{}, err
	}

	var device aic.Device
	if err := json.Unmarshal(b, &device); err != nil {
		return aic.Device{}, err
	}

	if device.DeviceID == "" {
		deviceID, err := aic.DeviceIDFromCert(device.CertPath)
		if err != nil {
			return aic.Device{}, err
		}
		device.DeviceID = deviceID
	}

	return device, nil
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
		flag.Usage()
		os.Exit(2)
	}

	// Parse and validate config file.
	b, err := os.ReadFile(flagConfigFilePath)
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
	device, err := parseDeviceFile(flagAWSDeviceFilePath)
	if err != nil {
		log.Fatalf("Failed to parse AWS device file: %v", err)
	}

	// Initialize periph.
	if _, err := host.Init(); err != nil {
		log.Fatalf("Failed to initialize periph: %v", err)
	}

	monitor, err := NewMonitor(&device, &config)
	if err != nil {
		log.Fatal(err)
	}

	// Start up a web server that provides basic info about the device.
	http.Handle("/{$}", monitor)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", flagPort), nil); err != nil {
		log.Fatal(err)
	}
}
