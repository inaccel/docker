FROM docker/compose:1.29.2
ENV COMPOSE_IGNORE_ORPHANS=true
COPY docker-compose.yml .
COPY .env .
ENTRYPOINT ["docker-compose"]
