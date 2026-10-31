# Deploy Tinychess and build an Android APK

Production runs on the OVH VPS behind `yourmove.fun`.
Use the [OVH deployment instructions](../deploy/ovh/README.md) for the server, database, HTTPS, and backups.

The backend container serves the Go API and the compiled React website on port 8080.
The OVH host publishes it through Caddy at `https://yourmove.fun`.
The Android app connects to the same HTTPS origin.
The runtime uses a non-root user and includes no shell or package manager.
PostgreSQL runs in a separate container with a persistent volume.

## Current limits

CAUTION: Run exactly one application instance. SSE connections and reaction cooldowns remain local to that process.

With `DATABASE_URL`, Postgres stores each accepted move and player seat before the server acknowledges the change.
After a restart, the next request restores the game from its full move history.
Both web and mobile clients recover through the same API.
Players retain their seats with the same browser profile or app installation.
Cleared browser storage or a reinstalled app loses that anonymous identity.

Without `DATABASE_URL`, games remain in memory and disappear after a restart.
The server removes inactive memory entries after 24 hours, but saved games remain in Postgres.
Redis is not required for recovery and is not connected yet.
Multiple replicas still require shared event delivery.
This deployment is suitable for a small testing group, not a highly available service.
Automatic database migrations run at startup when `DATABASE_URL` is set.

## Build and test locally

Requirements: Docker Engine with BuildKit, Node 22.13 or later, and Git.

Run these commands from the repository root:

```sh
make image
make image-test
```

The image test uses a temporary container and a random loopback port.
It checks health, API-only routing, player seats, legal captures, SSE updates, and emoji reactions.
It also checks the non-root user, a read-only filesystem, and a custom `PORT` value.
The test removes only its own container.

## Build and release through GitHub

The `CI` workflow runs once per pull request and after each push to `main`.
It builds the frontend, tests the backend image, and runs the web and mobile browser suites concurrently.
The web browser suite uses the compiled frontend that the workflow later publishes.
Each run retains the `frontend-linux-amd64` and `tinychess-linux-amd64` artifacts for seven days.

To test the backend artifact on an AMD64 Docker host, download and extract the artifact ZIP.
Then load the image:

```sh
docker load --input tinychess-linux-amd64.tar.gz
```

The imported image name is `tinychess:ci`.
CI also packages the compiled frontend with `Dockerfile.static` and checks the extraction Dockerfile against the original files.
This Dockerfile copies files into `/site`. It contains no compiler or dependency installation step.
After all checks pass on `main`, a separate job loads and publishes both tested images without another build.

Both images use the existing public `ghcr.io/dustywusty/tinychess` package:

| Tag | Content |
| --- | --- |
| `deploy-<full-commit>` | Tested AMD64 backend |
| `frontend-<full-commit>` | Compiled frontend and Arasan assets |

The workflow records both immutable digests in its summary.
Deployment uses these digests, so a later tag update cannot change the selected files.
Pull requests never publish images or deploy the app.
The publication job uses `GITHUB_TOKEN` with `packages: write` permission.
The OVH server pulls this public package without registry credentials.

The separate `Container` workflow handles version tags and manual runs.
It retains the AMD64/ARM64 release process, but no longer repeats the normal PR and `main` checks.
Docker documents the [multi-platform build process](https://docs.docker.com/build/ci/github-actions/multi-platform/).

For a release, select an unused semantic version on the tested commit.
Then create and push its tag:

```sh
git tag -a v0.1.0 -m "Tinychess 0.1.0"
git push origin v0.1.0
```

The example tag publishes `ghcr.io/dustywusty/tinychess:0.1.0`.
Each publication also includes a `sha-<full-commit>` tag.
The `edge` tag tracks the default branch, not a stable release.
The workflow summary contains an immutable `ghcr.io/dustywusty/tinychess@sha256:...` reference.

Deploy the digest from the successful workflow summary.
Retain the previous digest for rollback.
If a GHCR package is private, configure registry credentials on the deployment host.
Do not change package visibility without reviewing the intended audience.
The `Container` release workflow only publishes images.
The `CI` workflow publishes and deploys to OVH after successful checks on `main`.

## OVH deployment

The [OVH instructions](../deploy/ovh/README.md) describe installation, credentials, DNS, backups, and rollback.
GitHub deploys the tested image pair after a successful `main` build.
The deployment key accepts only the release command and cannot open a shell.
The host retains the previous image references for rollback.

The production origin remains `https://yourmove.fun` for both web and mobile.
Mobile builds that already use this origin do not require a connection change.
The previous DigitalOcean database was deleted.
The OVH database starts fresh and does not contain earlier games.

The `.do/app.yaml` file and renderer remain as references for the former deployment.
The active workflow no longer uses DigitalOcean credentials or updates its app.

### Persistence and anonymous identities

Postgres stores accepted moves and player seats before the server acknowledges changes.
A reconnect restores the saved position and seats after an application restart.
The anonymous client ID acts as a bearer credential.
Keep client IDs out of shared links and access logs.
The Caddy route does not enable access logging.

Run one application instance because SSE delivery remains local to that process.
Automatic database migrations run at startup.
Back up the database before schema changes and check compatibility before rollback.

## Build the shareable Android APK

The `preview` profile in `apps/mobile/eas.json` produces a standalone APK.
The app does not require Expo Go or Metro.
Expo provides a download link after the build.
See [Expo's APK instructions](https://docs.expo.dev/build-reference/apk/).

After the HTTPS endpoint works, link the app to your Expo project:

```sh
cd apps/mobile
npx eas-cli@latest login
npx eas-cli@latest init
```

Retain the project ID that EAS adds to `app.json`.
Commit that change before subsequent builds.

Set the API origin in the EAS `preview` environment:

```sh
npx eas-cli@latest env:set --name EXPO_PUBLIC_API_URL --value https://yourmove.fun --environment preview --visibility plaintext
```

Use this origin only after the domain passes the HTTPS checks.
Do not append `/api` because the client adds endpoint paths.
The pre-install check rejects a missing origin, HTTP, localhost, credentials, and URL paths.
`EXPO_PUBLIC_*` values are visible in the app bundle. They must not contain secrets.
See [Expo's environment-variable instructions](https://docs.expo.dev/eas/environment-variables/manage/).

For app-to-app testing, leave `EXPO_PUBLIC_WEB_URL` unset.
The share action then uses `yourmove://g/<id>` links for the installed app.
The friend can also paste a game ID or game link into the invite field.
The native configuration names `yourmove.fun`, but verified HTTPS app links still require website association files and app signing identities.
Setting `EXPO_PUBLIC_WEB_URL` alone does not enable verified Android app links.
See Expo's [Android app-link instructions](https://docs.expo.dev/linking/android-app-links/).

Before the EAS upload, build the local engine assets from the repository root:

```sh
pnpm engine:build
pnpm engine:check
```

This step requires Emscripten 4.0.14. See [computer opponents](computer-opponents.md) for toolchain and licensing details.
The EAS archive must include the generated engine assets. The post-install check rejects an archive without them.

From `apps/mobile`, build the APK:

```sh
npx eas-cli@latest build --platform android --profile preview
```

If EAS requests signing credentials, create a keystore for this app or select its existing keystore.
Retain the signing credentials for future upgrades.
Send the build download link to your friend.
Install the APK on both phones.
If Android requests permission, allow installation from the browser used for this download.
Create a game and share its link or ID.

For each app update, build and distribute another APK from the same Expo project and signing key.
The profile increments the Android version code automatically.
Changing the API origin requires another APK build.
The `production` profile produces an AAB for a later Play Store release, not a directly installable APK.
