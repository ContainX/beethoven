FROM nginx:mainline

ENV DOCKER_ENV true

RUN apt-get update \
    && apt-get install -y supervisor \
    && rm -rf /var/lib/apt/lists/*

COPY beethoven /usr/sbin/beethoven
COPY scripts/run.sh /bin/run.sh
RUN chmod a+x /bin/run.sh

COPY scripts/supervisord.conf /etc/supervisord.tmpl

ENTRYPOINT ["/bin/run.sh"]
