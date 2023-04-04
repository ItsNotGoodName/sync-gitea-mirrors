# sync-gitea-mirrors

Sync and mirror GitHub/Gitea repositories to Gitea.

# Config

| Environment Variable   | Default  | Required          | Description                                                                                                         |
| ---------------------- | -------- | ----------------- | ------------------------------------------------------------------------------------------------------------------- |
| `GITHUB_OWNER`         | ""       | maybe<sub>1</sub> | Owner of GitHub source repositories.                                                                                |
| `GITHUB_TOKEN`         | ""       | maybe<sub>1</sub> | Token for accessing GitHub.                                                                                         |
| `GITEA_OWNER`          | ""       | maybe<sub>2</sub> | Owner of Gitea source repositories.                                                                                 |
| `GITEA_TOKEN`          | ""       | maybe<sub>2</sub> | Token for accessing the source Gitea instance.                                                                      |
| `GITEA_URL`            | ""       | maybe             | URL of the source Gitea instance.                                                                                   |
| `SKIP_REPOS`           | ""       |                   | List of space seperated repositories to not sync (e.g. `ItsNotGoodName/example1 itsnotgoodname/example2 example3`). |
| `SKIP_FORKS`           | false    |                   | Skip fork repositories.                                                                                             |
| `SKIP_PRIVATE`         | false    |                   | Skip private repositories.                                                                                          |
| `MIGRATE_WIKI`         | false    |                   | Migrate wiki from source repositories.                                                                              |
| `SYNC_ALL`             | false    |                   | Sync everything.                                                                                                    |
| `SYNC_TOPICS`          | false    |                   | Sync topics of repository.                                                                                          |
| `SYNC_DESCRIPTION`     | false    |                   | Sync description of repository.                                                                                     |
| `SYNC_VISIBILITY`      | false    |                   | Sync private/public status of repository.                                                                           |
| `SYNC_MIRROR_INTERVAL` | false    |                   | Disable periodic sync if source repository is archived.                                                             |
| `DEST_URL`             | ""       | true              | URL of the destination Gitea instance.                                                                              |
| `DEST_TOKEN`           | ""       | true              | Token for accessing the destination Gitea instance.                                                                 |
| `DEST_OWNER`           | ""       |                   | Owner of the mirrored repositories on the destination Gitea instance.                                               |
| `DEST_MIRROR_INTERVAL` | "8h0m0s" |                   | Default mirror interval for new migrations on the destination Gitea instance.                                       |

# Examples

The following will sync all repositories from GitHub to Gitea located at `https://git.example.com`.

```
sync-gitea-mirrors -github-token="ghp_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" -dest-url="https://git.example.com" -dest-token="BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB" -sync-all -migrate-wiki
```
