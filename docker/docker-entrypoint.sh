#!/bin/sh
set -e

CONFIG="/etc/minecraft-skinsystem/config.yml"

if [ ! -f "$CONFIG" ]; then
    mkdir -p $(dirname "${CONFIG}")
    cp /usr/local/etc/minecraft-skinsystem/config.dist.yml "$CONFIG"
fi

if [ "$1" = "serve" ] || [ "$1" = "amqp-worker" ]; then
    set -- minecraft-skinsystem "$@"
fi

exec "$@"
