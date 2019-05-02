package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/mtraver/environmental-sensor/cmd/iotcorelogger/pending"
	"github.com/mtraver/environmental-sensor/iotcore"
	measurementpb "github.com/mtraver/environmental-sensor/measurement"
)

const (
	defaultRegion = "us-central1"
	defaultHost   = "mqtt.googleapis.com"
)

var (
	deviceConf iotcore.DeviceConfig
	bridge     iotcore.MQTTBridge

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
	flag.StringVar(&deviceConf.ProjectID, "project", "", "Google Cloud Platform project ID")
	flag.StringVar(&deviceConf.RegistryID, "registry", "", "Google Cloud IoT Core registry ID")
	flag.StringVar(&deviceConf.Region, "region", defaultRegion, "Google Cloud Platform region")
	flag.StringVar(&deviceConf.PrivKeyPath, "key", "", "path to device's private key")
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

	if deviceConf.ProjectID == "" {
		return fmt.Errorf("project flag must be given")
	}

	if deviceConf.RegistryID == "" {
		return fmt.Errorf("registry flag must be given")
	}

	if deviceConf.Region == "" {
		return fmt.Errorf("region flag must be given")
	}

	if deviceConf.PrivKeyPath == "" {
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

func mean(s []float32) float32 {
	var sum float32
	for _, e := range s {
		sum += e
	}

	return sum / float32(len(s))
}

func newClient() (mqtt.Client, error) {
	keyBytes, err := ioutil.ReadFile(deviceConf.PrivKeyPath)
	if err != nil {
		return nil, err
	}

	mqttOptions, err := iotcore.NewMQTTOptions(deviceConf, bridge, caCerts)
	if err != nil {
		return nil, err
	}

	jwt, err := deviceConf.NewJWT(keyBytes, time.Minute)
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

	var err error
	deviceConf.DeviceID, err = iotcore.DeviceIDFromCert(deviceConf.CertPath())
	if err != nil {
		log.Fatal(err)
	}

	client, err := newClient()
	if err != nil {
		log.Fatalf("Failed to make MQTT client: %v", err)
	}

	// Read the temp, construct a protobuf, and marshal it to bytes.
	temps, err := readTempMulti(numSamples, time.Duration(sampleInterval)*time.Second)
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
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		save(m)
		log.Fatalf("Failed to connect MQTT client: %v", token.Error())
	}

	// We don't currently do anything with configs from the server.
	// client.Subscribe(deviceConf.ConfigTopic(), 1, configHandler)

	token := client.Publish(deviceConf.TelemetryTopic(), 1, false, pbBytes)
	waitDur := 5 * time.Second
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
