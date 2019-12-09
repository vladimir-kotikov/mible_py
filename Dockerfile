FROM python:alpine
WORKDIR /wheels

RUN apk add --no-cache build-base glib-dev
COPY requirements.txt .
RUN pip install -U pip wheel && pip wheel -r requirements.txt

FROM python:alpine
WORKDIR /app

COPY --from=0 /wheels /wheels
RUN pip install -r /wheels/requirements.txt -f /wheels \
    && rm -rf /wheels /root/.cache/pip/*

COPY *.py ./
CMD [ "python", "main.py" ]
