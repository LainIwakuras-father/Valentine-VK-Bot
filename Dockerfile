FROM python:3.13-alpine

WORKDIR /app

COPY pyproject.toml .
RUN pip install . 

# код
COPY src ./src
COPY migrations ./migrations

ENV PYTHONPATH=/app/src
