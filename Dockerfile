FROM golang:alpine
WORKDIR /build

COPY *.go go.mod go.sum ./
RUN go build

FROM alpine
COPY --from=0 /build/mible /mible

ENV brokeraddress=
ENV deviceaddress=
ENV updateinterval=
ENV sentrydsn=

ENTRYPOINT [ "mible"]
