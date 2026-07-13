# syntax=docker/dockerfile:1.7
FROM python:3.12-slim

WORKDIR /app

COPY service/requirements.txt .
# Split into separate layers so no single blob exceeds the registry's
# client_max_body_size (push happens per-layer).
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install torch --index-url https://download.pytorch.org/whl/cpu
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install -r requirements.txt
RUN python -m spacy download en_core_web_trf
RUN useradd --create-home --uid 10001 appuser

COPY --chown=appuser:appuser service/ ./service/

USER appuser

CMD ["uvicorn", "service.app:app", "--host", "0.0.0.0", "--port", "8000"]
