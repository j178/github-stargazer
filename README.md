# GitHub Stargazer

Install [this app](https://github.com/apps/stars-notifier) to your GitHub account or an organization, and you will receive notifications when someone starred your project.

Click https://github.com/apps/stars-notifier/installations/new/ to install.

The configuration UI is now available locally and can be developed with Bun + Vite.

## Frontend Development

Install dependencies with Bun:

```sh
bun install
```

Run the frontend dev server:

```sh
bun run dev
```

Run the local Go backend in another terminal:

```sh
go run ./cmd/server
```

The Vite dev server proxies `/api/*` requests to `http://localhost:8080` by default.

Useful frontend commands:

```sh
bun run lint
bun run format
bun run build
```

## Want to run your own instance?

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https%3A%2F%2Fgithub.com%2Fj178%2Fgithub-stargazer&project-name=github-stargazer&repository-name=github-stargazer)

## TODO

- [x] personal account installation
- [x] org installation
- [x] telegram bot connect
- [x] test notification
- [ ] frontend config ui
- [x] custom HTTP request notification
- [ ] message template config
- [ ] detailed log
- [x] discord bot
- [ ] slack bot
