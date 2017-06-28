FROM golang:1.9-alpine

RUN apk add --no-cache git
RUN apk add --no-cache git curl \
 && curl https://glide.sh/get | sh \
 && apk del curl

RUN mkdir -p /go/src/elyby/minecraft-skinsystem \
             /go/src/elyby/minecraft-skinsystem/data/capes \
 && ln -s /go/src/elyby/minecraft-skinsystem /go/src/app

WORKDIR /go/src/app

COPY ./glide.* /go/src/app/

RUN glide install

COPY ./minecraft-skinsystem.go /go/src/app/
COPY ./lib /go/src/app/lib

RUN go build minecraft-skinsystem.go \
 && mv minecraft-skinsystem /usr/local/bin/

EXPOSE 80

VOLUME ["/go/src/app"]

CMD ["minecraft-skinsystem"]
