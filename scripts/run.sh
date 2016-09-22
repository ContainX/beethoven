#!/bin/bash

export BT_ARGS="$@"
export DOCKER_ENV=true
envsubst < /etc/supervisord.tmpl > /etc/supervisor/conf.d/supervisord.conf && /usr/bin/supervisord
