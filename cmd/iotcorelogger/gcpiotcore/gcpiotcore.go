package gcpiotcore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mtraver/iotcore"
)

const (
	certExtension = ".x509"
)

func certPath(keyPath string) string {
	ext := path.Ext(keyPath)
	return keyPath[:len(keyPath)-len(ext)] + certExtension
}

func ParseDeviceFile(filepath string) (iotcore.Device, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return iotcore.Device{}, err
	}

	var device iotcore.Device
	if err := json.Unmarshal(b, &device); err != nil {
		return iotcore.Device{}, err
	}

	if device.DeviceID == "" {
		deviceID, err := iotcore.DeviceIDFromCert(certPath(device.PrivKeyPath))
		if err != nil {
			return iotcore.Device{}, err
		}
		device.DeviceID = deviceID
	}

	return device, nil
}

func fileStore(dir string) func(*iotcore.Device, *mqtt.ClientOptions) error {
	return func(device *iotcore.Device, opts *mqtt.ClientOptions) error {
		opts.SetStore(mqtt.NewFileStore(dir))
		return nil
	}
}

func commandHandler(client mqtt.Client, msg mqtt.Message) {
	msg.Ack()
	log.Printf("[GCP] Received command message with ID %v", msg.MessageID())
}

func configHandler(client mqtt.Client, msg mqtt.Message) {
	msg.Ack()
	log.Printf("[GCP] Received config message with ID %v", msg.MessageID())
}

func onConnect(device *iotcore.Device, opts *mqtt.ClientOptions) error {
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Printf("[GCP] Connected to MQTT broker")

		waitDur := 10 * time.Second

		// Subscribe to the command topic.
		topic := device.CommandTopic()
		if token := client.Subscribe(topic, 1, commandHandler); !token.WaitTimeout(waitDur) {
			log.Printf("[GCP] Subscription attempt to command topic %s timed out after %v", topic, waitDur)
		} else if token.Error() != nil {
			log.Printf("[GCP] Failed to subscribe to command topic %s: %v", topic, token.Error())
		} else {
			log.Printf("[GCP] Subscribed to command topic %s", topic)
		}

		// Subscribe to the config topic.
		topic = device.ConfigTopic()
		if token := client.Subscribe(topic, 1, configHandler); !token.WaitTimeout(waitDur) {
			log.Printf("[GCP] Subscription attempt to config topic %s timed out after %v", topic, waitDur)
		} else if token.Error() != nil {
			log.Printf("[GCP] Failed to subscribe to config topic %s: %v", topic, token.Error())
		} else {
			log.Printf("[GCP] Subscribed to config topic %s", topic)
		}
	})
	return nil
}

func onConnectionLost(device *iotcore.Device, opts *mqtt.ClientOptions) error {
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("[GCP] Connection to MQTT broker lost: %v", err)
	})
	return nil
}

func MQTTConnect(device iotcore.Device, caCertsPath, jwtPath, mqttStoreDir string) (mqtt.Client, error) {
	certsFile, err := os.Open(caCertsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open certs file: %v", err)
	}
	defer certsFile.Close()

	client, err := device.NewClient(iotcore.DefaultBroker, certsFile,
		iotcore.PersistentlyCacheJWT(60*time.Minute, jwtPath), fileStore(mqttStoreDir), onConnect, onConnectionLost)
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
