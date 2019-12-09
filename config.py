from dataclasses import dataclass, fields, Field, MISSING
from typing import TypeVar, Type, Dict, Any


@dataclass
class Config:
    foo: str
    bar: str = "baz"

class InvalidConfigError(Exception):
    pass


T = TypeVar("T")


def load_config(conf_cls: Type[T], src: Dict[str, Any]) -> T:
    init_kw = {}
    for field in fields(conf_cls):
        f: Field = field

        try:
            val = src[f.name]
        except KeyError:
            if f.default == MISSING and f.default_factory == MISSING:
                raise InvalidConfigError(f"{f.name} is required but is missing")

        if not isinstance(val, f.type):
            raise InvalidConfigError(f"{f.name} is of wrong type ({type(val)}, expected {f.type})")

        init_kw[f.name] = src[f.name]

    return conf_cls(**init_kw)


if __name__ == "__main__":
    cfg = load_config(Config, {"foo": "foo", "bar": "bar"})
    assert cfg.foo == "foo"
    assert cfg.bar == "bar"

    cfg = load_config(Config, {"foo": "foo"})
    assert cfg.foo == "foo"
    assert cfg.bar == "baz"

    cfg = load_config(Config, {"foo": "foo", "quux": "blep"})
    assert cfg.foo == "foo"
    assert cfg.bar == "baz"

    cfg = load_config(Config, {"foo": 100})
    assert cfg.foo == "foo"
    assert cfg.bar == "baz"

    cfg = load_config(Config, {})
    assert cfg.foo == "foo"
    assert cfg.bar == "bar"