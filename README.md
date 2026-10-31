# Crewapparat

![crewapparat logo](doc/crewapparat.svg)

> **Part of the Apparat Suite** ⚙️
> This tool is part of a decoupled set of single-purpose event management tools built for demoparties.
> [➔ Read more about the full Apparat suite here](https://github.com/potibm/apparat)

_Crewapparat_ is the central service for managing the crew (organizers) of a demoparty: master
data, per-event participation, onboarding of new organizers, annual check-in and a contact book
during the event. Identity and access is delegated to an external OIDC provider.

> **Status: Grundgerüst.** This repository currently contains the complete platform
> (backend, frontend, infrastructure, CI/release) but **no business logic yet**. The data model,
> workflows and API endpoints are still to be implemented — see [CONCEPT.md](CONCEPT.md) for the
> intended feature set.

## Tooling

- **Backend**
  - [Go](https://go.dev)
  - [Gin Web Framework](https://gin-gonic.com)
  - [GORM](https://gorm.io) + [SQLite](https://www.sqlite.org) (pure-Go driver, no cgo)
  - [Cobra](https://cobra.dev) & [Viper](https://github.com/spf13/viper)
  - [GORM audit callbacks](https://gorm.io/docs/hooks.html) for `created_by` / `modified_by`
- **Frontend**
  - [React](https://react.dev)
  - [Vite](https://vitejs.dev/)
  - [React Admin](https://marmelab.com/react-admin/)
  - [Tailwind CSS v4](https://tailwindcss.com) + [Flowbite React](https://flowbite-react.com)
  - [Zod](https://zod.dev) for runtime config validation
- **Identity & Local Infrastructure**
  - [Dex](https://dexidp.io/) (OIDC provider for local development; production uses Rauthy)
  - [Traefik](https://traefik.io/) (local edge router)
  - [mkcert](https://github.com/FiloSottile/mkcert) & dnsmasq (local TLS & `.test` routing)
  - [Redis](https://redis.io) (event sync)
- **Observability**
  - [Sentry](https://sentry.io)
  - [OpenTelemetry](https://opentelemetry.io) + [OpenObserve](https://openobserve.ai/)
- **Development & Ops**
  - [mise](https://mise.jdx.dev/)
  - [Docker](https://www.docker.com) with Cosign signing and SBOM attestation

## Quickstart

We use `mise` to automatically manage all tool versions (Go, Node, etc.) and project tasks.

```bash
# 1. Install mise (if not already installed)
curl https://mise.run | sh

# 2. Install mise tools + project dependencies
mise install
mise run deps:install

# 3. Prepare local infrastructure
# Generates local certificates, configures dnsmasq/resolver, and updates /etc/hosts (Linux)
mise run infra:prepare

# 4. Start local services (Traefik, Dex, OpenObserve, Redis)
mise run infra:up

# 5. Start the development server (hot-reload for backend & frontend)
overmind s --timeout 10
```

First run only: `mise run be:setup` creates `backend/data/` and the dummy `backend/cmd/assets/`
that `//go:embed assets` requires to compile.

## Local Environment

Once the stack is running, you can access the applications via:

| Service           | URL                              | Credentials                               |
| ----------------- | -------------------------------- | ----------------------------------------- |
| Crewapparat Admin | https://crewapparat.test         | via Dex: `admin@example.com` / `password` |
| OpenObserve       | https://observe.crewapparat.test | `admin@example.com` / `password123`       |
| RedisInsight      | https://redis.crewapparat.test   | —                                         |
| Dex               | https://dex.crewapparat.test     | —                                         |

Development ports: frontend `3400`, backend `3401`, OTel gRPC `3417` / HTTP `3418`,
Redis `3451`.

## Everyday Commands

```bash
mise run be:dev      # backend with hot reload (Air)
mise run fe:dev      # frontend with hot reload (Vite)
mise run test        # all tests with coverage
mise run lint        # Go, TypeScript, Prettier, shell, workflows
mise run lint --fix  # same, auto-fixing
mise run build       # full frontend + backend build into ./dist
mise run docker:build # container image
```

## Documentation

- [concept & planned data model](CONCEPT.md)
- [configuration](doc/configuration.md)
- [authentication](doc/auth.md)
- [setting up a global traefik service](doc/global-traefik.md)
