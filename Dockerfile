FROM alpine
ENTRYPOINT ["/usr/bin/sync-gitea-mirrors"]
COPY sync-gitea-mirrors /usr/bin/sync-gitea-mirrors
