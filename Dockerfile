FROM alpine:3.9.3

EXPOSE 80

RUN apk add --no-cache ca-certificates

ENV STORAGE_REDIS_HOST=redis
ENV STORAGE_FILESYSTEM_HOST=/data

COPY docker-entrypoint.sh /usr/local/bin/
COPY release/chrly /usr/local/bin/

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["serve"]
