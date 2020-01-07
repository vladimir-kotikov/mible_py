package main

import (
	"github.com/crgimenes/goconfig"
)

// Config holds handler configuration
type Config struct {
	BrokerAddress  string `cfg:"broker_address" cfgRequired:"true"`
	Debug          bool   `cfg:"debug" cfgDefault:"false"`
	DeviceAddress  string `cfg:"device_address" cfgRequired:"true"`
	DeviceName     string `cfg:"device_name" cfgDefault:"Mible"`
	SentryDSN      string `cfg:"sentry_dsn"`
	UpdateInterval int    `cfg:"update_interval" cfgDefault:"30"`
}

// LoadConfig loads config from file
func LoadConfig() (Config, error) {
	conf := Config{}
	err := goconfig.Parse(&conf)
	return conf, err
}
