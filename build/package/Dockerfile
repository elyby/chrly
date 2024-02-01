# syntax=docker/dockerfile:1
ARG BINARY

FROM scratch

EXPOSE 80

COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ${BINARY} /usr/local/bin/chrly

ENTRYPOINT ["/usr/local/bin/chrly"]
CMD ["serve"]
