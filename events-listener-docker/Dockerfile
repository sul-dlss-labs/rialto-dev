FROM docker:stable

RUN apk add --no-cache bash
COPY listen-docker-events.sh /var/listen-docker-events.sh
CMD ["/bin/bash", "/var/listen-docker-events.sh"]
