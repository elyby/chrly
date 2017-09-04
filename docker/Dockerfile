FROM alpine:3.6

COPY docker/docker-entrypoint.sh /usr/local/bin/
COPY docker/config.dist.yml /usr/local/etc/minecraft-skinsystem/

COPY minecraft-skinsystem /usr/local/bin/

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["serve"]
