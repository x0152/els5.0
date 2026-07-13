from enum import Enum


class POS(str, Enum):
    ADJ = "ADJ"
    ADP = "ADP"
    ADV = "ADV"
    AUX = "AUX"
    CCONJ = "CCONJ"
    DET = "DET"
    INTJ = "INTJ"
    NOUN = "NOUN"
    NUM = "NUM"
    PART = "PART"
    PRON = "PRON"
    PROPN = "PROPN"
    PUNCT = "PUNCT"
    SCONJ = "SCONJ"
    SYM = "SYM"
    VERB = "VERB"
    X = "X"
    SPACE = "SPACE"


class UnitType(str, Enum):
    SINGLE = "single_unit"
    COMPOUND = "compound_unit"
    PHRASE = "phrase_unit"
    ENTITY = "entity_unit"
    PHRASAL = "phrasal_unit"
    IDIOM = "idiom_unit"
    COLLOCATION = "collocation_unit"
    PROVERB = "proverb_unit"
    TIME = "time_unit"
    MONEY = "money_unit"
    QUANTITY = "quantity_unit"
    PERCENT = "percent_unit"
    PRODUCT = "product_unit"
    TECHNICAL = "technical_unit"


class Language(str, Enum):
    EN = "en"
    RU = "ru"
    DE = "de"
    FR = "fr"
    ES = "es"
    IT = "it"
    PT = "pt"
    ZH = "zh"
    JA = "ja"
    KO = "ko"
