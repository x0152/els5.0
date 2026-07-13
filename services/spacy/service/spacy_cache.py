import hashlib
from typing import Any

import spacy


class SpacyDocumentCache:
    def __init__(self, model: str = "en_core_web_trf"):
        self._nlp = spacy.load(model)
        self._nlp.max_length = 2_000_000
        self._cache: dict[str, Any] = {}

    def get(self, text: str) -> Any:
        key = self._hash(text)
        if key not in self._cache:
            self._cache[key] = self._nlp(text)
        return self._cache[key]

    def clear(self) -> None:
        self._cache.clear()

    def _hash(self, text: str) -> str:
        return hashlib.md5(text.encode()).hexdigest()
