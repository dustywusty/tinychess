# Deploy Tinychess and build an Android APK

The container serves the Go API and the compiled website on port 8080.
The Android app connects to the same API through HTTPS.
The runtime uses a non-root user and includes no shell or package manager.
DigitalOcean App Platform is the primary deployment target. The `.do/app.yaml` file defines its service.

## Current limits

CAUTION: Run exactly one application instance. Live games, seats, and reaction streams exist only in that process.

CAUTION: Deploy between games. A restart or deployment clears active games, even when Postgres stores their history.

Postgres persistence is optional and best-effort. It does not provide active-game recovery or synchronization between replicas.
The server also removes games after 24 hours without activity.
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
For a Droplet trial, set `TINYCHESS_IMAGE=tinychess:ci` in the deployment environment file.
Do not run the registry pull command for this local artifact.

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
Publishing an image does not deploy it to DigitalOcean or AWS.

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

### Optional Postgres history

For game history, attach managed Postgres and add `DATABASE_URL` as an encrypted runtime variable on the service.
Use the database provider's TLS configuration and restrict database access to the app.
The spec omits this variable so credentials cannot enter Git accidentally.
Postgres does not restore active games or player seats after a deployment.

### Deployments and live games

CAUTION: Finish games before deployment. Platform restarts and deployments replace the process that owns the live games.

One instance does not provide continuity during a deployment.
Players must start a new game after the replacement.
Keep scaling fixed at one instance until shared live state and recovery are implemented.

For updates, change the image digest in the existing app's component source and deploy manually.
Retain the previous digest for rollback.
If you update through `doctl`, export the current app spec before you change its image digest.
Preserve domains, encrypted credentials, and runtime variables from that exported spec.
Do not replace an existing app with the initial template because that can remove its later configuration.

## Alternative: DigitalOcean Droplet

This path uses Docker Compose and Caddy on one Droplet.
Caddy provides HTTPS and forwards requests to the application.
It flushes streaming responses immediately.
See the [Caddy reverse-proxy reference](https://caddyserver.com/docs/caddyfile/directives/reverse_proxy).

1. Install Docker Engine and the Compose plugin on the Droplet.
2. Point your domain's DNS records to the Droplet.
3. Permit inbound TCP ports 80 and 443 through the firewall.
4. Restrict SSH access to your administration addresses.
5. Copy the repository's `deploy` directory to the Droplet.
6. Copy `deploy/.env.example` to `deploy/.env`.
7. Set `DOMAIN` to the hostname without a scheme or path.
8. Set `TINYCHESS_IMAGE` to the published image digest.
9. If you want game history, set `DATABASE_URL` to your managed Postgres connection string.

Keep `deploy/.env` private. It can contain database credentials.
For managed Postgres, use the provider's TLS configuration and restrict database access to the application host.
The application does not require a writable volume.
The Caddy volumes retain certificate data across container replacements.

From the repository root, start the deployment:

```sh
docker compose --env-file deploy/.env -f deploy/compose.yml pull
docker compose --env-file deploy/.env -f deploy/compose.yml up -d
docker compose --env-file deploy/.env -f deploy/compose.yml ps
```

If you copied only the deployment directory, adjust the file paths to its location.
Do not expose port 8080 directly to the internet.

Check the public endpoint:

```sh
curl --fail https://chess.example.com/healthz
curl --fail https://chess.example.com/api/version
```

Replace `chess.example.com` with your domain.
The expected health response is `{"ok":true}`.
The version response identifies the server commit.
The health endpoint checks HTTP availability, not Postgres durability.

## Alternative: ECS Fargate

The same image runs on Fargate without the Caddy container.
An Application Load Balancer (ALB) provides HTTPS.
The task accepts HTTP traffic on port 8080 from the ALB security group only.

Configure the task and service with these values:

| Setting | Value |
| --- | --- |
| Image | Published digest, or the same image copied to ECR |
| Runtime | Linux, X86_64 or ARM64 to match the image |
| Initial CPU and memory | 0.25 vCPU and 512 MiB, then measure usage |
| Network mode | `awsvpc` |
| Container port | 8080 |
| User | `65532:65532` |
| Read-only root filesystem | Enabled |
| Health command | `["CMD", "/app/tinychess", "-healthcheck"]` |
| Health interval / timeout / retries | 30 seconds / 5 seconds / 3 |
| Health start period | 10 seconds |
| Stop timeout | 15 seconds or more |
| Desired count | 1 |
| Autoscaling | Disabled |
| Deployment minimum / maximum | 0% / 100% |
| ALB target type | `ip` |
| ALB health path | `/healthz`, success code 200 |
| ALB idle timeout | 60 seconds or more |
| Logs | CloudWatch through the `awslogs` driver |

The server sends an SSE heartbeat every 15 seconds.
The 0% / 100% deployment configuration stops the old task before it starts the new task.
This causes downtime but prevents two independent live-game servers during a rolling deployment.
See the [ECS service parameters](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service_definition_parameters.html).

Store `DATABASE_URL` in Secrets Manager or Systems Manager Parameter Store, not in a committed task definition.
Grant the task execution role access to the selected secret and registry.
For private GHCR images, configure ECS repository credentials or copy the image to private ECR.
Configure outbound connectivity for image pulls and logs.
For private subnets, provide the required NAT access or service endpoints.

## Update or roll back

CAUTION: Finish active games before replacement. Neither an update nor a rollback restores in-memory games.

Before an update, back up Postgres if persistence is enabled.
Record the current image digest.
Select the new tested digest.

For App Platform, update the existing component's image digest and deploy from the control panel.
For a Droplet, change `TINYCHESS_IMAGE` in `deploy/.env`.
Then run the Compose pull and up commands again.
For ECS, register a task revision with the new digest and update the service.
Check `/healthz`, `/api/version`, and a two-phone game after deployment.

For rollback, repeat the same process with the previous digest.
Check database schema compatibility before rollback because startup migrations can change the schema.
Do not remove the Caddy volumes during an application update.

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
npx eas-cli@latest env:set --name EXPO_PUBLIC_API_URL --value https://chess.example.com --environment preview --visibility plaintext
```

Replace the example domain with your live domain.
Do not append `/api` because the client adds endpoint paths.
The pre-install check rejects a missing origin, HTTP, localhost, credentials, and URL paths.
`EXPO_PUBLIC_*` values are visible in the app bundle. They must not contain secrets.
See [Expo's environment-variable instructions](https://docs.expo.dev/eas/environment-variables/manage/).

For app-to-app testing, leave `EXPO_PUBLIC_WEB_URL` unset.
The share action then uses `yourmove://g/<id>` links for the installed app.
The friend can also paste a game ID or game link into the invite field.
HTTPS app links still use `yourmove.example` placeholders in `app.json` and are not configured for your domain.
Setting `EXPO_PUBLIC_WEB_URL` alone does not enable verified Android app links.

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
