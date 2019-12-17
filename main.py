import json
import logging
import time
from dataclasses import dataclass

from btlewrap.bluepy import BluepyBackend
import magiconf
from mitemp_bt.mitemp_bt_poller import MI_HUMIDITY, MI_TEMPERATURE, MiTempBtPoller
from paho.mqtt.client import Client
import sentry_sdk

logger = logging.getLogger("mible")

KNOWN_CHARS = {
    MI_TEMPERATURE: "°C",
    MI_HUMIDITY: "%",
}


@dataclass
class Config:
    mible_address: str
    broker_address: str = "localhost"
    poll_interval: int = 30  # seconds
    sentry_endpoint: str = ""
    debug: bool = False

    @classmethod
    def load(cls):
        return magiconf.load(cls)


class Sensor:
    def __init__(self, char: str, unit: str, poller):
        self.char = char
        self.unit = unit
        self._poller = poller

    @property
    def disco_payload(self):
        return {
            "name": f"Mible {self.char}",
            "device_class": self.char,
            "unit_of_measurement": self.unit,
            "value_template": "{{ value_json.%s}}" % self.char,
        }

    def read(self) -> str:
        return self._poller.parameter_value(self.char)


class MibleDevice:
    def __init__(self, mible_addr: str, cache_timeout: int):
        # Sanitize address - semicolons are not allowed in topic names
        self._device_addr = mible_addr.replace(":", "_").lower()
        self._poller = MiTempBtPoller(
            mible_addr, BluepyBackend, cache_timeout=cache_timeout
        )
        self.sensors = [
            Sensor(char, unit, self._poller) for char, unit in KNOWN_CHARS.items()
        ]

    def connect(self):
        logger.debug(f"Connected to Mible at {self._device_addr}")

    @property
    def safe_address(self):
        return self._device_addr

    @property
    def state_topic(self):
        return f"mible/{self._device_addr}/state"


cfg = Config.load()

if cfg.sentry_endpoint:
    sentry_sdk.init(cfg.sentry_endpoint)

logging.basicConfig(
    level=logging.DEBUG if cfg.debug else logging.INFO,
    handlers=[logging.StreamHandler()],
    format=logging.BASIC_FORMAT,
)
logger.info(f"Starting mible (polling every {cfg.poll_interval} sec)...")

client = Client()
client.connect(cfg.broker_address)
logger.debug(f"Connected to broker at {cfg.broker_address}")

device = MibleDevice(cfg.mible_address, cfg.poll_interval)
device.connect()
for sensor in device.sensors:
    disco_topic = (
        f"homeassistant/sensor/mible/{device.safe_address}/{sensor.char}/config"
    )
    payload = {"state_topic": device.state_topic, **sensor.disco_payload}
    client.publish(disco_topic, json.dumps(payload), qos=1)

while True:
    payload = {sensor.char: sensor.read() for sensor in device.sensors}
    client.publish(device.state_topic, json.dumps(payload), retain=True)
    logger.debug(json.dumps(payload))
    time.sleep(cfg.poll_interval)
