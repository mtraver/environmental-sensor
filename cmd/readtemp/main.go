package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/mcp9808"
	"periph.io/x/periph/host"

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

func makeProto(temp physic.Temperature) (*measurementpb.Measurement, error) {
	timepb, err := ptypes.TimestampProto(time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return &measurementpb.Measurement{
		DeviceId:  "none",
		Timestamp: timepb,
		Temp:      float32(temp.Celsius()),
	}, nil
}

func toJSONProto(temp physic.Temperature) (string, error) {
	m, err := makeProto(temp)
	if err != nil {
		return "", err
	}

	marshaler := jsonpb.Marshaler{
		Indent: "  ",
	}
	return marshaler.MarshalToString(m)
}

func toBinaryProto(temp physic.Temperature) ([]byte, error) {
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

	// Initialize periph.
	if _, err := host.Init(); err != nil {
		fatalf("Failed to initialize periph: %v", err)
	}

	// Open default I²C bus.
	bus, err := i2creg.Open("")
	if err != nil {
		fatalf("Failed to open I²C bus: %v", err)
	}
	defer bus.Close()

	sensor, err := mcp9808.New(bus, &mcp9808.DefaultOpts)
	if err != nil {
		fatalf("Failed to initialize MCP9808: %v", err)
	}

	temp, err := sensor.SenseTemp()
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
