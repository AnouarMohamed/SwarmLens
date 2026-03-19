FROM python:3.12-slim

WORKDIR /app

RUN groupadd --gid 1001 predictor && \
    useradd --uid 1001 --gid 1001 --no-create-home --shell /usr/sbin/nologin predictor

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app/ ./app/

USER predictor

EXPOSE 8001

HEALTHCHECK --interval=15s --timeout=5s --retries=3 \
  CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8001/health')"

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8001"]
