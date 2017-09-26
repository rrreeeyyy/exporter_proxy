FROM golang:1.9

RUN go get github.com/lestrrat/go-server-starter/cmd/start_server

WORKDIR /go/src/app
COPY . .

RUN go-wrapper download
RUN go-wrapper install

ENTRYPOINT ["go-wrapper", "run"]
