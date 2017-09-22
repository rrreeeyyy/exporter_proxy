FROM golang:1.9

WORKDIR /go/src/app
COPY . .

RUN go-wrapper download
RUN go-wrapper install

ENTRYPOINT ["go-wrapper", "run"]
