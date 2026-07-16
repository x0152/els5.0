# syntax=docker/dockerfile:1.7
FROM python:3.12-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ffmpeg && rm -rf /var/lib/apt/lists/*

COPY service/requirements.txt .
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install torch --index-url https://download.pytorch.org/whl/cpu
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install -r requirements.txt

RUN useradd --create-home --uid 10001 appuser
USER appuser
ENV HF_HOME=/home/appuser/.cache/huggingface
ARG HF_TOKEN
RUN python -c "import nltk; nltk.download('averaged_perceptron_tagger'); nltk.download('cmudict')"
# The model repo is gated: bake it into the image only when a token is provided (prod).
# Without a token (local dev) the model is read from the mounted host HF cache.
RUN if [ -n "$HF_TOKEN" ]; then \
        HF_TOKEN=$HF_TOKEN python -c "from huggingface_hub import snapshot_download; import os; snapshot_download(os.environ.get('SPEECH_MODEL_ID','KoelLabs/xlsr-english-01'), token=os.environ.get('HF_TOKEN') or None)"; \
    fi

COPY --chown=appuser:appuser service/ ./service/

CMD ["uvicorn", "service.app:app", "--host", "0.0.0.0", "--port", "8000"]
