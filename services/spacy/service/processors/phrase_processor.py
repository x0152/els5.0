from service.enums import UnitType
from service.spacy_cache import SpacyDocumentCache
from service.text_utils import strip_html
from service.types import BaseForm, ProcessorContext, Unit, UnitSpan


class PhraseProcessor:
    def __init__(self, cache: SpacyDocumentCache):
        self._cache = cache

    def process(self, html_content: str, context: ProcessorContext) -> list[Unit]:
        text = strip_html(html_content)
        doc = self._cache.get(text)
        units: list[Unit] = []
        sentences = list(doc.sents)

        for chunk in doc.noun_chunks:
            if len(chunk) < 2:
                continue

            spans = []
            for token in chunk:
                if token.is_space or token.pos_ in ["PUNCT", "SPACE"]:
                    continue
                spans.append(
                    UnitSpan(
                        position=len(spans),
                        span_type="word",
                        start=token.idx,
                        end=token.idx + len(token.text),
                        text=token.text,
                    )
                )

            if len(spans) < 2:
                continue

            sent_idx = next(
                (si for si, s in enumerate(sentences) if chunk.start >= s.start and chunk.end <= s.end), 0
            )
            base_form = " ".join(t.lemma_ for t in chunk if not t.is_space and t.pos_ not in ["PUNCT", "SPACE"])

            units.append(
                Unit(
                    unit_type=UnitType.PHRASE,
                    base_form=base_form,
                    sentence_idx=sent_idx,
                    metadata={"root": chunk.root.text, "root_dep": chunk.root.dep_},
                    language=context.language,
                    spans=spans,
                )
            )

            if base_form not in context.base_forms:
                context.base_forms[base_form] = BaseForm(text=base_form)

        return units
