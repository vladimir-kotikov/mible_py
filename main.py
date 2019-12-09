import json
import time
from btlewrap.bluepy import BluepyBackend
from mitemp_bt.mitemp_bt_poller import (
    MI_HUMIDITY,
    MI_TEMPERATURE,
    MiTempBtPoller,
)
from paho.mqtt.client import Client

POLL_INTERVAL = 10  # seconds
BROKER_ADDRESS = "192.168.10.100"
MIBLE_ADDRESS = "4C:65:A8:DC:07:4D"
STATE_TOPIC = f"mible/${MIBLE_ADDRESS}/state"
DISCO_PREFIX = "homeassistant"
DISCO_TOPIC = f"${DISCO_PREFIX}/sensor/mible/config"
DISCO_TEMP_PAYLOAD = {
    "name": "Mible temperature",
    "state_topic": STATE_TOPIC,
    "device_class": "temperature",
    "unit_of_measurement": "°C",
    "value_template": "{{ value_json.temperature}}",
}
DISCO_HUMIDITY_PAYLOAD = {
    "name": "Mible humidity",
    "state_topic": STATE_TOPIC,
    "device_class": "humidity",
    "unit_of_measurement": "%",
    "value_template": "{{ value_json.humidity}}",
}

client = Client()
client.connect(BROKER_ADDRESS)
for payload in (DISCO_TEMP_PAYLOAD, DISCO_HUMIDITY_PAYLOAD):
    client.publish(DISCO_TOPIC, json.dumps(payload), qos=1)

poller = MiTempBtPoller(MIBLE_ADDRESS, BluepyBackend)
while True:
    temp = poller.parameter_value(MI_TEMPERATURE)
    humidity = poller.parameter_value(MI_HUMIDITY)
    payload = json.dumps({
        "temperature": temp,
        "humidity": humidity,
    })
    print(payload)
    client.publish(STATE_TOPIC, payload, retain=True)
    time.sleep(POLL_INTERVAL)
