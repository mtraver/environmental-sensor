// Program iotcorelogger reads the temperature from an MCP9808 sensor and publishes
// it to Google Cloud IoT Core over MQTT.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	homedir "github.com/mitchellh/go-homedir"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"github.com/mtraver/iotcore"
	cron "github.com/robfig/cron/v3"
	"google.golang.org/protobuf/proto"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/mcp9808"
	"periph.io/x/periph/host"
)

const (
	port = 8080
)

var (
	deviceFilePath string
	caCerts        string

	numSamples     int
	sampleInterval int

	cronSpec string

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
	flag.StringVar(&deviceFilePath, "device", "", "path to a file containing a JSON-encoded Device struct (see github.com/mtraver/iotcore)")
	flag.StringVar(&caCerts, "cacerts", "", "Path to a set of trustworthy CA certs.\nDownload Google's from https://pki.google.com/roots.pem.")
	flag.IntVar(&numSamples, "numsamples", 3, "number of samples to take")
	flag.IntVar(&sampleInterval, "interval", 1, "number of seconds to wait between samples")
	flag.StringVar(&cronSpec, "cronspec", "", "cron spec that specifies when to take and publish measurements")

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

	if cronSpec == "" {
		return fmt.Errorf("cronspec flag must be given")
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

func takeMeasurement(sensor *mcp9808.Dev, device iotcore.Device) (*mpb.Measurement, error) {
	// Read the temp and construct a protobuf.
	temps, err := readTempMulti(sensor, numSamples, time.Duration(sampleInterval)*time.Second)
	if err != nil {
		return nil, err
	}
	timepb := tspb.New(time.Now().UTC())
	if err := timepb.CheckValid(); err != nil {
		return nil, err
	}
	m := &mpb.Measurement{
		DeviceId:  device.DeviceID,
		Timestamp: timepb,
		Temp:      wpb.Float(mean(temps)),
	}

	return m, nil
}

func publishMeasurement(client mqtt.Client, device iotcore.Device, m *mpb.Measurement) error {
	// Marshal to bytes for publication.
	pbBytes, err := proto.Marshal(m)
	if err != nil {
		return err
	}

	waitDur := 10 * time.Second
	token := client.Publish(device.TelemetryTopic(), 1, false, pbBytes)
	if ok := token.WaitTimeout(waitDur); !ok {
		// Timed out.
		return fmt.Errorf("publish timed out after %v", waitDur)
	} else if token.Error() != nil {
		// Finished before timeout but failed to publish.
		return fmt.Errorf("failed to publish: %v", token.Error())
	}

	return nil
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

	client, err := mqttConnect(device)
	if err != nil {
		log.Fatal(err)
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

	// Schedule the measurement publication routine.
	cr := cron.New()
	log.Printf("Starting cron scheduler with spec %q", cronSpec)
	cr.AddFunc(cronSpec, func() {
		m, err := takeMeasurement(sensor, device)
		if err != nil {
			log.Printf("Failed to take measurement: %v", err)
			return
		}

		if err := publishMeasurement(client, device, m); err != nil {
			log.Printf("Failed to publish measurement: %v", err)
		}
	})
	cr.Start()

	// Start up a web server that provides basic info about the device.
	http.Handle("/", indexHandler{
		device: device,
	})
	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
		log.Fatal(err)
	}
}
