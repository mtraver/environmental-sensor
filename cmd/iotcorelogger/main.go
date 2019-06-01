// Program iotcorelogger reads the temperature from an MCP9808 sensor and publishes
// it to Google Cloud IoT Core over MQTT.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/mtraver/iotcore"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/mcp9808"
	"periph.io/x/periph/host"

	"github.com/mtraver/environmental-sensor/cmd/iotcorelogger/pending"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
)

var (
	deviceFilePath string
	caCerts        string

	numSamples     int
	sampleInterval int

	// This directory is where we'll store anything the program needs to persist, like JWTs and
	// measurements that are pending upload. This is joined with the user's home directory in init.
	dotDir = ".iotcorelogger"

	// The directory in which to store measurements that failed to publish, e.g. because
	// the network went down. Publication will be retried later. This is joined with the
	// user's home directory in init.
	pendingDir = path.Join(dotDir, "pending")

	// This is joined with the user's home directory in init.
	jwtPath = path.Join(dotDir, "iotcorelogger.jwt")
)

// We don't currently do anything with configs from the server
// func configHandler(client mqtt.Client, msg mqtt.Message) {
// 	log.Printf("config handler: topic: %v\n", msg.Topic())
// 	log.Printf("config handler: tayload: %v\n", msg.Payload())
// }

func init() {
	flag.StringVar(&deviceFilePath, "device", "", "path to a file containing a JSON-encoded Device struct (see github.com/mtraver/iotcore)")
	flag.StringVar(&caCerts, "cacerts", "", "Path to a set of trustworthy CA certs.\nDownload Google's from https://pki.google.com/roots.pem.")
	flag.IntVar(&numSamples, "numsamples", 3, "number of samples to take")
	flag.IntVar(&sampleInterval, "interval", 1, "number of seconds to wait between samples")

	// Update directory and file paths by joining them to the user's home directory.
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("Failed to get home dir: %v", err)
	}
	dotDir = path.Join(home, dotDir)
	pendingDir = path.Join(home, pendingDir)
	jwtPath = path.Join(home, jwtPath)

	// Make all directories required by the program.
	dirs := []string{dotDir, pendingDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			log.Fatalf("Failed to make dir %s: %v", dir, err)
		}
	}
}

func parseFlags() error {
	flag.Parse()

	if deviceFilePath == "" {
		return fmt.Errorf("device flag must be given")
	}

	if caCerts == "" {
		return fmt.Errorf("cacerts flag must be given")
	}

	if numSamples <= 0 {
		return fmt.Errorf("numsamples must be > 0")
	}

	if sampleInterval <= 0 {
		return fmt.Errorf("interval must be > 0")
	}

	return nil
}

func mean(s []physic.Temperature) float32 {
	var sum float64
	for _, t := range s {
		sum += t.Celsius()
	}

	return float32(sum / float64(len(s)))
}

func save(m *mpb.Measurement) {
	if err := pending.Save(m, pendingDir); err != nil {
		log.Printf("Failed to save measurement: %v", err)
	}
}

func parseDeviceFile(filepath string) (iotcore.Device, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return iotcore.Device{}, err
	}

	var device iotcore.Device
	if err := json.Unmarshal(b, &device); err != nil {
		return iotcore.Device{}, err
	}

	return device, nil
}

func main() {
	if err := parseFlags(); err != nil {
		fmt.Printf("argument error: %v\n", err)
		os.Exit(2)
	}

	device, err := parseDeviceFile(deviceFilePath)
	if err != nil {
		log.Fatalf("Failed to parse device file: %v", err)
	}

	certsFile, err := os.Open(caCerts)
	if err != nil {
		log.Fatalf("Failed to open certs file: %v", err)
	}
	defer certsFile.Close()

	client, err := device.NewClient(iotcore.DefaultBroker, certsFile, iotcore.PersistentlyCacheJWT(60*time.Minute, jwtPath))
	if err != nil {
		log.Fatalf("Failed to make MQTT client: %v", err)
	}

	// Initialize periph.
	if _, err := host.Init(); err != nil {
		log.Fatalf("Failed to initialize periph: %v", err)
	}

	// Open default I²C bus.
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatalf("Failed to open I²C bus: %v", err)
	}
	defer bus.Close()

	sensor, err := mcp9808.New(bus, &mcp9808.DefaultOpts)
	if err != nil {
		log.Fatalf("Failed to initialize MCP9808: %v", err)
	}

	// Read the temp, construct a protobuf, and marshal it to bytes.
	temps, err := readTempMulti(sensor, numSamples, time.Duration(sampleInterval)*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	timepb, err := ptypes.TimestampProto(time.Now().UTC())
	if err != nil {
		log.Fatal(err)
	}
	m := &mpb.Measurement{
		DeviceId:  device.DeviceID,
		Timestamp: timepb,
		Temp:      mean(temps),
	}
	pbBytes, err := proto.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}

	// Attempt to connect using the MQTT client. If it fails, save the Measurement for later publication.
	waitDur := 10 * time.Second
	token := client.Connect()
	if ok := token.WaitTimeout(waitDur); !ok {
		// Timed out. Save the Measurement to retry later.
		save(m)

		log.Fatalf("MQTT connection attempt timed out after %v", waitDur)
	} else if token.Error() != nil {
		// Finished before the timeout but failed to connect. Save the Measurement to retry later.
		save(m)

		log.Fatalf("Failed to connect to MQTT server: %v", token.Error())
	}

	// We don't currently do anything with configs from the server.
	// client.Subscribe(device.ConfigTopic(), 1, configHandler)

	token = client.Publish(device.TelemetryTopic(), 1, false, pbBytes)
	if ok := token.WaitTimeout(waitDur); !ok {
		// Timed out. Save the Measurement to retry later.
		log.Printf("Publish timed out after %v", waitDur)
		save(m)
	} else if token.Error() != nil {
		// Finished before timeout but failed to publish. Save the Measurement to retry later.
		log.Printf("Failed to publish: %v", token.Error())
		save(m)
	} else {
		// Publish succeeded, so attempt to publish any pending measurements.
		if err := pending.PublishAll(client, device.TelemetryTopic(), pendingDir); err != nil {
			log.Printf("Failed to publish all pending measurements: %v", err)
		}
	}

	client.Disconnect(250)

	time.Sleep(500 * time.Millisecond)
}
