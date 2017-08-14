FROM golang:1.9-alpine

RUN mkdir -p /go/src/elyby/minecraft-skinsystem \
             /go/src/elyby/minecraft-skinsystem/data/capes \
 && ln -s /go/src/elyby/minecraft-skinsystem /go/src/app

WORKDIR /go/src/app

COPY ./Gopkg.* /go/src/app/
COPY ./main.go /go/src/app/
COPY ./cmd /go/src/app/cmd
COPY ./daemon /go/src/app/daemon
COPY ./db /go/src/app/db
COPY ./model /go/src/app/model
COPY ./repositories /go/src/app/repositories
COPY ./ui /go/src/app/ui
COPY ./utils /go/src/app/utils

RUN apk add --no-cache git \
 && go get -u github.com/golang/dep/cmd/dep \
 && dep ensure \
 && go clean -i github.com/golang/dep \
 && rm -rf $GOPATH/src/github.com/golang/dep \
 && apk del git \
 && go build main.go \
 && mv main /usr/local/bin/minecraft-skinsystem

EXPOSE 80

ENTRYPOINT ["minecraft-skinsystem"]
CMD ["serve"]
