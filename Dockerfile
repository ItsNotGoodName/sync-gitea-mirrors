FROM alpine
ENTRYPOINT ["/usr/bin/sync-gitea-mirrors"]
WORKDIR /config
COPY sync-gitea-mirrors /usr/bin/sync-gitea-mirrors
