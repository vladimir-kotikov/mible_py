package main

import (
	"context"
	"strings"

	"encoding/binary"

	log "github.com/sirupsen/logrus"

	"github.com/go-ble/ble"
	"gobot.io/x/gobot"
)

var miUUID = ble.UUID16(0xfe95)

const (
	allowDuplicates = true
	temperature     = 0x04
	humidity        = 0x06
	composite       = 0x0d
)

// MibleDriver ...
type MibleDriver struct {
	name    string
	address string
	adapter *BluetoothAdapter

	ctx    context.Context
	cancel context.CancelFunc

	characteristics map[string]float32
}

// NewMibleDriver ...
func NewMibleDriver(adapter *BluetoothAdapter, name, address string) *MibleDriver {
	return &MibleDriver{
		adapter: adapter,
		address: address,
		name:    name,

		characteristics: make(map[string]float32),
	}
}

// Name returns the label for the Driver
func (driver *MibleDriver) Name() string {
	return driver.name
}

// SetName sets the label for the Driver
func (driver *MibleDriver) SetName(s string) {
	driver.name = s
}

func (driver *MibleDriver) updateCharacteristics(updates map[string]float32) {
	for k, v := range updates {
		driver.characteristics[k] = v
	}
}

// Start initiates the Driver
func (driver *MibleDriver) Start() error {
	// Scan for specified durantion, or until interrupted by user.
	driver.ctx, driver.cancel = context.WithCancel(context.Background())

	advHandler := func(a ble.Advertisement) {
		chars := parseServiceData(a.ServiceData())
		log.Debugln("characteristics:", chars)
		driver.updateCharacteristics(chars)
	}

	advFilter := func(a ble.Advertisement) bool {
		return strings.ToLower(a.Addr().String()) == strings.ToLower(driver.address)
	}

	go func() {
		ble.Scan(driver.ctx, allowDuplicates, advHandler, advFilter)
	}()

	return nil
}

// Halt terminates the Driver
func (driver *MibleDriver) Halt() error {
	driver.cancel()
	<-driver.ctx.Done()
	err := driver.ctx.Err()

	driver.ctx = nil
	driver.cancel = nil

	if err == context.Canceled || err == context.DeadlineExceeded {
		return nil
	}

	return err
}

// Connection returns the Connection associated with the Driver
func (driver *MibleDriver) Connection() gobot.Connection {
	return driver.adapter
}

// ID returns device's uniquee name (based on MAC address)
func (driver *MibleDriver) ID() string {
	return strings.ToLower(strings.Replace(driver.address, ":", "_", -1))
}

// GetCharacteristic returns characteristic value by name or nil if no such characteristic found
func (driver *MibleDriver) GetCharacteristic(charName string) *float32 {
	val, ok := driver.characteristics[charName]
	if !ok {
		return nil
	}

	return &val
}

func parseServiceData(serviceData []ble.ServiceData) map[string]float32 {
	result := make(map[string]float32)
	if len(serviceData) == 0 {
		return result
	}

	for _, sData := range serviceData {
		if !sData.UUID.Equal(miUUID) {
			continue
		}

		rawData := sData.Data[11:]
		dataType := uint8(rawData[0])
		dataLen := uint8(rawData[2])
		sensorData := rawData[3 : 3+dataLen]

		switch dataType {
		case temperature:
			result["temperature"] = parseTemperature(sensorData)
		case humidity:
			result["humidity"] = parseHumidity(sensorData)
		case composite:
			result["temperature"] = parseTemperature(sensorData)
			result["humidity"] = parseHumidity(sensorData[2:])
		}
	}

	return result
}

func parseTemperature(sensorData []byte) float32 {
	return float32(int16(binary.LittleEndian.Uint16(sensorData))) / 10
}

func parseHumidity(sensorData []byte) float32 {
	return float32(binary.LittleEndian.Uint16(sensorData)) / 10
}
