package main

import (
	"github.com/crgimenes/goconfig"
)

// Config holds handler configuration
type Config struct {
	BrokerAddress  string `cfg:"broker_address" cfgRequired:"true"`
	DeviceAddress  string `cfg:"device_address" cfgRequired:"true"`
	UpdateInterval uint   `cfg:"update_interval" cfgDefault:"30"`
	SentryDSN      string `cfg:"sentry_dsn"`
}

// LoadConfig loads config from file
func LoadConfig() (Config, error) {
	conf := Config{}
	err := goconfig.Parse(&conf)
	return conf, err
}
