# Deploy Tinychess and build an Android APK

The container serves the Go API and the compiled website on port 8080.
The Android app connects to the same API through HTTPS.
The runtime uses a non-root user and includes no shell or package manager.
DigitalOcean App Platform is the primary deployment target. The `.do/app.yaml` file defines its service.

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
It checks health, the embedded website, player seats, legal captures, SSE updates, and emoji reactions.
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
This artifact is for local Docker testing. Hosted services use the published registry image.

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
Publishing an image does not deploy it to DigitalOcean App Platform or Railway.

## DigitalOcean App Platform

App Platform runs the container and provides its public HTTPS endpoint.
The website and API use one Web Service component, not a separate Static Site.
The provided spec uses one instance, port 8080, and `/healthz` for health checks.
It does not create a database or enable automatic deployments.
The region defaults to `nyc`. The instance size defaults to `apps-s-1vcpu-1gb`.

The published image includes Linux AMD64, which App Platform requires.
GHCR images support manual deployments through a tag or digest.
See DigitalOcean's [container deployment instructions](https://docs.digitalocean.com/products/app-platform/how-to/deploy-from-container-images/).

### First deployment

1. Publish a tested image with the GitHub release process described in this guide.
2. In DigitalOcean, select **Create → App Platform → Container image**.
3. Select **GitHub Container Registry**.
4. Enter `ghcr.io/dustywusty/tinychess` as the image repository.
5. Select the digest from the successful publishing workflow.
6. For a private image, enter a read-only GHCR credential in the format `username:token`.
7. Select **Web Service**, port `8080`, and one instance.
8. Set the health check path to `/healthz`.
9. Leave build and run commands empty. The image includes its executable and website.
10. Review the region, instance size, and estimated cost before you create the app.

Keep registry credentials out of Git and chat.
The control panel stores the credentials needed to pull private images.
App Platform does not automatically redeploy GHCR images when a tag changes.

Alternatively, copy `.do/app.yaml` and replace its example `tag` with your published `digest`.
Do not specify both fields.
After you authenticate `doctl`, create the app from that spec:

```sh
doctl apps spec validate .do/app.yaml --schema-only
doctl apps create --spec .do/app.yaml --wait
```

The first command checks structure only. It does not check registry access or create resources.
The second command creates billable resources.
For a private image, enter registry credentials through the control panel or a private spec outside Git.
See the [App Platform spec reference](https://docs.digitalocean.com/products/app-platform/reference/app-spec/).

After deployment, use the assigned `https://<app-name>.ondigitalocean.app` URL.
A custom domain is optional for phone testing.
Check the endpoints with your assigned hostname:

```sh
curl --fail https://YOUR-APP.ondigitalocean.app/healthz
curl --fail https://YOUR-APP.ondigitalocean.app/api/version
```

Set the APK's `EXPO_PUBLIC_API_URL` to this same HTTPS origin.
Do not use the internal port, container address, or `/api` path in the APK origin.
The expected health response is `{"ok":true}`.
The version response identifies the server commit.

### Domain: yourmove.fun

The app spec declares `yourmove.fun` as the primary domain without taking over its DNS zone.
The website and API will share `https://yourmove.fun`.

1. In App Platform, add `yourmove.fun` under the app's domain configuration if the spec did not add it.
2. At your DNS provider, add the exact records that DigitalOcean supplies for this domain.
3. Wait for domain verification and the HTTPS certificate.
4. Check `https://yourmove.fun/healthz` and `https://yourmove.fun/api/version` before you build the APK.

No DNS records or cloud resources change when you edit this repository.
See DigitalOcean's [domain instructions](https://docs.digitalocean.com/products/app-platform/how-to/manage-domains/).

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

Deployments interrupt SSE connections. After reconnect, clients restore the saved position and their original seats.
Missed emoji reactions do not replay.
One instance still has temporary downtime during replacement.
Keep scaling fixed at one instance until shared event delivery is implemented.

For updates, change the image digest in the existing app's component source and deploy manually.
Retain the previous digest for rollback.
If you update through `doctl`, export the current app spec before you change its image digest.
Preserve domains, encrypted credentials, and runtime variables from that exported spec.
Do not replace an existing app with the initial template because that can remove its later configuration.

## Alternative: Railway

Railway can run the same published GHCR image without a separate proxy container.
Keep the website and API together in one service.
Private registry credentials require Railway's Pro plan. See the [private registry guide](https://docs.railway.com/guides/private-container-registry).
Public images do not require registry credentials.

To deploy the service:

1. Create a Railway service from a Docker image.
2. Enter the published image reference, such as `ghcr.io/dustywusty/tinychess:0.1.0`.
3. For a private image, configure read-only registry credentials in Railway.
4. Set the service variable `PORT=8080`.
5. Set the health check path to `/healthz`.
6. Keep exactly one replica in one region.
7. Disable Serverless and image auto updates.
8. Leave the start command empty so Railway uses the image entrypoint.
9. Generate a public domain with target port `8080`.
10. Review the estimated cost before deployment.

Replace the example tag with your tested, uniquely versioned release tag.
Keep registry credentials out of Git.
See Railway's [image deployment guide](https://docs.railway.com/services).

Railway checks the configured health endpoint before it activates a deployment.
It does not continuously poll that endpoint after activation.
See [Railway health checks](https://docs.railway.com/deployments/healthchecks).

CAUTION: Keep Serverless disabled for live streams. Without Postgres, a sleeping service can also lose its games and player seats.

After deployment, check `/healthz` and `/api/version` on the generated HTTPS domain.
Set the APK's `EXPO_PUBLIC_API_URL` to that HTTPS origin without an `/api` suffix.
Railway provides HTTPS through its [public networking](https://docs.railway.com/networking/public-networking).

For Postgres recovery, set `DATABASE_URL` through Railway service variables.
The same recovery rules and single-instance limit apply on both platforms.

## Update or roll back

CAUTION: Without Postgres, finish active games before replacement. Restarts erase in-memory games.

Before an update, back up Postgres if persistence is enabled.
Record the current image reference.
Select the new tested digest for App Platform or release tag for Railway.

For App Platform, update the existing component's image digest and deploy from the control panel.
For Railway, update the service's image reference to the tested release and deploy manually.
Check `/healthz`, `/api/version`, and a two-phone game after deployment.

For rollback, repeat the same process with the previous image reference.
Check database schema compatibility before rollback because startup migrations can change the schema.
Do not roll back active games to a version without recovery support.
Older versions do not update the saved recovery state.

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

Build the APK:

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
