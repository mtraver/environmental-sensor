// Program readtemp reads the temperature from an MCP9808 sensor and prints it to stdout.
// By default it simply prints the value (in degrees Celsius), but it may optionally create
// a Measurement proto and print it in JSON or binary form.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/mcp9808"
	"periph.io/x/host/v3"
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

func makeProto(temp physic.Temperature) (*mpb.Measurement, error) {
	timepb := tspb.New(time.Now().UTC())
	if err := timepb.CheckValid(); err != nil {
		return nil, err
	}

	return &mpb.Measurement{
		DeviceId:  "none",
		Timestamp: timepb,
		Temp:      wpb.Float(float32(temp.Celsius())),
	}, nil
}

func toJSONProto(temp physic.Temperature) (string, error) {
	m, err := makeProto(temp)
	if err != nil {
		return "", err
	}

	marshaler := protojson.MarshalOptions{
		Indent: "  ",
	}
	b, err := marshaler.Marshal(m)
	return string(b), err
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
