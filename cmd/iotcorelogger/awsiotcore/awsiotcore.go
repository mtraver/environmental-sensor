package awsiotcore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	aic "github.com/mtraver/awsiotcore"
)

func ParseDeviceFile(filepath string) (aic.Device, error) {
	b, err := ioutil.ReadFile(filepath)
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

func fileStore(dir string) func(*aic.Device, *mqtt.ClientOptions) error {
	return func(device *aic.Device, opts *mqtt.ClientOptions) error {
		opts.SetStore(mqtt.NewFileStore(dir))
		return nil
	}
}

func commandHandler(client mqtt.Client, msg mqtt.Message) {
	msg.Ack()
	log.Printf("[AWS] Received command message with ID %v", msg.MessageID())
}

func onConnect(device *aic.Device, opts *mqtt.ClientOptions) error {
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Printf("[AWS] Connected to MQTT broker")

		// TODO subscribe to cloud-to-device topics
	})
	return nil
}

func onConnectionLost(device *aic.Device, opts *mqtt.ClientOptions) error {
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("[AWS] Connection to MQTT broker lost: %v", err)
	})
	return nil
}

func MQTTConnect(device aic.Device, mqttStoreDir string) (mqtt.Client, error) {
	// Make sure the MQTT store dir exists.
	if err := os.MkdirAll(mqttStoreDir, 0700); err != nil {
		return nil, err
	}

	client, err := device.NewClient(fileStore(mqttStoreDir), onConnect, onConnectionLost)
	if err != nil {
		return nil, fmt.Errorf("failed to make MQTT client: %v", err)
	}

	// Connect to the MQTT server.
	waitDur := 10 * time.Second
	if token := client.Connect(); !token.WaitTimeout(waitDur) {
		return nil, fmt.Errorf("MQTT connection attempt timed out after %v", waitDur)
	} else if token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %v", token.Error())
	}

	return client, nil
}
