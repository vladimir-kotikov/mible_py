package main

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	online  = []byte("online")
	offline = []byte("offline")
)

// DiscoPayload is a struct to advertise sensor to home assistant over MQTT
type DiscoPayload struct {
	AvailabilityTopic string `json:"availability_topic"`
	DeviceClass       string `json:"device_class"`
	Name              string `json:"name"`
	StateTopic        string `json:"state_topic"`
	Unit              string `json:"unit_of_measurement"`
}

// Sensor represents a single sensor in mible device
type Sensor struct {
	broker    *Broker
	name      string
	char      string
	unit      string
	baseTopic string
}

// NewSensor creates a new sensor
func NewSensor(broker *Broker, name, char, unit, baseTopic string) *Sensor {
	return &Sensor{
		broker:    broker,
		name:      name,
		char:      char,
		unit:      unit,
		baseTopic: baseTopic,
	}
}

// availTopic ...
func (s Sensor) availTopic() string {
	return s.baseTopic + "/avail"
}

// discoTopic ...
func (s Sensor) discoTopic() string {
	return s.baseTopic + "/config"
}

// stateTopic ...
func (s Sensor) stateTopic() string {
	return s.baseTopic + "/state"
}

// Characteristic returns sensor's characteristic name
func (s Sensor) Characteristic() string {
	return s.char
}

// PublishDisco sends sensor's discoveery information oveer MQTT
func (s Sensor) PublishDisco() error {
	payload := &DiscoPayload{
		Name:              s.name,
		DeviceClass:       s.char,
		Unit:              s.unit,
		AvailabilityTopic: s.availTopic(),
		StateTopic:        s.stateTopic(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "can't serialize disco payload")
	}

	log.Debugf(">> %s: %s", s.discoTopic(), string(data))
	poken, err := s.broker.PublishWithQOS(s.discoTopic(), 1, data)
	if err != nil {
		return errors.Wrap(err, "can't send discovery info to Home Assistant")
	}

	if poken.WaitTimeout(1 * time.Second); poken.Error() != nil {
		return errors.Wrap(poken.Error(), "can't send discovery info to Home Assistant")
	}

	return nil
}

// PublishState sends sensor's state over MQTT to device's state topic
func (s *Sensor) PublishState(state []byte) {
	log.Debugf(">> %s: %s", s.availTopic(), string(online))
	s.broker.PublishAndRetain(s.availTopic(), online)

	log.Debugf(">> %s: %s", s.stateTopic(), string(state))
	s.broker.PublishAndRetain(s.stateTopic(), state)
}
