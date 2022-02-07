# Build binary
FROM golang:1.14-alpine as builder

WORKDIR /build

ARG BUILD_VERSION=unknown
ARG BUILD_COMMIT=unknown
ARG BUILD_TYPE=release

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN export BUILD_TAGS="" \
 && if [ "${BUILD_TYPE}" == "dev" ]; then \
      export BUILD_TAGS="$BUILD_TAGS --tags profiling"; \
    fi \
 && env CGO_ENABLED=0 \
    go build "$BUILD_TAGS" \
    -v \
    -o chrly \
    -ldflags "-X github.com/elyby/chrly/version.version=${BUILD_VERSION} -X github.com/elyby/chrly/version.commit=${BUILD_COMMIT}"

# Build resulting image
FROM alpine:3.9.3

EXPOSE 80

RUN apk add --no-cache ca-certificates

ENV STORAGE_REDIS_HOST=redis
ENV STORAGE_FILESYSTEM_HOST=/data

COPY --from=builder /build/chrly /usr/local/bin
COPY docker-entrypoint.sh /usr/local/bin/

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["serve"]
