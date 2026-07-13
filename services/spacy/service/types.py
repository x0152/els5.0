from dataclasses import dataclass, field
from typing import Any

from service.enums import POS, Language, UnitType


@dataclass
class BaseForm:
    text: str
    pos: POS | None = None
    is_stop: bool = False
    language: Language = Language.EN
    meaning: str | None = None


@dataclass
class UnitSpan:
    position: int
    span_type: str
    start: int
    end: int
    text: str


@dataclass
class Unit:
    unit_type: UnitType
    base_form: str
    sentence_idx: int
    metadata: dict[str, Any] = field(default_factory=dict)
    language: Language | str = Language.EN
    spans: list[UnitSpan] = field(default_factory=list)
    pos: POS | None = None


@dataclass
class ProcessorContext:
    language: str = "en"
    sentence_count: int = 0
    units: list[Unit] = field(default_factory=list)
    base_forms: dict[str, BaseForm] = field(default_factory=dict)
    used_spans: set[tuple[int, int]] = field(default_factory=set)
