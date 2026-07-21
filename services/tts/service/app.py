import io
import os

import soundfile as sf
from fastapi import FastAPI, Response
from kittentts import KittenTTS
from pydantic import BaseModel

MODEL_ID = os.environ.get("TTS_MODEL_ID", "KittenML/kitten-tts-mini-0.8")

app = FastAPI()
model = KittenTTS(MODEL_ID)


class SynthesizeRequest(BaseModel):
    text: str
    voice: str = "Jasper"
    speed: float = 1.0


@app.post("/tts")
def tts(req: SynthesizeRequest):
    audio = model.generate(req.text, voice=req.voice, speed=req.speed, clean_text=True)
    buf = io.BytesIO()
    sf.write(buf, audio, 24000, format="WAV")
    return Response(content=buf.getvalue(), media_type="audio/wav")


@app.get("/voices")
def voices():
    return {"voices": model.available_voices}


@app.get("/healthz")
def healthz():
    return {"ok": True, "model": MODEL_ID}
