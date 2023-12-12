# syntax=docker/dockerfile:1

FROM golang:1.21-alpine AS builder

COPY . /build
WORKDIR /build
RUN go mod download

RUN CGO_ENABLED=0 \
    go build \
    -trimpath \
    -ldflags="-w -s" \
    -o chrly \
    main.go

FROM alpine:3.19

EXPOSE 80
ENV STORAGE_REDIS_HOST=redis
ENV STORAGE_FILESYSTEM_HOST=/data

COPY docker-entrypoint.sh /
COPY --from=builder /build/chrly /usr/local/bin/chrly

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["serve"]
