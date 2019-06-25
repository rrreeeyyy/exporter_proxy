FROM golang:1.12 as builder
ENV GO111MODULE=on

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...
RUN CGO_ENABLED=0 GOOS=linux make

FROM alpine:latest
COPY --from=builder /go/src/app/bin/exporter_proxy /exporter_proxy

USER nobody
EXPOSE 9099

CMD ["/exporter_proxy", "-config", "/config.yml"]
