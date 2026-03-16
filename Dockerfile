FROM golang:1.26 AS builder

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux make

FROM alpine:latest
COPY --from=builder /go/src/app/bin/exporter_proxy /exporter_proxy

USER nobody
EXPOSE 9099

CMD ["/exporter_proxy", "-config", "/config.yml"]
