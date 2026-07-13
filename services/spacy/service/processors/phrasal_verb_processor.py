from service.enums import UnitType
from service.spacy_cache import SpacyDocumentCache
from service.text_utils import strip_html
from service.types import BaseForm, ProcessorContext, Unit, UnitSpan


class PhrasalVerbProcessor:
    def __init__(self, cache: SpacyDocumentCache):
        self._cache = cache

    def process(self, html_content: str, context: ProcessorContext) -> list[Unit]:
        text = strip_html(html_content)
        doc = self._cache.get(text)
        units: list[Unit] = []
        sentences = list(doc.sents)
        used_tokens = set()

        for token in doc:
            if token.pos_ != "VERB" or token.i in used_tokens:
                continue

            particles = []
            particle_tokens = []

            for child in token.children:
                if child.dep_ == "prt":
                    particles.append(child.text.lower())
                    particle_tokens.append(child)
                elif child.dep_ == "prep" and child.pos_ == "ADP" and abs(child.i - token.i) <= 3:
                    particles.append(child.text.lower())
                    particle_tokens.append(child)
                    break

            if not particles:
                continue

            all_tokens = sorted([token] + particle_tokens, key=lambda t: t.idx)
            if any(t.i in used_tokens for t in all_tokens):
                continue

            spans = [
                UnitSpan(
                    position=i,
                    span_type="word",
                    start=t.idx,
                    end=t.idx + len(t.text),
                    text=t.text,
                )
                for i, t in enumerate(all_tokens)
            ]

            sent_idx = next((si for si, s in enumerate(sentences) if token.i >= s.start), 0)
            base_form = f"{token.lemma_} {' '.join(particles)}"

            units.append(
                Unit(
                    unit_type=UnitType.PHRASAL,
                    base_form=base_form,
                    sentence_idx=sent_idx,
                    metadata={"verb": token.lemma_, "particles": particles},
                    language=context.language,
                    spans=spans,
                )
            )

            if base_form not in context.base_forms:
                context.base_forms[base_form] = BaseForm(text=base_form)

            used_tokens.update(t.i for t in all_tokens)

        return units
