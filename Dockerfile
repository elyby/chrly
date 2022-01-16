FROM golang:1.14-alpine as builder
WORKDIR /build
COPY . .
RUN go build -v -o chrly

FROM alpine:3.9.3

EXPOSE 80

COPY --from=builder /build/chrly /usr/local/bin
COPY docker-entrypoint.sh /usr/local/bin/

RUN apk add --no-cache ca-certificates

ENV STORAGE_REDIS_HOST=redis
ENV STORAGE_FILESYSTEM_HOST=/data

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["serve"]
