// Program iotcorelogger reads the temperature from an MCP9808 sensor and publishes
// it to Google Cloud IoT Core over MQTT.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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

const (
	defaultRegion = "us-central1"
	defaultHost   = "mqtt.googleapis.com"

	certExtension = ".x509"
)

var (
	projectID   string
	registryID  string
	region      string
	privKeyPath string

	broker iotcore.MQTTBroker

	caCerts string

	numSamples     int
	sampleInterval int

	// The allowed ports for connecting to the IoT Core MQTT broker
	allowedPorts = map[int]bool{
		8883: true,
		443:  true,
	}

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
	flag.StringVar(&projectID, "project", "", "Google Cloud Platform project ID")
	flag.StringVar(&registryID, "registry", "", "Google Cloud IoT Core registry ID")
	flag.StringVar(&region, "region", defaultRegion, "Google Cloud Platform region")
	flag.StringVar(&privKeyPath, "key", "", "path to device's private key")
	flag.StringVar(&caCerts, "cacerts", "",
		"Path to a set of trustworthy CA certs.\n"+
			"Download Google's from https://pki.google.com/roots.pem.")

	flag.StringVar(&broker.Host, "mqtthost", defaultHost, "MQTT host")
	flag.IntVar(&broker.Port, "mqttport", 8883, "MQTT port")

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

	if projectID == "" {
		return fmt.Errorf("project flag must be given")
	}

	if registryID == "" {
		return fmt.Errorf("registry flag must be given")
	}

	if region == "" {
		return fmt.Errorf("region flag must be given")
	}

	if privKeyPath == "" {
		return fmt.Errorf("key flag must be given")
	}

	if caCerts == "" {
		return fmt.Errorf("cacerts flag must be given")
	}

	if broker.Host == "" {
		return fmt.Errorf("mqtthost flag must be given")
	}

	if _, ok := allowedPorts[broker.Port]; !ok {
		return fmt.Errorf("mqttport must be one of %v", allowedPorts)
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

func certPath(keyPath string) string {
	ext := path.Ext(keyPath)
	return keyPath[:len(keyPath)-len(ext)] + certExtension
}

func existingJWT(device iotcore.Device) (string, error) {
	f, err := os.Open(jwtPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}

		// There is no existing JWT.
		return "", fmt.Errorf("%s does not exist", jwtPath)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	jwt := string(b)
	if ok, err := device.VerifyJWT(jwt); !ok {
		return "", err
	}

	return jwt, nil
}

func newClient(device iotcore.Device) (mqtt.Client, error) {
	mqttOptions, err := iotcore.NewMQTTOptions(device, broker, caCerts)
	if err != nil {
		return nil, err
	}

	jwt, err := existingJWT(device)
	if err != nil {
		jwt, err = device.NewJWT(60 * time.Minute)
		if err != nil {
			return nil, err
		}

		// Persist the JWT.
		if err := ioutil.WriteFile(jwtPath, []byte(jwt), 0600); err != nil {
			log.Printf("Failed to save JWT to %s: %v", jwtPath, err)
		}
	}
	mqttOptions.SetPassword(jwt)

	return mqtt.NewClient(mqttOptions), nil
}

func save(m *mpb.Measurement) {
	if err := pending.Save(m, pendingDir); err != nil {
		log.Printf("Failed to save measurement: %v", err)
	}
}

func main() {
	if err := parseFlags(); err != nil {
		fmt.Printf("argument error: %v\n", err)
		os.Exit(2)
	}

	deviceID, err := iotcore.DeviceIDFromCert(certPath(privKeyPath))
	if err != nil {
		log.Fatal(err)
	}

	device := iotcore.Device{
		ProjectID:   projectID,
		RegistryID:  registryID,
		DeviceID:    deviceID,
		PrivKeyPath: privKeyPath,
		Region:      region,
	}

	client, err := newClient(device)
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
