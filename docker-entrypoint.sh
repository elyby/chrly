#!/bin/sh
set -e

if [ ! -d /data/capes ]; then
    mkdir -p /data/capes
fi

if [ "$1" = "serve" ] || [ "$1" = "token" ] || [ "$1" = "version" ]; then
    set -- /usr/local/bin/chrly "$@"
fi

exec "$@"
