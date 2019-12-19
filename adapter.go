package main

import (
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux/hci/cmd"
	"gobot.io/x/gobot"
)

// BluetoothAdapter ...
type BluetoothAdapter struct {
	name string
}

// NewBluetoothAdapter ...
func NewBluetoothAdapter() *BluetoothAdapter {
	return &BluetoothAdapter{
		name: gobot.DefaultName("MibleAdapter"),
	}
}

// Name returns the label for the Adaptor
func (adaptor *BluetoothAdapter) Name() string {
	return adaptor.name
}

// SetName sets the label for the Adaptor
func (adaptor *BluetoothAdapter) SetName(n string) {
	adaptor.name = n
}

// Connect initiates the Adaptor
func (adaptor *BluetoothAdapter) Connect() error {
	// 0x00 is passive scan
	scanParams := cmd.LESetScanParameters{LEScanType: 0x00}
	d, err := DefaultDevice(ble.OptScanParams(scanParams))
	if err != nil {
		return err
	}

	ble.SetDefaultDevice(d)
	return nil
}

// Finalize terminates the Adaptor
func (adaptor *BluetoothAdapter) Finalize() error {
	return nil
}
