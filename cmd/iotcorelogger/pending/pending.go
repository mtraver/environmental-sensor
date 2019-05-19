// Package pending allows measurements to be persisted to disk and then published later.
// This is useful if the network connection goes down or if there are errors while
// connecting or publishing to the MQTT server.
package pending

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
)

const fileExt = ".json"

// Save converts the given Measurement to JSON and saves it to disk.
func Save(m *mpb.Measurement, dir string) error {
	marshaler := jsonpb.Marshaler{
		Indent: "  ",
	}
	json, err := marshaler.MarshalToString(m)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%x%s", sha256.Sum256([]byte(json)), fileExt)
	filepath := path.Join(dir, filename)

	return ioutil.WriteFile(filepath, []byte(json), 0644)
}

// PublishAll reads any Measurements saved to disk and attempts to publish
// them using the given MQTT client. It returns the first error encountered,
// or nil if all publishes succeed.
func PublishAll(client mqtt.Client, topic string, dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), fileExt) {
			filepath := path.Join(dir, file.Name())
			if err := publish(client, topic, filepath); err != nil {
				return err
			}

			os.Remove(filepath)
		}
	}

	return nil
}

func publish(client mqtt.Client, topic string, filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	m := mpb.Measurement{}
	if err := jsonpb.Unmarshal(f, &m); err != nil {
		return err
	}

	// Set the upload timestamp, since this is a delayed upload.
	timepb, err := ptypes.TimestampProto(time.Now().UTC())
	if err != nil {
		return err
	}
	m.UploadTimestamp = timepb

	pbBytes, err := proto.Marshal(&m)
	if err != nil {
		return err
	}

	token := client.Publish(topic, 1, false, pbBytes)
	waitDur := 10 * time.Second
	if ok := token.WaitTimeout(waitDur); !ok {
		return fmt.Errorf("pending: publish timed out after %v", waitDur)
	} else if token.Error() != nil {
		return fmt.Errorf("pending: publish failed: %v", token.Error())
	}

	return nil
}
