package main

import (
	"context"
	"strings"

	"encoding/binary"

	"github.com/go-ble/ble"
	"gobot.io/x/gobot"
)

var miUUID = ble.UUID16(0xfe95)

const (
	allowDuplicates = true
)

// MibleDriver ...
type MibleDriver struct {
	name    string
	address string
	adapter *BluetoothAdapter

	ctx    context.Context
	cancel context.CancelFunc

	temp float32
	humi float32
}

// NewMibleDriver ...
func NewMibleDriver(adapter *BluetoothAdapter, address string) *MibleDriver {
	return &MibleDriver{
		adapter: adapter,
		address: address,
		name:    gobot.DefaultName("MibleDriver"),
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

// Start initiates the Driver
func (driver *MibleDriver) Start() error {
	// Scan for specified durantion, or until interrupted by user.
	ctx, cancel := context.WithCancel(context.Background())
	driver.ctx = ctx
	driver.cancel = cancel

	advHandler := func(a ble.Advertisement) {
		temp, humi := parseServiceData(a.ServiceData())
		if temp != nil {
			driver.temp = *temp
		}
		if humi != nil {
			driver.humi = *humi
		}
	}

	advFilter := func(a ble.Advertisement) bool {
		return strings.ToLower(a.Addr().String()) == strings.ToLower(driver.address)
	}

	go func() {
		ble.Scan(ctx, allowDuplicates, advHandler, advFilter)
	}()

	// return ble.Scan(ctx, allowDuplicates, advHandler, advFilter)
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

// Temperature returns last read temperature value
func (driver *MibleDriver) Temperature() float32 {
	return driver.temp
}

// Humidity returns last read humidity value
func (driver *MibleDriver) Humidity() float32 {
	return driver.humi
}

func parseServiceData(serviceData []ble.ServiceData) (*float32, *float32) {
	if len(serviceData) == 0 {
		return nil, nil
	}

	var temp, humi float32
	for _, sData := range serviceData {
		if !sData.UUID.Equal(miUUID) {
			continue
		}

		rawData := sData.Data[11:]
		dataType := uint8(rawData[0])
		dataLen := uint8(rawData[2])
		sensorData := rawData[3 : 3+dataLen]

		switch dataType {
		case 0x0d:
			temp = float32(int16(binary.LittleEndian.Uint16(sensorData))) / 10
			humi = float32(binary.LittleEndian.Uint16(sensorData[2:])) / 10
		case 0x04:
			temp = float32(int16(binary.LittleEndian.Uint16(sensorData))) / 10
		case 0x06:
			humi = float32(binary.LittleEndian.Uint16(sensorData)) / 10
		}
	}

	return &temp, &humi
}
