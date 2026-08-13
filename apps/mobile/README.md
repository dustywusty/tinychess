# Your Move mobile

Expo SDK 56 / React Native client for iOS and Android. SDK 56 intentionally
tracks the Expo Go runtime currently distributed through the app stores; move
to SDK 57 with a development build once that becomes the default workflow.

In development the app derives the Go API host from Metro's LAN address, so a
phone opened from the QR code reaches port 8080 on the development machine.
Set `EXPO_PUBLIC_API_URL` to override it. Set `EXPO_PUBLIC_WEB_URL` to the
deployed web origin used in shares.

```sh
corepack pnpm@9.15.0 install
corepack pnpm@9.15.0 --filter @yourmove/mobile start
```

Routes mirror web links: `/g/[id]`. `yourmove://g/<id>` works once a development
build is installed. Universal/app links require replacing `yourmove.example`
and hosting the platform association files.
