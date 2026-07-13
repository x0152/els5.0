from service.enums import POS, Language, UnitType
from service.spacy_cache import SpacyDocumentCache
from service.text_utils import strip_html
from service.types import BaseForm, ProcessorContext, Unit, UnitSpan

try:
    from langdetect import LangDetectException, detect_langs

    LANGDETECT_AVAILABLE = True
except ImportError:
    LANGDETECT_AVAILABLE = False


class SpacyAnalyzer:
    def __init__(self, cache: SpacyDocumentCache):
        self._cache = cache

    def analyze(self, html_content: str) -> ProcessorContext:
        text = strip_html(html_content)
        doc_language = self._detect_language(text)

        doc = self._cache.get(text)
        sentences = list(doc.sents)

        context = ProcessorContext(
            language=doc_language,
            sentence_count=len(sentences),
        )

        self._extract_single_units(doc, sentences, context)
        return context

    def _detect_language(self, text: str, fallback: str = "en") -> str:
        if not LANGDETECT_AVAILABLE or len(text.strip()) < 15:
            return fallback
        try:
            lang_probs = detect_langs(text)
            if lang_probs:
                if lang_probs[0].lang == "en":
                    return "en"
                if lang_probs[0].prob > 0.85:
                    return lang_probs[0].lang
            return fallback
        except (LangDetectException, Exception):
            return fallback

    def _get_sentence_idx(self, token, sentences) -> int:
        for idx, sent in enumerate(sentences):
            if token.i >= sent.start and token.i < sent.end:
                return idx
        return 0

    def _extract_single_units(self, doc, sentences, context: ProcessorContext) -> None:
        for token in doc:
            if token.is_space or not token.text.strip():
                continue
            if token.pos_ in ["PUNCT", "SYM", "SPACE"]:
                continue

            sent_idx = self._get_sentence_idx(token, sentences)
            pos = self._get_pos(token.pos_)
            lang = self._get_language(context.language)

            unit = Unit(
                unit_type=UnitType.SINGLE,
                base_form=token.lemma_,
                pos=pos,
                sentence_idx=sent_idx,
                metadata={"tag": token.tag_, "dep": token.dep_},
                language=lang,
                spans=[
                    UnitSpan(
                        position=0,
                        span_type="word",
                        start=token.idx,
                        end=token.idx + len(token.text),
                        text=token.text,
                    )
                ],
            )
            context.units.append(unit)
            context.used_spans.add((token.idx, token.idx + len(token.text)))

            if token.lemma_ not in context.base_forms:
                context.base_forms[token.lemma_] = BaseForm(
                    text=token.lemma_,
                    pos=pos,
                    is_stop=token.is_stop,
                    language=lang,
                )

    def _get_pos(self, pos_str: str) -> POS | None:
        try:
            return POS(pos_str)
        except ValueError:
            return None

    def _get_language(self, lang_str: str) -> Language:
        try:
            return Language(lang_str)
        except ValueError:
            return Language.EN
