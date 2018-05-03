// Package mcp9808 provides an interface to the MCP9808 temperature sensor.
// Datasheet: http://ww1.microchip.com/downloads/en/DeviceDoc/25095A.pdf
package mcp9808

import (
	"encoding/binary"
	"fmt"

	"golang.org/x/exp/io/i2c"
)

const (
	// Default I2C address for device
	defaultI2CAddr = 0x18

	// Register addresses. See page 16 of the datasheet.
	regConfig           = 0x01
	regUpperTempTrip    = 0x02
	regLowerTempTrip    = 0x03
	regCriticalTempTrip = 0x04
	regTemp             = 0x05
	regManufacturerID   = 0x06
	regDeviceID         = 0x07
	regResolution       = 0x08

	devPath = "/dev/i2c-1"

	// Expected values for manufacturer and device ID
	manufacturerID = 0x54
	deviceID       = 0x400
)

type MCP9808 struct {
	device *i2c.Device
}

func NewMCP9808() (*MCP9808, error) {
	d, err := i2c.Open(&i2c.Devfs{Dev: devPath}, defaultI2CAddr)
	if err != nil {
		return nil, fmt.Errorf("mcp9808: %v", err)
	}

	return &MCP9808{
		device: d,
	}, nil
}

func (m *MCP9808) Close() error {
	return m.device.Close()
}

func (m *MCP9808) Check() error {
	mID, err := m.ReadUint16(regManufacturerID)
	if err != nil {
		return fmt.Errorf("mcp9808: error reading manufacturer ID: %v", err)
	}

	dID, err := m.ReadUint16(regDeviceID)
	if err != nil {
		return fmt.Errorf("mcp9808: error reading device ID: %v", err)
	}

	if mID != manufacturerID || dID != deviceID {
		return fmt.Errorf(
			"mcp9808: incorrect manufacturer or device ID. Got 0x%x, 0x%x. Expected 0x%x, 0x%x.",
			mID, dID, manufacturerID, deviceID)
	}

	return nil
}

func (m *MCP9808) ReadUint16(reg byte) (uint16, error) {
	b := make([]byte, 2)
	if err := m.device.ReadReg(reg, b); err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint16(b), nil
}

func (m *MCP9808) ReadTemp() (float32, error) {
	regVal, err := m.ReadUint16(regTemp)
	if err != nil {
		return 0, fmt.Errorf("mcp9808: error reading temp: %v", err)
	}

	temp := float32(regVal&0x0fff) / 16.0
	if regVal&0x1000 == 1 {
		temp -= 256.0
	}

	return temp, nil
}
