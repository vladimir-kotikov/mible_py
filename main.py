import json
import logging
import time
from dataclasses import dataclass
from typing import Any, Dict, Tuple

from btlewrap.bluepy import BluepyBackend
from magiconf import load
from mitemp_bt.mitemp_bt_poller import MI_HUMIDITY, MI_TEMPERATURE, MiTempBtPoller
from paho.mqtt.client import Client


@dataclass
class Config:
    mible_address: str
    broker_address: str = "localhost"
    poll_interval: int = 30  # seconds
    state_topic: str = ""
    disco_topic: str = ""
    debug: bool = False


def make_payloads(state_topic: str) -> Tuple[Dict[str, Any], Dict[str, Any]]:
    temp_payload = {
        "name": "Mible temperature",
        "state_topic": state_topic,
        "device_class": "temperature",
        "unit_of_measurement": "°C",
        "value_template": "{{ value_json.temperature}}",
    }
    humidity_payload = {
        "name": "Mible humidity",
        "state_topic": state_topic,
        "device_class": "humidity",
        "unit_of_measurement": "%",
        "value_template": "{{ value_json.humidity}}",
    }

    return temp_payload, humidity_payload


cfg = load(Config)
logger = logging.getLogger("mible")
logging.basicConfig(
    level=logging.DEBUG if cfg.debug else logging.INFO,
    handlers=[logging.StreamHandler()],
    format=logging.BASIC_FORMAT,
)

if not cfg.state_topic:
    # Sanitize address - semicolons are not allowed in topic names
    mible_address = cfg.mible_address.replace(":", "_").lower()
    cfg.state_topic = f"mible/{mible_address}/state"

if not cfg.disco_topic:
    cfg.disco_topic = "homeassistant/sensor/mible/config"


logger.info("Starting mible...")

client = Client()
client.connect(cfg.broker_address)
logger.debug(f"Connected to broker at {cfg.broker_address}")

for disco in make_payloads(cfg.state_topic):
    client.publish(cfg.disco_topic, json.dumps(disco), qos=1)

poller = MiTempBtPoller(cfg.mible_address, BluepyBackend, cache_timeout=cfg.poll_interval)
logger.debug(f"Connected to Mible at {cfg.mible_address}")
while True:
    temp = poller.parameter_value(MI_TEMPERATURE)
    humidity = poller.parameter_value(MI_HUMIDITY)
    payload = json.dumps({"temperature": temp, "humidity": humidity})
    client.publish(cfg.state_topic, payload, retain=True)

    logger.debug(f"temp = {temp} humidity = {humidity}")
    time.sleep(cfg.poll_interval)
