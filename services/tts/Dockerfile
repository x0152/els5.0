# syntax=docker/dockerfile:1.7
FROM python:3.12-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends espeak-ng && rm -rf /var/lib/apt/lists/*

COPY service/requirements.txt .
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install torch --index-url https://download.pytorch.org/whl/cpu
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install -r requirements.txt

RUN useradd --create-home --uid 10001 appuser
USER appuser
ENV HF_HOME=/home/appuser/.cache/huggingface
ARG TTS_MODEL_ID=KittenML/kitten-tts-mini-0.8
ENV TTS_MODEL_ID=$TTS_MODEL_ID
RUN python -c "from huggingface_hub import snapshot_download; import os; snapshot_download(os.environ['TTS_MODEL_ID'])"

COPY --chown=appuser:appuser service/ ./service/

CMD ["uvicorn", "service.app:app", "--host", "0.0.0.0", "--port", "8000"]
