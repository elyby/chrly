FROM golang:1.7

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

COPY ./src /go/src/app

RUN go-wrapper download
RUN go-wrapper install

EXPOSE 80

VOLUME ["/go/src/app"]

CMD ["go-wrapper", "run"]
