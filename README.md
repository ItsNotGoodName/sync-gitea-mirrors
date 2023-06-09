# sync-gitea-mirrors

Sync and mirror GitHub/Gitea repositories to Gitea.

# Config

| Environment Variable       | Default             | Required         | Description                                                                       |
| -------------------------- | ------------------- | ---------------- | --------------------------------------------------------------------------------- |
| `DAEMON`                   | 0                   |                  | Seconds between each run where 0 means running only once (e.g. `86400` is a day). |
| `DAEMON_ERROR`             | 300                 |                  | Seconds between each run when error occurs (e.g. "300" is a 5 minutes).           |
| `DAEMON_SKIP_FIRST`        | false               |                  | Skip first daemon run.                                                            |
| `DAEMON_EXIT_ERROR`        | false               |                  | Exit daemon when error occurs.                                                    |
| `GITHUB_OWNER`<sub>1</sub> | ""                  |                  | Owner of GitHub source repositories.                                              |
| `GITHUB_TOKEN`             | ""                  | true<sub>2</sub> | Token for accessing GitHub.                                                       |
| `GITEA_OWNER`              | ""                  |                  | Owner of Gitea source repositories.                                               |
| `GITEA_TOKEN`              | ""                  | true<sub>2</sub> | Token for accessing the source Gitea instance.                                    |
| `GITEA_URL`                | "https://gitea.com" |                  | URL of the source Gitea instance.                                                 |
| `SKIP_REPOS`               | ""                  |                  | List of space seperated repositories to not sync (e.g. `repo1 repo2 repo3`).      |
| `SKIP_FORKS`               | false               |                  | Skip fork repositories.                                                           |
| `SKIP_PRIVATE`             | false               |                  | Skip private repositories.                                                        |
| `MIGRATE_WIKI`             | false               |                  | Migrate wiki from source repositories.                                            |
| `MIGRATE_LFS`              | false               |                  | Migrate lfs from source repositories.                                             |
| `SYNC_ALL`                 | false               |                  | Sync everything.                                                                  |
| `SYNC_TOPICS`              | false               |                  | Sync topics of repository.                                                        |
| `SYNC_DESCRIPTION`         | false               |                  | Sync description of repository.                                                   |
| `SYNC_VISIBILITY`          | false               |                  | Sync private/public status of repository.                                         |
| `SYNC_MIRROR_INTERVAL`     | false               |                  | Disable periodic sync if source repository is archived.                           |
| `DEST_URL`                 | ""                  | true             | URL of the destination Gitea instance.                                            |
| `DEST_TOKEN`               | ""                  | true             | Token for accessing the destination Gitea instance.                               |
| `DEST_OWNER`               | ""                  |                  | Owner of the mirrored repositories in the destination Gitea instance.             |
| `DEST_MIRROR_INTERVAL`     | "8h0m0s"            |                  | Default mirror interval for new migrations in the destination Gitea instance.     |

1. Setting `GITHUB_OWNER` will only show public repositories.
2. Depends on the selected repository source.

# GitHub to Gitea Example

Sync repositories from GitHub to a Gitea instance that is located at `https://gitea.example.com` on a daily interval.
If a repository does not exist in Gitea then it will create a migration.
It will sync description, topics, and visiblity.
If the GitHub repository is archived then it will set the `mirror-interval` to `0s` in the Gitea repository.

GitHub token can generated [here](https://github.com/settings/tokens).
Scope must be "repo".

## cli

```
sync-gitea-mirrors \
  -github-token="ghp_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" \
  -dest-url="https://gitea.example.com" \
  -dest-token="BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB" \
  -sync-all \
  -daemon=86400
```

## docker-compose

```yaml
version: "3"
services:
  sync-gitea-mirrors:
    container_name: sync-gitea-mirrors
    image: ghcr.io/itsnotgoodname/sync-gitea-mirrors:latest
    environment:
      - GITHUB_TOKEN=ghp_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      - DEST_URL=https://gitea.example.com
      - DEST_TOKEN=BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB
      - SYNC_ALL=true
      - DAEMON=86400
    user: 1000:1000
    restart: unless-stopped
```

## docker cli

```
docker run -d \
  --name=sync-gitea-mirrors \
  --user 1000:1000 \
  --restart unless-stopped \
  -e GITHUB_TOKEN="ghp_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" \
  -e DEST_URL="https://gitea.example.com" \
  -e DEST_TOKEN="BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB" \
  -e SYNC_ALL=true \
  -e DAEMON=86400 \
  ghcr.io/itsnotgoodname/sync-gitea-mirrors:latest
```

# To Do

- Update credentials of Gitea mirrors
- Add GitLab as a repository source
