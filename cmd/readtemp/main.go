package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/mtraver/mcp9808"

	measurementpb "github.com/mtraver/environmental-sensor/measurement"
)

var (
	jsonOut   bool
	binaryOut bool
)

func fatalf(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
	os.Exit(1)
}

func init() {
	flag.BoolVar(&jsonOut, "json", false, "output a JSON-encoded proto instead of plain temp value")
	flag.BoolVar(&binaryOut, "binary", false, "output a binary-encoded proto instead of plain temp value")
}

func parseFlags() error {
	flag.Parse()

	if jsonOut && binaryOut {
		return fmt.Errorf("only one of -json and -binary may be given")
	}

	return nil
}

func makeProto(temp float32) (*measurementpb.Measurement, error) {
	timepb, err := ptypes.TimestampProto(time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return &measurementpb.Measurement{
		DeviceId:  "none",
		Timestamp: timepb,
		Temp:      temp,
	}, nil
}

func toJSONProto(temp float32) (string, error) {
	m, err := makeProto(temp)
	if err != nil {
		return "", err
	}

	marshaler := jsonpb.Marshaler{
		Indent: "  ",
	}
	return marshaler.MarshalToString(m)
}

func toBinaryProto(temp float32) ([]byte, error) {
	m, err := makeProto(temp)
	if err != nil {
		return nil, err
	}

	return proto.Marshal(m)
}

func main() {
	if err := parseFlags(); err != nil {
		fmt.Printf("argument error: %v\n", err)
		flag.Usage()
		os.Exit(2)
	}

	sensor, err := mcp9808.NewMCP9808()
	if err != nil {
		fatalf("Error connecting to sensor: %v", err)
	}
	defer sensor.Close()

	if err = sensor.Check(); err != nil {
		fatalf("Sensor check failed: %v", err)
	}

	temp, err := sensor.ReadTemp()
	if err != nil {
		fatalf("Failed to read temp: %v", err)
	}

	if jsonOut {
		json, err := toJSONProto(temp)
		if err != nil {
			fatalf("Failed to make JSON proto: %v", err)
		}

		fmt.Println(json)
	} else if binaryOut {
		pbBytes, err := toBinaryProto(temp)
		if err != nil {
			fatalf("Failed to make binary proto: %v", err)
		}

		fmt.Print(pbBytes)
	} else {
		fmt.Println(temp)
	}
}
