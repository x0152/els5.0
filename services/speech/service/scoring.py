import json
import os
import re

import numpy as np
import torch
import torchaudio.functional as AF
from g2p_en import G2p
from huggingface_hub import hf_hub_download
from transformers import Wav2Vec2FeatureExtractor, Wav2Vec2ForCTC

ARPA_IPA = {
    "AA": ["ɑː", "ɑ", "a"],
    "AE": ["æ", "a"],
    "AH0": ["ə", "ʌ"],
    "AH": ["ʌ", "ə"],
    "AO": ["ɔː", "ɔ"],
    "AW": ["aʊ"],
    "AY": ["aɪ"],
    "B": ["b"],
    "CH": ["tʃ", "ʧ"],
    "D": ["d"],
    "DH": ["ð"],
    "EH": ["ɛ", "e"],
    "ER0": ["ɚ", "ə˞", "əɹ", "ə"],
    "ER": ["ɜː", "ɝ", "ɚ", "ə˞", "ɜɹ", "ɜ"],
    "EY": ["eɪ"],
    "F": ["f"],
    "G": ["ɡ", "g"],
    "HH": ["h"],
    "IH": ["ɪ", "i"],
    "IY": ["iː", "i"],
    "JH": ["dʒ", "ʤ"],
    "K": ["k"],
    "L": ["l"],
    "M": ["m"],
    "N": ["n"],
    "NG": ["ŋ"],
    "OW": ["oʊ", "əʊ", "o"],
    "OY": ["ɔɪ"],
    "P": ["p"],
    "R": ["ɹ", "r"],
    "S": ["s"],
    "SH": ["ʃ"],
    "T": ["t"],
    "TH": ["θ"],
    "UH": ["ʊ"],
    "UW": ["uː", "u"],
    "V": ["v"],
    "W": ["w"],
    "Y": ["j"],
    "Z": ["z"],
    "ZH": ["ʒ"],
}

VOWEL_CHARS = set("ɑaæəʌɔɒeɛɪiɚɜɝoʊuʉɐyᵻɨ")

STRIP_MARKS = dict.fromkeys(map(ord, "ːʰ\u0303\u0329\u030d\u0325\u0361\u032f\u0334"))
CANON = {"ɝ": "ɚ", "ɜ": "ɚ", "ə˞": "ɚ", "ɜ˞": "ɚ", "əʊ": "oʊ", "ʧ": "tʃ", "ʤ": "dʒ", "g": "ɡ", "ɦ": "h"}
NEAR = [
    {"i", "ɪ", "ᵻ", "ɨ"},
    {"u", "ʊ", "ʉ"},
    {"ə", "ʌ", "ɐ"},
    {"ə", "ɚ"},
    {"ɑ", "ɒ", "a"},
    {"ɔ", "ɑ"},
    {"ɛ", "e"},
    {"e", "eɪ"},
    {"ɹ", "ɾ"},
    {"t", "ɾ"},
]


def canon(p):
    p = p.translate(STRIP_MARKS)
    return CANON.get(p, p)


def is_vowel(p):
    return p[0] in VOWEL_CHARS


def similarity(a, b):
    ca, cb = canon(a), canon(b)
    if ca == cb:
        return 1.0
    if any(ca in g and cb in g for g in NEAR):
        return 0.85
    if is_vowel(ca) == is_vowel(cb):
        return 0.3 if is_vowel(ca) else 0.2
    return 0.0


def align_sequences(ref, obs):
    n, m = len(ref), len(obs)
    gap = -0.45
    score = np.zeros((n + 1, m + 1))
    move = np.zeros((n + 1, m + 1), dtype=int)
    for i in range(1, n + 1):
        score[i][0], move[i][0] = i * gap, 1
    for j in range(1, m + 1):
        score[0][j], move[0][j] = j * gap, 2
    for i in range(1, n + 1):
        for j in range(1, m + 1):
            options = [
                (score[i - 1][j - 1] + similarity(ref[i - 1], obs[j - 1]), 0),
                (score[i - 1][j] + gap, 1),
                (score[i][j - 1] + gap, 2),
            ]
            score[i][j], move[i][j] = max(options)
    pairs = []
    i, j = n, m
    while i > 0 or j > 0:
        if i > 0 and j > 0 and move[i][j] == 0:
            pairs.append((i - 1, j - 1))
            i, j = i - 1, j - 1
        elif i > 0 and (j == 0 or move[i][j] == 1):
            pairs.append((i - 1, None))
            i -= 1
        else:
            pairs.append((None, j - 1))
            j -= 1
    pairs.reverse()
    return pairs


class Assessor:
    def __init__(self, model_id):
        token = os.environ.get("HF_TOKEN") or None
        self.model_id = model_id
        self.extractor = Wav2Vec2FeatureExtractor.from_pretrained(model_id, token=token)
        self.model = Wav2Vec2ForCTC.from_pretrained(model_id, token=token).eval()
        with open(hf_hub_download(model_id, "vocab.json", token=token)) as f:
            self.vocab = json.load(f)
        self.id2tok = {i: t for t, i in self.vocab.items()}
        self.blank = self.vocab.get("<pad>", 0)
        self.g2p = G2p()

    def to_ipa(self, arpa):
        key = arpa if arpa in ARPA_IPA else arpa.rstrip("012")
        for cand in ARPA_IPA.get(key, []):
            if cand in self.vocab:
                return cand
        return None

    def word_phonemes(self, word, unmapped):
        result = []
        for ph in self.g2p(word):
            ph = ph.strip()
            if not ph or not ph[0].isalpha():
                continue
            ipa = self.to_ipa(ph)
            if ipa:
                result.append(ipa)
            else:
                unmapped.append(ph)
        return result

    def assess(self, samples, text, strictness=1.0):
        words = re.findall(r"[A-Za-z']+", text)
        ref, word_of, unmapped = [], [], []
        for wi, w in enumerate(words):
            for p in self.word_phonemes(w, unmapped):
                ref.append(p)
                word_of.append(wi)
        if not ref:
            return {"error": "no phonemes for text"}

        inputs = self.extractor(samples, sampling_rate=16000, return_tensors="pt")
        with torch.no_grad():
            logits = self.model(inputs.input_values).logits
        logp = torch.log_softmax(logits, dim=-1)
        if logp.shape[1] < len(ref):
            return {"error": "audio too short for this text"}

        targets = torch.tensor([[self.vocab[p] for p in ref]])
        aligned, fa_scores = AF.forced_align(logp, targets, blank=self.blank)
        spans = AF.merge_tokens(aligned[0], fa_scores[0].exp(), blank=self.blank)
        gop = [float(s.score) for s in spans]

        obs = []
        prev = self.blank
        for i in logp[0].argmax(-1).tolist():
            if i != prev and i != self.blank:
                tok = self.id2tok.get(i, "")
                if tok and not tok.startswith("<") and tok.strip():
                    obs.append(tok)
            prev = i

        pairs = align_sequences(ref, obs)
        heard = {ri: oi for ri, oi in pairs if ri is not None and oi is not None}
        extras, last_ref = [], 0
        for ri, oi in pairs:
            if ri is not None:
                last_ref = ri
            elif oi is not None:
                extras.append((last_ref, obs[oi]))

        out_words = [{"word": w, "phonemes": [], "extra": []} for w in words]
        scores = []
        for idx, p in enumerate(ref):
            o = obs[heard[idx]] if idx in heard else None
            g = gop[idx] if idx < len(gop) else 0.0
            sim = similarity(p, o) if o else 0.0
            sc = sim * (0.4 + 0.6 * g)
            if sim >= 1.0:
                sc = max(sc, 0.8)
            sc = sc ** strictness
            verdict = (
                "missing" if not o
                else "good" if sc >= 0.6
                else "close" if sc >= 0.35
                else "wrong"
            )
            out_words[word_of[idx]]["phonemes"].append(
                {"expected": p, "heard": o, "score": round(sc, 2), "verdict": verdict}
            )
            scores.append(sc)
        for after, tok in extras:
            out_words[word_of[after]]["extra"].append(tok)
        for w in out_words:
            ps = [p["score"] for p in w["phonemes"]]
            w["score"] = round(float(np.mean(ps)) * 100) if ps else None
            w["ipa"] = " ".join(p["expected"] for p in w["phonemes"])

        result = {
            "overall": round(float(np.mean(scores)) * 100),
            "heard": " ".join(obs),
            "words": out_words,
        }
        if unmapped:
            result["unmapped"] = sorted(set(unmapped))
        return result
