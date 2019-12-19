package main

import (
	"github.com/crgimenes/goconfig"
)

// Config holds handler configuration
type Config struct {
	BrokerAddress  string `cfgRequired:"true"`
	DeviceAddress  string `cfgRequired:"true"`
	UpdateInterval uint   `cfgDefault:"30"`
	SentryDSN      string
}

// LoadConfig loads config from file
func LoadConfig() (Config, error) {
	conf := Config{}
	err := goconfig.Parse(&conf)
	return conf, err
}
