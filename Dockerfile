FROM golang:1.7

RUN mkdir -p /go/src/elyby/minecraft-skinsystem \
             /go/src/elyby/minecraft-skinsystem/data/capes \
 && ln -s /go/src/elyby/minecraft-skinsystem /go/src/app

WORKDIR /go/src/app

COPY ./minecraft-skinsystem.go /go/src/app/
COPY ./lib /go/src/app/lib

RUN go-wrapper download
RUN go-wrapper install

EXPOSE 80

VOLUME ["/go/src/app"]

CMD ["go-wrapper", "run"]
