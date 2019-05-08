// Program iotcorelogger reads the temperature from an MCP9808 sensor and publishes
// it to Google Cloud IoT Core over MQTT.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/mcp9808"
	"periph.io/x/periph/host"

	"github.com/mtraver/environmental-sensor/cmd/iotcorelogger/pending"
	"github.com/mtraver/environmental-sensor/iotcore"
	measurementpb "github.com/mtraver/environmental-sensor/measurement"
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

	bridge iotcore.MQTTBridge

	caCerts string

	numSamples     int
	sampleInterval int

	// The allowed ports for connecting to the IoT Core MQTT bridge
	allowedPorts = map[int]bool{
		8883: true,
		443:  true,
	}

	// The directory in which to store measurements that failed to publish.
	pendingDir string
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

	flag.StringVar(&bridge.Host, "mqtthost", defaultHost, "MQTT host")
	flag.IntVar(&bridge.Port, "mqttport", 8883, "MQTT port")

	flag.IntVar(&numSamples, "numsamples", 3, "number of samples to take")
	flag.IntVar(&sampleInterval, "interval", 1, "number of seconds to wait between samples")

	flag.StringVar(
		&pendingDir, "pendingdir", "",
		"Directory in which to store measurements that failed to publish, e.g. because the network went down. Publication will be retried later.")
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

	if bridge.Host == "" {
		return fmt.Errorf("mqtthost flag must be given")
	}

	if _, ok := allowedPorts[bridge.Port]; !ok {
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

func newClient(conf iotcore.DeviceConfig) (mqtt.Client, error) {
	mqttOptions, err := iotcore.NewMQTTOptions(conf, bridge, caCerts)
	if err != nil {
		return nil, err
	}

	jwt, err := conf.NewJWT(time.Minute)
	if err != nil {
		return nil, err
	}
	mqttOptions.SetPassword(jwt)

	return mqtt.NewClient(mqttOptions), nil
}

func save(m *measurementpb.Measurement) {
	if pendingDir != "" {
		if err := pending.Save(m, pendingDir); err != nil {
			log.Printf("Failed to save measurement: %v", err)
		}
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

	deviceConf := iotcore.DeviceConfig{
		ProjectID:   projectID,
		RegistryID:  registryID,
		DeviceID:    deviceID,
		PrivKeyPath: privKeyPath,
		Region:      region,
	}

	client, err := newClient(deviceConf)
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
	m := &measurementpb.Measurement{
		DeviceId:  deviceConf.DeviceID,
		Timestamp: timepb,
		Temp:      mean(temps),
	}
	pbBytes, err := proto.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}

	// Attempt to connect using the MQTT client. If it fails, save the Measurement for later publication.
	waitDur := 5 * time.Second
	if token := client.Connect(); token.WaitTimeout(waitDur) && token.Error() != nil {
		save(m)
		log.Fatalf("Failed to connect MQTT client: %v", token.Error())
	}

	// We don't currently do anything with configs from the server.
	// client.Subscribe(deviceConf.ConfigTopic(), 1, configHandler)

	token := client.Publish(deviceConf.TelemetryTopic(), 1, false, pbBytes)
	if ok := token.WaitTimeout(waitDur); !ok {
		log.Printf("Failed to publish after %v", waitDur)

		// Save the Measurement to retry later.
		save(m)
	} else if pendingDir != "" {
		// Publish succeeded, so attempt to publish any pending measurements.
		if err := pending.PublishAll(client, deviceConf.TelemetryTopic(), pendingDir); err != nil {
			log.Printf("Failed to publish all pending measurements: %v", err)
		}
	}

	client.Disconnect(250)

	time.Sleep(500 * time.Millisecond)
}
