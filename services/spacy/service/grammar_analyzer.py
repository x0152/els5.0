import os
import re
from typing import Any

from service.text_utils import strip_html

try:
    from grammardetector import GrammarDetector
except Exception:
    GrammarDetector = None

TENSE_RULES = {
    "present_simple": ("present_simple", "active", ""),
    "present_simple_passive": ("present_simple", "passive", ""),
    "past_simple": ("past_simple", "active", ""),
    "past_simple_passive": ("past_simple", "passive", ""),
    "future_simple_will": ("future_simple", "active", "will"),
    "future_simple_will_passive": ("future_simple", "passive", "will"),
    "future_simple_be_going_to": ("future_simple", "active", ""),
    "future_simple_be_going_to_passive": ("future_simple", "passive", ""),
    "present_continuous": ("present_continuous", "active", ""),
    "present_continuous_passive": ("present_continuous", "passive", ""),
    "past_continuous": ("past_continuous", "active", ""),
    "past_continuous_passive": ("past_continuous", "passive", ""),
    "future_continuous": ("future_continuous", "active", "will"),
    "future_continuous_passive": ("future_continuous", "passive", "will"),
    "present_perfect": ("present_perfect", "active", ""),
    "present_perfect_passive": ("present_perfect", "passive", ""),
    "past_perfect": ("past_perfect", "active", ""),
    "past_perfect_passive": ("past_perfect", "passive", ""),
    "future_perfect": ("future_perfect", "active", "will"),
    "future_perfect_passive": ("future_perfect", "passive", "will"),
    "present_perfect_continuous": ("present_perfect_continuous", "active", ""),
    "present_perfect_continuous_passive": ("present_perfect_continuous", "passive", ""),
    "past_perfect_continuous": ("past_perfect_continuous", "active", ""),
    "past_perfect_continuous_passive": ("past_perfect_continuous", "passive", ""),
    "future_perfect_continuous": ("future_perfect_continuous", "active", "will"),
    "future_perfect_continuous_passive": ("future_perfect_continuous", "passive", "will"),
}


class DeterministicGrammarAnalyzer:
    def __init__(self):
        self._detector = None
        if GrammarDetector is None:
            return
        model = os.getenv("GRAMMAR_DETECTOR_MODEL", "en_core_web_sm")
        try:
            self._detector = GrammarDetector(language_model=model)
        except Exception:
            self._detector = None

    def analyze(self, html_content: str) -> list[dict[str, Any]]:
        if self._detector is None:
            return []
        text = strip_html(html_content)
        if not text.strip():
            return []

        try:
            results = self._detector(text)
        except Exception:
            return []
        if not isinstance(results, dict):
            return []

        items: list[dict[str, Any]] = []
        seen: set[tuple[Any, ...]] = set()
        search_from = 0

        for feature in ("tense_aspects", "determiners", "voices"):
            matches = results.get(feature, [])
            if not isinstance(matches, list):
                continue
            for m in matches:
                rule = self._normalize_rule(self._field(m, "rulename"))
                fragment = str(self._field(m, "span")).strip()
                if not rule or not fragment:
                    continue

                fingerprint = self._fingerprint(feature, rule)
                if not fingerprint["category"] or not fingerprint["sub_type"]:
                    continue

                start, end = self._find_span(text, fragment, search_from)
                if start < 0:
                    continue
                search_from = end

                key = (
                    fingerprint["category"],
                    fingerprint["sub_type"],
                    fingerprint["voice"],
                    fingerprint["modality"],
                    start,
                    end,
                )
                if key in seen:
                    continue
                seen.add(key)

                items.append(
                    {
                        "fingerprint": fingerprint,
                        "title": self._title(feature, rule),
                        "example": fragment,
                        "explanation": self._explanation(feature, rule),
                        "start_index": start,
                        "end_index": end,
                    }
                )

        return items

    def _fingerprint(self, feature: str, rule: str) -> dict[str, str]:
        if feature == "tense_aspects":
            sub_type, voice, modality = self._tense_from_rule(rule)
            return {
                "category": "tense" if sub_type else "",
                "sub_type": sub_type,
                "voice": voice,
                "modality": modality,
            }
        if feature == "determiners":
            if rule == "definite":
                return {"category": "article", "sub_type": "definite", "voice": "", "modality": ""}
            if rule == "indefinite":
                return {"category": "article", "sub_type": "indefinite", "voice": "", "modality": ""}
            if rule in {"none", "zero"}:
                return {"category": "article", "sub_type": "zero", "voice": "", "modality": ""}
            return {"category": "", "sub_type": "", "voice": "", "modality": ""}
        if feature == "voices" and rule == "passive":
            return {"category": "passive", "sub_type": "simple", "voice": "passive", "modality": ""}
        return {"category": "", "sub_type": "", "voice": "", "modality": ""}

    def _tense_from_rule(self, rule: str) -> tuple[str, str, str]:
        if rule in TENSE_RULES:
            return TENSE_RULES[rule]

        voice = "passive" if "passive" in rule else "active"
        modality = self._modality(rule)
        by_order = (
            ("present_perfect_continuous", "present_perfect_continuous"),
            ("past_perfect_continuous", "past_perfect_continuous"),
            ("future_perfect_continuous", "future_perfect_continuous"),
            ("present_perfect", "present_perfect"),
            ("past_perfect", "past_perfect"),
            ("future_perfect", "future_perfect"),
            ("present_continuous", "present_continuous"),
            ("past_continuous", "past_continuous"),
            ("future_continuous", "future_continuous"),
            ("present_simple", "present_simple"),
            ("past_simple", "past_simple"),
            ("future_simple", "future_simple"),
        )
        for needle, sub_type in by_order:
            if needle in rule:
                return sub_type, voice, modality

        compact = rule.replace("_", " ")
        if "future" in compact and "perfect" in compact and "continuous" in compact:
            return "future_perfect_continuous", voice, modality
        if "present" in compact and "perfect" in compact and "continuous" in compact:
            return "present_perfect_continuous", voice, modality
        if "past" in compact and "perfect" in compact and "continuous" in compact:
            return "past_perfect_continuous", voice, modality
        if "future" in compact and "perfect" in compact:
            return "future_perfect", voice, modality
        if "present" in compact and "perfect" in compact:
            return "present_perfect", voice, modality
        if "past" in compact and "perfect" in compact:
            return "past_perfect", voice, modality
        if "future" in compact and "continuous" in compact:
            return "future_continuous", voice, modality
        if "present" in compact and "continuous" in compact:
            return "present_continuous", voice, modality
        if "past" in compact and "continuous" in compact:
            return "past_continuous", voice, modality
        if "future" in compact:
            return "future_simple", voice, modality
        if "past" in compact:
            return "past_simple", voice, modality
        if "present" in compact:
            return "present_simple", voice, modality
        return "", "", ""

    def _field(self, item: Any, key: str) -> Any:
        if isinstance(item, dict):
            return item.get(key, "")
        return getattr(item, key, "")

    def _normalize_rule(self, s: Any) -> str:
        text = str(s).strip().lower()
        if not text:
            return ""
        text = text.replace("-", " ")
        text = re.sub(r"[^a-z0-9 ]+", " ", text)
        return "_".join(text.split())

    def _modality(self, rule: str) -> str:
        for name in ("have_to", "ought_to", "can", "could", "may", "might", "must", "shall", "should", "will", "would", "need"):
            if name in rule:
                return name
        return ""

    def _title(self, feature: str, rule: str) -> str:
        clean = " ".join(rule.split("_")).strip()
        if not clean:
            return feature.replace("_", " ").title()
        if feature == "voices":
            return f"{clean.title()} Voice"
        if feature == "determiners":
            return f"{clean.title()} Article"
        return clean.title()

    def _explanation(self, feature: str, rule: str) -> str:
        return f"deterministic:{feature}:{rule}"

    def _find_span(self, text: str, fragment: str, from_idx: int) -> tuple[int, int]:
        idx = text.find(fragment, max(0, from_idx))
        if idx < 0:
            idx = text.find(fragment)
        if idx < 0:
            return -1, -1
        return idx, idx + len(fragment)
