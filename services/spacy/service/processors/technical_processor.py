import re

from service.enums import UnitType
from service.text_utils import strip_html
from service.types import BaseForm, ProcessorContext, Unit, UnitSpan


class TechnicalProcessor:
    PATTERNS = {
        "ssn": r"\b\d{3}-\d{2}-\d{4}\b",
        "zip_code": r"\b\d{5}(-\d{4})?\b",
        "twitter_handle": r"@\w+",
        "hashtag": r"#\w+",
        "version": r"v?\d+\.\d+\.\d+",
    }

    def process(self, html_content: str, context: ProcessorContext) -> list[Unit]:
        text = strip_html(html_content)
        units: list[Unit] = []

        for tech_type, pattern in self.PATTERNS.items():
            for match in re.finditer(pattern, text, re.IGNORECASE):
                base_form = match.group().lower()
                units.append(
                    Unit(
                        unit_type=UnitType.TECHNICAL,
                        base_form=base_form,
                        sentence_idx=0,
                        metadata={"tech_type": tech_type},
                        language=context.language,
                        spans=[
                            UnitSpan(
                                position=0,
                                span_type="word",
                                start=match.start(),
                                end=match.end(),
                                text=match.group(),
                            )
                        ],
                    )
                )
                if base_form not in context.base_forms:
                    context.base_forms[base_form] = BaseForm(text=base_form)

        return units
