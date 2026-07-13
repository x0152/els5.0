from service.enums import UnitType
from service.spacy_cache import SpacyDocumentCache
from service.text_utils import strip_html
from service.types import BaseForm, ProcessorContext, Unit, UnitSpan


class CompoundProcessor:
    def __init__(self, cache: SpacyDocumentCache):
        self._cache = cache

    def process(self, html_content: str, context: ProcessorContext) -> list[Unit]:
        text = strip_html(html_content)
        doc = self._cache.get(text)
        units: list[Unit] = []
        sentences = list(doc.sents)
        i = 0

        while i < len(doc):
            if not doc[i].is_alpha:
                i += 1
                continue

            parts = [doc[i]]
            j = i + 1

            while j < len(doc) - 1:
                if doc[j].text == "-" and doc[j + 1].is_alpha:
                    parts.append(doc[j])
                    parts.append(doc[j + 1])
                    j += 2
                else:
                    break

            if len(parts) >= 3:
                spans = []
                word_parts = []
                for p in parts:
                    span_type = "symbol" if p.text == "-" else "word"
                    spans.append(
                        UnitSpan(
                            position=len(spans),
                            span_type=span_type,
                            start=p.idx,
                            end=p.idx + len(p.text),
                            text=p.text,
                        )
                    )
                    if p.text != "-":
                        word_parts.append(p.lemma_)

                sent_idx = next((si for si, s in enumerate(sentences) if parts[0].i >= s.start), 0)
                base_form = "-".join(word_parts)

                units.append(
                    Unit(
                        unit_type=UnitType.COMPOUND,
                        base_form=base_form,
                        sentence_idx=sent_idx,
                        metadata={"parts": word_parts},
                        language=context.language,
                        spans=spans,
                    )
                )

                if base_form not in context.base_forms:
                    context.base_forms[base_form] = BaseForm(text=base_form)

                i = j
            else:
                i += 1

        return units
