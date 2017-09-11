FROM alpine:3.6

RUN apk --update add ca-certificates \
 && update-ca-certificates \
 && rm -rf /var/cache/apk/*

COPY docker/docker-entrypoint.sh /usr/local/bin/
COPY docker/config.dist.yml /usr/local/etc/minecraft-skinsystem/

COPY minecraft-skinsystem /usr/local/bin/

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["serve"]
