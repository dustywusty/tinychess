# Your Move mobile

Expo SDK 57 / React Native client for iOS and Android.

Set `EXPO_PUBLIC_API_URL` when a device cannot reach the default simulator
address. Set `EXPO_PUBLIC_WEB_URL` to the deployed web origin used in shares.

```sh
corepack pnpm@9.15.0 install
corepack pnpm@9.15.0 --filter @yourmove/mobile start
```

Routes mirror web links: `/g/[id]`. `yourmove://g/<id>` works once a development
build is installed. Universal/app links require replacing `yourmove.example`
and hosting the platform association files.
