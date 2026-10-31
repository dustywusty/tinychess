# Your Move on OVH

The website and mobile API share `https://yourmove.fun`.
The OVH VPS is `vps-ea9f56f0.vps.ovh.us` (`40.160.152.121`).
The SSH user is `ubuntu`.

The host Caddy service forwards requests to `127.0.0.1:18080`.
The application serves the tested frontend files and API from one container.
PostgreSQL runs in a separate container with a persistent volume.
The database port is not published on the host.
The application database user has no PostgreSQL administrator permissions.

The application has a 512 MB memory limit.
PostgreSQL has a 768 MB memory limit.
Each container has a one-CPU limit.
The two Discourse installations keep their existing containers and databases.

## Installation

Copy this directory to `/opt/yourmove` on the server.
Set executable permissions on `deploy.sh`, `backup.sh`, and `init-db.sh`.
Create `/opt/yourmove/.env` from `.env.example` with mode `600` and root ownership.
Generate separate random hexadecimal passwords for the administrator and application database users.
Set `POSTGRES_IMAGE` to a PostgreSQL 18 image digest.

Deploy a tested backend and frontend from the same successful CI run:

```sh
sudo /opt/yourmove/deploy.sh \
  ghcr.io/dustywusty/tinychess@sha256:BACKEND_DIGEST \
  ghcr.io/dustywusty/tinychess@sha256:FRONTEND_DIGEST \
  FULL_GIT_COMMIT
```

The script checks both image revisions and the frontend assets.
For subsequent deployments, it backs up the database before starting the new application.
It retains the previous environment file and restores it if startup or the version check fails.
An automatic rollback does not reverse database migrations.
Check migration compatibility before deploying schema changes.

Copy `Caddyfile` to `/etc/caddy/yourmove.caddy`.
Add `import /etc/caddy/yourmove.caddy` to the existing host Caddyfile.
Keep the existing forum routes.
Check the complete configuration before reloading Caddy:

```sh
sudo caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile
sudo systemctl reload caddy
```

Set the Gandi DNS records:

| Name | Type | Value |
| --- | --- | --- |
| `@` | `A` | `40.160.152.121` |
| `www` | `CNAME` | `yourmove.fun.` |

Remove old web records that conflict with these records.
Keep mail, verification, and unrelated records.
Do not retain an apex AAAA record pointing to the previous host.
Caddy obtains public HTTPS certificates after DNS points to this server.

## Checks

From the repository, run the deployment smoke test:

```sh
node scripts/smoke.mjs https://yourmove.fun --with-static
```

The test creates a game and checks player seats, moves, captures, SSE, reactions, website routes, and engine assets.
Check that `/api/version` reports the intended commit.
Check a two-player game from the website and the installed mobile app.

Mobile builds must use `EXPO_PUBLIC_API_URL=https://yourmove.fun`.
Existing builds with this origin require no connection change.
Build a new APK if an installed build uses another origin.

## Automatic releases

After all CI checks pass on `main`, GitHub deploys the tested backend and frontend image digests to OVH.
Pull requests run checks without deploying.
The workflow checks the public version, website, engine assets, and game API after deployment.

The deploy job runs in the `production` GitHub environment, which only `main` can use.
Store the deployment key there, not as a repository secret, so pull requests and other branches cannot read it.

Configure these GitHub values:

| Kind | Name | Value |
| --- | --- | --- |
| `production` environment secret | `OVH_DEPLOY_SSH_KEY` | Dedicated deployment private key |
| Variable | `OVH_SSH_HOST` | `vps-ea9f56f0.vps.ovh.us` |
| Variable | `OVH_SSH_KNOWN_HOSTS` | Verified SSH host key line |

Install `ci-command.sh` as `/usr/local/bin/yourmove-ci-deploy`, owned by root with mode `755`.
Add the deployment public key to the `ubuntu` authorized keys with this prefix:

```text
restrict,command="/usr/local/bin/yourmove-ci-deploy" ssh-ed25519 PUBLIC_KEY yourmove-github-actions
```

This key accepts only pinned images from the Tinychess package and a full commit.
It cannot open an interactive shell or forward ports.
The server deployment script checks that both image revisions match the requested commit.
Host scripts require a separate administrative update when their implementation changes.

## Backups

Install the timer and service:

```sh
sudo cp /opt/yourmove/yourmove-backup.{service,timer} /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now yourmove-backup.timer
sudo systemctl start yourmove-backup.service
```

Daily backups run after 06:00 UTC and remain in `/opt/yourmove/backups` for 14 days.
The backup script checks each archive before publishing the completed file.
These backups remain on the same VPS and do not protect against losing that server.
Keep a separate off-server copy of production backups.

To create a manual backup:

```sh
sudo /opt/yourmove/backup.sh
```

## Recovery

Database storage survives application restarts and container replacement.
Never run `docker compose down --volumes` against this deployment.
That command removes the database volume.

To roll back a release, pass its retained image digests and commit to `deploy.sh`.
Check database schema compatibility before rollback.
The release directories retain earlier frontend files.

The former DigitalOcean database is separate from this installation.
Creating this database does not recover games from the previous server.
Restore a database backup before accepting traffic if previous games must remain available.
