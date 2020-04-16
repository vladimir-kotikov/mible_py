package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	log "github.com/sirupsen/logrus"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/mqtt"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal("couldn't load config", err)
	}

	log.Info("debug: ", cfg.Debug)
	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	if cfg.SentryDSN != "" {
		sentry.Init(sentry.ClientOptions{Dsn: cfg.SentryDSN})
	}

	adaptor := NewBluetoothAdapter()
	driver := NewMibleDriver(adaptor, cfg.DeviceName, cfg.DeviceAddress)
	broker := mqtt.NewAdaptor(cfg.BrokerAddress, "Mible")

	unitByChar := map[string]string{
		"temperature": "°C",
		"humidity":    "%",
	}

	sensors := []*Sensor{}

	for char, unit := range unitByChar {
		safeDeviceName := strings.ToLower(strings.Replace(cfg.DeviceName, " ", "_", -1))
		baseTopic := fmt.Sprintf("homeassistant/sensor/mible/%s_%s_%s", safeDeviceName, driver.ID(), char)
		sensorName := cfg.DeviceName + " " + char
		sensors = append(sensors, NewSensor(broker, sensorName, char, unit, baseTopic))
	}

	work := func() {
		for _, sensor := range sensors {
			err := sensor.PublishDisco()
			if err != nil {
				sentry.CaptureException(err)
				log.Fatal(err)
			}

			val := driver.GetCharacteristic(sensor.Characteristic())
			if val != nil {
				state := []byte(fmt.Sprintf("%.2f", *val))
				sensor.PublishState(state)
			}
		}

		gobot.Every(time.Duration(cfg.UpdateInterval)*time.Second, func() {
			for _, sensor := range sensors {
				val := driver.GetCharacteristic(sensor.Characteristic())
				if val != nil {
					state := []byte(fmt.Sprintf("%.2f", *val))
					sensor.PublishState(state)
				}
			}
		})
	}

	robot := gobot.NewRobot("Mible",
		[]gobot.Connection{adaptor, broker},
		[]gobot.Device{driver},
		work,
	)

	robot.Start()
}
