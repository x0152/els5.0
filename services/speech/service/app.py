import base64
import os
import subprocess

import numpy as np
from fastapi import FastAPI
from pydantic import BaseModel

from service.scoring import Assessor

MODEL_ID = os.environ.get("SPEECH_MODEL_ID", "KoelLabs/xlsr-english-01")

app = FastAPI()
assessor = Assessor(MODEL_ID)


class AssessRequest(BaseModel):
    audio_base64: str
    text: str
    strictness: float = 1.0


@app.post("/assess")
def assess(req: AssessRequest):
    data = base64.b64decode(req.audio_base64)
    proc = subprocess.run(
        ["ffmpeg", "-i", "pipe:0", "-f", "f32le", "-ac", "1", "-ar", "16000", "pipe:1"],
        input=data, capture_output=True,
    )
    samples = np.frombuffer(proc.stdout, dtype=np.float32).copy()
    if len(samples) < 1600:
        return {"error": "empty or unreadable audio"}
    return assessor.assess(samples, req.text, req.strictness)


@app.get("/healthz")
def healthz():
    return {"ok": True, "model": MODEL_ID}
