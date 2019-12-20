package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"

	"gobot.io/x/gobot"
)

var sensors = []*Sensor{
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

	broker := NewAdaptor(cfg.BrokerAddress, "Mible")

	work := func() {
		safeAddress := strings.ToLower(strings.Replace(cfg.DeviceAddress, ":", "_", -1))
		stateTopic := fmt.Sprintf("mible/%s/state", safeAddress)
		for _, sensor := range sensors {
			discoTopic := fmt.Sprintf("homeassistant/sensor/mible/%s/%s/config", safeAddress, sensor.Char)
			err := publishDisco(sensor, broker, stateTopic, discoTopic)
			if err != nil {
				sentry.CaptureException(err)
				log.Fatal(err)
			}
		}

		publishReadings(driver, broker, stateTopic)
		gobot.Every(time.Duration(cfg.UpdateInterval)*time.Second, func() {
			publishReadings(driver, broker, stateTopic)
		})
	}

	robot := gobot.NewRobot("Mible",
		[]gobot.Connection{adaptor, broker},
		[]gobot.Device{driver},
		work,
	)

	robot.Start()
}

func publishDisco(sensor *Sensor, broker *Adaptor, stateTopic, discoTopic string) error {
	payload := sensor.DiscoPayload()
	payload.StateTopic = stateTopic
	data, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "can't serialize disco payload")
	}

	poken, err := broker.PublishWithQOS(discoTopic, 1, data)
	if err != nil {
		return errors.Wrap(err, "can't send discovery info to Home Assistant")
	}

	if poken.WaitTimeout(1 * time.Second); poken.Error() != nil {
		return errors.Wrap(poken.Error(), "can't send discovery info to Home Assistant")
	}

	return nil
}

func publishReadings(driver *MibleDriver, broker *Adaptor, topic string) {
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
	broker.PublishAndRetain(topic, data)
}
