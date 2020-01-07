FROM golang:alpine
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build

FROM alpine
COPY --from=0 /build/mible /mible

ENV BROKER_ADDRESS=
ENV DEVICE_ADDRESS=
ENV DEVICE_NAME=
ENV UPDATE_INTERVAL=
ENV SENTRY_DSN=
ENV DEBUG=

ENTRYPOINT [ "/mible"]
