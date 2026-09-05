FROM python:3.11-slim AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

COPY service/ai-security/requirements.txt .
RUN pip install --user --no-cache-dir -r requirements.txt

FROM python:3.11-slim

WORKDIR /app

COPY --from=builder /root/.local /root/.local

# Copy source code
COPY service/ai-security/ /app/

ENV PATH=/root/.local/bin:$PATH
ENV PYTHONPATH=/app/generated/ai_security:${PYTHONPATH:-}

EXPOSE 50051

CMD ["python", "main.py"]
