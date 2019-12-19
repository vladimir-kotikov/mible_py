package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/mqtt"
)

var sensors = []Sensor{
	NewSensor("Mible temperature", "temperature", "C"),
	NewSensor("Mible humidity", "humidity", "%"),
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal("couldn't load config", err)
	}

	if cfg.SentryDSN != "" {
		sentry.Init(sentry.ClientOptions{Dsn: cfg.SentryDSN})
	}

	adaptor := NewBluetoothAdapter()
	driver := NewMibleDriver(adaptor, cfg.DeviceAddress)

	broker := mqtt.NewAdaptor(cfg.BrokerAddress, "Mible")
	queue := mqtt.NewDriver(broker, "")

	work := func() {
		safeAddress := strings.ToLower(strings.Replace(cfg.DeviceAddress, ":", "_", -1))
		stateTopic := fmt.Sprintf("mible/%s/state", safeAddress)

		for _, sensor := range sensors {
			discoTopic := fmt.Sprintf("homeassistant/sensor/mible/%s/%s/config", safeAddress, sensor.Char)
			queue.SetTopic(discoTopic)

			payload := sensor.DiscoPayload()
			payload.StateTopic = stateTopic
			data, err := json.Marshal(payload)
			if err != nil {
				sentry.CaptureException(err)
				log.Print("can't serialize disco payload")
				continue
			}

			queue.Publish(data)
		}

		queue.SetTopic(stateTopic)
		gobot.Every(time.Duration(cfg.UpdateInterval)*time.Second, func() {
			p := StatePayload{
				Temperature: driver.Temperature(),
				Humidity:    driver.Humidity(),
			}

			data, err := json.Marshal(p)
			if err != nil {
				sentry.CaptureException(err)
				log.Print("can't convert sensor readings to JSON")
				return
			}

			log.Print("data:", p)
			queue.Publish(data)
		})
	}

	robot := gobot.NewRobot("Mible",
		[]gobot.Connection{adaptor, broker},
		[]gobot.Device{driver, queue},
		work,
	)

	robot.Start()
}
