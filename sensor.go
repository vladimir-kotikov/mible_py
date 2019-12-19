package main

import (
	"fmt"
)

// DiscoPayload is a struct to advertise sensor to home assistant over MQTT
type DiscoPayload struct {
	Name          string `json:"name"`
	DeviceClass   string `json:"device_class"`
	StateTopic    string `json:"state_topic"`
	Unit          string `json:"unit_of_measurement"`
	ValueTemplate string `json:"value_template"`
}

// StatePayload represents a payload sent to HomeAssistant over MQTT
type StatePayload struct {
	Temperature float32 `json:"temperature"`
	Humidity    float32 `json:"humidity"`
}

// Sensor represents a single sensor in mible device
type Sensor struct {
	Name string
	Char string
	Unit string
}

// NewSensor creates a new sensor
func NewSensor(name, char, unit string) Sensor {
	return Sensor{name, char, unit}
}

// DiscoPayload returns MQTT disco payload to send to home assistant
func (s Sensor) DiscoPayload() *DiscoPayload {
	return &DiscoPayload{
		DeviceClass:   s.Char,
		Name:          s.Name,
		Unit:          s.Unit,
		ValueTemplate: fmt.Sprintf("{{ value_json.%s}}", s.Char),
	}
}
