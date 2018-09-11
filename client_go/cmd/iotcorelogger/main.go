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

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/mtraver/environmental-sensor/client_go/iotcore"
	measurementpb "github.com/mtraver/environmental-sensor/client_go/measurement"
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
)

// We don't currently do anything with configs from the server
// func configHandler(client MQTT.Client, msg MQTT.Message) {
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

func main() {
	if err := parseFlags(); err != nil {
		fmt.Printf("argument error: %v\n", err)
		os.Exit(2)
	}

	keyBytes, err := ioutil.ReadFile(deviceConf.PrivKeyPath)
	if err != nil {
		log.Fatalf("Failed to read private key: %v", err)
	}

	deviceConf.DeviceID, err = iotcore.DeviceIDFromCert(deviceConf.CertPath())
	if err != nil {
		log.Fatal(err)
	}

	mqttOptions, err := iotcore.NewMQTTOptions(deviceConf, bridge, caCerts)
	if err != nil {
		log.Fatal(err)
	}

	tokenStr, err := deviceConf.NewJWT(keyBytes, time.Minute)
	if err != nil {
		log.Fatalf("Failed to sign JWT: %v", err)
	}
	mqttOptions.SetPassword(tokenStr)

	client := MQTT.NewClient(mqttOptions)
	if mqtt := client.Connect(); mqtt.Wait() && mqtt.Error() != nil {
		log.Fatalf("Failed to connect MQTT client: %v", mqtt.Error())
	}

	// We don't currently do anything with configs from the server
	// client.Subscribe(deviceConf.ConfigTopic(), 1, configHandler)

	// Read temp and construct a protobuf
	temps, err := readTempMulti(numSamples, time.Duration(sampleInterval)*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	timepb, err := ptypes.TimestampProto(time.Now().UTC())
	if err != nil {
		log.Fatal(err)
	}
	pbBytes, err := proto.Marshal(&measurementpb.Measurement{
		DeviceId:  deviceConf.DeviceID,
		Timestamp: timepb,
		Temp:      mean(temps),
	})
	if err != nil {
		log.Fatal(err)
	}

	pubToken := client.Publish(deviceConf.TelemetryTopic(), 1, false, pbBytes)
	pubToken.WaitTimeout(5 * time.Second)

	client.Disconnect(250)

	time.Sleep(500 * time.Millisecond)
}
