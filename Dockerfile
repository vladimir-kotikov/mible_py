FROM golang:alpine
WORKDIR /build

COPY *.go go.mod go.sum ./
RUN go build

FROM alpine
COPY --from=0 /build/mible /mible

ENV broker_address=
ENV device_address=
ENV update_interval=
ENV sentry_dsn=

ENTRYPOINT [ "/mible"]
