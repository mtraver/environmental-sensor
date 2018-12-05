package main

import (
	"fmt"
	"os"

	"github.com/mtraver/mcp9808"
)

func fatal(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
	os.Exit(1)
}

func main() {
	sensor, err := mcp9808.NewMCP9808()
	if err != nil {
		fatal("Error connecting to sensor: %v", err)
	}
	defer sensor.Close()

	if err = sensor.Check(); err != nil {
		fatal("Sensor check failed: %v", err)
	}

	temp, err := sensor.ReadTemp()
	if err != nil {
		fatal("Failed to read temp: %v", err)
	}

	fmt.Println(temp)
}
