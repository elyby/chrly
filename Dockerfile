FROM alpine:3.7

EXPOSE 80

ENV STORAGE_REDIS_HOST=redis
ENV STORAGE_FILESYSTEM_HOST=/data

COPY docker-entrypoint.sh /usr/local/bin/
COPY release/chrly /usr/local/bin/

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["serve"]
