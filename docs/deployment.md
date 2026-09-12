# Deploy Tinychess and build an Android APK

The backend container serves the Go API on port 8080.
A separate App Platform Static Site serves the compiled React website and Arasan engine assets.
The Android app connects to the same API through HTTPS.
The runtime uses a non-root user and includes no shell or package manager.
DigitalOcean App Platform is the primary deployment target. The `.do/app.yaml` file defines both components.

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

The `Container` workflow tests pull requests and builds an AMD64 image without publishing to a registry.
Each successful run includes a `tinychess-linux-amd64` download artifact, retained for seven days.

To test that artifact on an AMD64 Docker host, download and extract the artifact ZIP.
Then load the image:

```sh
docker load --input tinychess-linux-amd64.tar.gz
```

The imported image name is `tinychess:ci`.
This artifact is for local Docker testing. App Platform builds the backend from the repository Dockerfile.

After successful tests, pushes to `main`, version tags, and manual runs publish AMD64 and ARM64 images to GHCR.
Pull requests never publish registry images.
The workflow uses the repository `GITHUB_TOKEN` with `packages: write` permission.
It does not require a separate registry token.
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
Publishing an image does not deploy it to DigitalOcean App Platform.

## DigitalOcean App Platform

The spec updates the existing `hammerhead-app` in `nyc`, with app ID `e80ec2bc-7a78-4b44-b856-d2c26a1a1ca5`.
The primary domain is `yourmove.fun`. The previous domain, `pawnd.dusty.wtf`, remains an alias.
Both components use the `main` branch of `dustywusty/tinychess`.

| Component | Source | Public routes |
| --- | --- | --- |
| `tinychess` | `Dockerfile`, Go API only | `/api/*` |
| `frontend` | `Dockerfile.frontend`, static output in `/site` | `/` and game links |

The frontend build compiles Arasan with Emscripten and builds the Vite client from `web/`.
The static site serves `index.html` as the fallback for `/g/<game-id>` and legacy game links.
The API ingress preserves the `/api` prefix. Web and mobile clients use the same public HTTPS origin.
The backend retains one 512 MB instance and a direct `/healthz` probe on port 8080.
The public version endpoint is `/api/version`.

### Automatic deployment

The `CI` workflow deploys after a push to `main`, including a pull request merge.
It requires the application checks, backend image check, and static frontend build to pass.
The application checks include Postgres recovery tests and web/mobile browser regressions.
Pull requests and other branches run checks without deployment.
A manual workflow run on `main` also permits deployment.

GitHub Actions needs these repository settings:

| Kind | Name | Value |
| --- | --- | --- |
| Secret | `DIGITALOCEAN_ACCESS_TOKEN` | A token with permission to read and update the app. |
| Variable | `DIGITALOCEAN_APP_ID` | `e80ec2bc-7a78-4b44-b856-d2c26a1a1ca5` |

The workflow validates the spec, updates both components, waits for deployment, and checks `/api/version` through the starter domain.
Deployments run one at a time. An older queued run skips deployment if `main` already has a newer commit.
DigitalOcean's direct GitHub deployment trigger is disabled so it cannot bypass these checks.
The separate `Container` workflow can publish backend images to GHCR. App Platform does not require those published images.

### Manual deployment

Push the code to the branch named in both components of `.do/app.yaml`.
Then validate and update the existing app:

```sh
doctl auth init
doctl apps propose --app e80ec2bc-7a78-4b44-b856-d2c26a1a1ca5 --spec .do/app.yaml
doctl apps update e80ec2bc-7a78-4b44-b856-d2c26a1a1ca5 --spec .do/app.yaml --update-sources --wait
```

The update starts a deployment. It does not create another app.
See the [App Platform spec reference](https://docs.digitalocean.com/products/app-platform/reference/app-spec/).

Check the public site and API after deployment:

```sh
curl --fail https://yourmove.fun/
curl --fail https://yourmove.fun/api/version
```

Set `EXPO_PUBLIC_API_URL` to `https://yourmove.fun` when building the APK.
Do not include `/api` or an internal port in this origin.

### Other Docker hosts

The backend image contains no web assets. Build or export the matching static files separately.
The Compose example mounts `web/dist` and starts the backend with `-static-dir /site` for hosts that need one public service.
Build `web/dist` before starting Compose. For App Platform, the Static Site component serves these files instead.

### Postgres recovery

Use a dedicated Tinychess database and database user on your existing PostgreSQL cluster.
Add its `DATABASE_URL` as an encrypted runtime variable on the service.
Use the database provider's TLS configuration and restrict database access to the app.
The spec omits this variable so credentials cannot enter Git accidentally.
Startup adds a nullable `games.live_state` JSONB column through GORM auto-migration.
The database user requires permission to migrate the Tinychess tables.
Connection or migration failure prevents startup.

CAUTION: Back up the database before the first upgrade. Finish games that started on the older server before replacement.

Older rows lack saved player identities and cannot recover safely.
The API returns HTTP 409 for those games and preserves their existing review history.
New games created with this version support recovery.
Invalid recovery data also returns HTTP 409, without overwriting the saved game.
An unavailable database returns HTTP 503 instead of creating an empty game or accepting an unsaved move.

### Deployments and live games

The anonymous client ID acts as a bearer credential. Anyone with that ID can act as its player.
New identities use cryptographically random UUIDv4 values.
Web stores the ID in localStorage. Native apps store it in AsyncStorage.
Existing IDs remain valid, including IDs from older app builds.
The website migrates a legacy tab identity when no persistent identity exists.

CAUTION: Keep client IDs out of shared links, screenshots, analytics, and access logs.

The current SSE transport sends the credential in its connection query string.
Proxy access logs must omit query strings. Application logs do not record these request URLs.
Emoji events give other viewers a public alias, not the player credential.
Snapshot and stream responses prohibit caching.
Account recovery and identity transfer between devices are not available yet.

Deployments interrupt SSE connections. After reconnect, clients restore the saved position and their original seats.
Missed emoji reactions do not replay.
One instance still has temporary downtime during replacement.
Keep scaling fixed at one instance until shared event delivery is implemented.

## Update or roll back

Merge tested changes into `main` to deploy both App Platform components.
Without Postgres, finish active games before replacement because restarts erase in-memory games.
Back up Postgres before updates if persistence is enabled.

For rollback, revert the release commit through a pull request and merge it into `main`.
The deployment workflow rebuilds both components from that revision.
Check database schema compatibility before rollback because startup migrations can change the schema.
Do not roll back active games to a version without recovery support.
After deployment, check the website, `/api/version`, and a two-player game.

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
