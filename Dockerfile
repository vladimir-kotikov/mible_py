FROM python:alpine
WORKDIR /wheels

RUN apk add --no-cache build-base glib-dev
COPY requirements.txt .
RUN pip install -U pip wheel && pip wheel -r requirements.txt

FROM python:alpine
WORKDIR /app

RUN apk add --no-cache glib
COPY --from=0 /wheels /wheels
RUN pip install -r /wheels/requirements.txt -f /wheels \
    && rm -rf /wheels /root/.cache/pip/*

COPY *.py ./

ENV broker_address=localhost
ENV mible_address=

ENTRYPOINT [ "python", "main.py"]
