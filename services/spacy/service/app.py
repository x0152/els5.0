import dataclasses
from fastapi import FastAPI
from pydantic import BaseModel

from service.spacy_cache import SpacyDocumentCache
from service.spacy_analyzer import SpacyAnalyzer
from service.grammar_analyzer import DeterministicGrammarAnalyzer
from service.processors.entity_processor import EntityProcessor
from service.processors.compound_processor import CompoundProcessor
from service.processors.phrasal_verb_processor import PhrasalVerbProcessor
from service.processors.phrase_processor import PhraseProcessor
from service.processors.technical_processor import TechnicalProcessor

app = FastAPI()

cache = SpacyDocumentCache()
analyzer = SpacyAnalyzer(cache)
grammar_analyzer = DeterministicGrammarAnalyzer()
processors = [
    EntityProcessor(cache),
    CompoundProcessor(cache),
    PhrasalVerbProcessor(cache),
    PhraseProcessor(cache),
    TechnicalProcessor(),
]


class AnalyzeRequest(BaseModel):
    html_content: str


@app.post("/analyze")
def analyze(req: AnalyzeRequest):
    context = analyzer.analyze(req.html_content)

    for proc in processors:
        new_units = proc.process(req.html_content, context)
        context.units.extend(new_units)

    return {
        "sentence_count": context.sentence_count,
        "language": context.language,
        "units": [dataclasses.asdict(u) for u in context.units],
        "base_forms": {k: dataclasses.asdict(v) for k, v in context.base_forms.items()},
    }


@app.post("/analyze_grammar")
def analyze_grammar(req: AnalyzeRequest):
    return {"constructions": grammar_analyzer.analyze(req.html_content)}
