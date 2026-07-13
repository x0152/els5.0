from service.enums import UnitType
from service.spacy_cache import SpacyDocumentCache
from service.text_utils import strip_html
from service.types import BaseForm, ProcessorContext, Unit, UnitSpan


class EntityProcessor:
    ENTITY_MAPPING = {
        "DATE": UnitType.TIME,
        "TIME": UnitType.TIME,
        "MONEY": UnitType.MONEY,
        "QUANTITY": UnitType.QUANTITY,
        "PERCENT": UnitType.PERCENT,
        "PRODUCT": UnitType.PRODUCT,
        "ORG": UnitType.ENTITY,
        "PERSON": UnitType.ENTITY,
        "GPE": UnitType.ENTITY,
        "EVENT": UnitType.ENTITY,
        "WORK_OF_ART": UnitType.ENTITY,
        "LAW": UnitType.ENTITY,
        "LANGUAGE": UnitType.ENTITY,
        "NORP": UnitType.ENTITY,
        "FAC": UnitType.ENTITY,
        "LOC": UnitType.ENTITY,
    }

    def __init__(self, cache: SpacyDocumentCache):
        self._cache = cache

    def process(self, html_content: str, context: ProcessorContext) -> list[Unit]:
        text = strip_html(html_content)
        doc = self._cache.get(text)
        units: list[Unit] = []
        sentences = list(doc.sents)

        for ent in doc.ents:
            unit_type = self.ENTITY_MAPPING.get(ent.label_)
            if not unit_type:
                continue

            spans = []
            lemmas = []
            for token in ent:
                if token.is_space or token.pos_ in ["PUNCT", "SPACE"]:
                    continue
                if token.pos_ == "DET":
                    continue
                if token.text == "'s" or token.dep_ == "case":
                    continue

                span_type = "symbol" if token.pos_ in ["SYM", "NUM"] else "word"
                spans.append(
                    UnitSpan(
                        position=len(spans),
                        span_type=span_type,
                        start=token.idx,
                        end=token.idx + len(token.text),
                        text=token.text,
                    )
                )
                lemmas.append(token.lemma_.lower())

            if len(spans) < 2:
                continue

            sent_idx = next((i for i, s in enumerate(sentences) if ent.start >= s.start and ent.end <= s.end), 0)
            base_form = " ".join(lemmas)

            units.append(
                Unit(
                    unit_type=unit_type,
                    base_form=base_form,
                    sentence_idx=sent_idx,
                    metadata={"label": ent.label_},
                    language=context.language,
                    spans=spans,
                )
            )

            if base_form not in context.base_forms:
                context.base_forms[base_form] = BaseForm(text=base_form)

        return units
