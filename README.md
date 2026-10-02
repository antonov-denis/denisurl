# denisurl

A minimal URL shortener, served at **https://go301.link**.

## Purpose

Paste a long URL, get a short one. A short link is a 7-character code (`https://go301.link/<code>`) that answers with an HTTP **301 Moved Permanently** to the original URL. The service has no accounts, analytics, or link editing. It stores the code-to-URL mapping and redirects.

Links are created from a small web page (`/`) or via a JSON API (`POST /api/links`). Link creation is rate-limited per client IP.

## Software

### Stack

| Item | Value |
|---|---|
| Language | Go `1.26.7` (`go.mod`), module `github.com/antonov-denis/denisurl` |
| HTTP | Standard library `net/http` (`ServeMux` with method/path patterns) |
| Database driver | [`github.com/jackc/pgx/v5`](https://github.com/jackc/pgx) (`pgxpool`), the only direct dependency |
| Database | PostgreSQL 17 |
| Frontend | One `html/template` page, plain CSS and JS, self-hosted Geist fonts, all embedded in the binary with `go:embed` |

### Layout

```
main.go                     wiring: store -> migrations -> limiter -> web server on :8000
internal/store/             Postgres access (pgxpool) and embedded SQL migrations
  migrations/0001_init.sql  `links` table
internal/limiter/           in-memory per-IP sliding-window rate limiter
internal/web/               routes, handlers, middleware, URL validation, templates, static assets
  views/index.html          landing page / create form / API usage example
  static/                   style.css, app.js, favicon.svg, og.png, fonts/
assets/og.svg               source artwork for the Open Graph image (static/og.png)
k8s/denisurl.yaml           Kubernetes manifests
.github/workflows/          CI (ci.yml) and build+deploy (deploy.yml)
```

| Package | Responsibility |
|---|---|
| `internal/store` | Opens a `pgxpool` from `DATABASE_URL`. `GetTarget(code)` returns the target or `ErrNotFound`. `CreateTarget(code, target)` inserts a row. `Migrate()` runs every embedded `migrations/*.sql` file in order on every startup, so migrations must be idempotent (`CREATE TABLE IF NOT EXISTS`). |
| `internal/limiter` | Per-IP request timestamps held in memory: at most **10 per minute** and **50 per hour**. A background goroutine evicts idle IPs every 30 minutes. State is per process and is lost on restart. |
| `internal/web` | HTTP server: routing, handlers, the rate-limit middleware, target URL normalization, template rendering, and embedded static files. |

### HTTP routes

| Method and path | Handler | Behaviour |
|---|---|---|
| `GET /` | `handleIndex` | Renders `index.html` with `BaseURL`. |
| `GET /static/...` | file server | Embedded static assets. |
| `GET /health` | `handleHealth` | `200 Ok!` (used by the k8s probes). |
| `GET /{code}` | `handleRedirect` | `301` to the stored target, `404` if unknown, `500` on DB error. |
| `POST /api/links` | `handleCreate` (rate-limited) | Creates a link. See below. |

Creating a link:

```sh
curl -X POST https://go301.link/api/links \
  -H 'Content-Type: application/json' \
  -d '{"target": "example.com/some/long/path"}'
# 201 Created
# {"short_url":"https://go301.link/ABCDEFG"}
```

- If the target has no `://`, `https://` is prepended. Only `http` and `https` URLs with a host are accepted. Anything else gets `400`. This blocks `javascript:` and similar schemes.
- The code is the first 7 characters of `crypto/rand.Text()`, which is uppercase base32. A collision is not retried. The primary-key violation surfaces as a `500`.
- Over the rate limit: `429 too many requests`. The client IP comes from the `X-Real-IP` header when present, otherwise from the connection's remote address.

### Configuration

The application reads only environment variables. It has no flags.

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | yes | Postgres connection string passed to `pgxpool.New` |
| `BASE_URL` | yes (in practice) | Public origin with no trailing slash, e.g. `https://go301.link`. Used to build `short_url` and in page meta tags. Defaults to empty. |

These are fixed in code: listen address `:8000`, read/write/idle timeouts of 60s, rate limits of 10/min and 50/hour.

### Data storage

A single table, created at startup:

```sql
CREATE TABLE IF NOT EXISTS links (
    code       TEXT PRIMARY KEY,
    target_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Running locally

1. Start Postgres with docker-compose. It runs `postgres:17` with user, password and db all set to `denisurl`, publishes port `5432`, and keeps data in the named volume `db_data`:
   ```sh
   docker compose up -d db
   ```
   Note: `docker-compose.yaml` is listed in `.gitignore`, so a fresh clone may not include it.
2. Create a `.env` file (gitignored). The Makefile `-include`s and exports it:
   ```sh
   DATABASE_URL=postgres://denisurl:denisurl@localhost:5432/denisurl
   BASE_URL=http://localhost:8000
   ```
3. Run:
   ```sh
   make run        # = go run .
   ```
   Then open http://localhost:8000.

`run` is the only Makefile target. Building a binary by hand with `go build -o main .` gives you the `main` file at the repo root, which is gitignored.

### Checks / tests

The repository has no `_test.go` files yet. CI (`.github/workflows/ci.yml`, on pushes to `main` and on PRs) runs:

```sh
gofmt -l .          # must print nothing
go vet ./...
go build ./...
go test -race ./...
```

## Infrastructure

### Container image (`Dockerfile`)

| Stage | Base image | What it does |
|---|---|---|
| `build` | `golang:1.26-alpine` | `go mod download`, then a static build: `CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/denisurl .` |
| final | `gcr.io/distroless/static-debian12:nonroot` | Copies `/denisurl`, runs as `nonroot:nonroot`, `EXPOSE 8000`, `ENTRYPOINT ["/denisurl"]` |

Templates, static files and migrations are embedded, so the image contains only the binary. `.dockerignore` excludes `.git`, `.env`, `main`, `docker-compose.yaml`, `Makefile` and `README.md`.

```sh
docker build -t denisurl .
docker run -p 8000:8000 -e DATABASE_URL=... -e BASE_URL=http://localhost:8000 denisurl
```

### docker-compose

The compose file runs only the database. The app runs on the host with `make run`.

| Service | Image | Port | Volume |
|---|---|---|---|
| `db` | `postgres:17` | `5432:5432` | `db_data:/var/lib/postgresql/data` |

### Kubernetes (`k8s/denisurl.yaml`)

The target is a single-node k3s cluster. It uses the `local-path` storage class and Traefik ingress, the k3s defaults, and is reached over SSH from CI. Every resource lives in namespace `default`.

| Kind | Name | Details |
|---|---|---|
| PersistentVolumeClaim | `denisurl-db` | 2Gi, `ReadWriteOnce`, storageClass `local-path` |
| Deployment | `denisurl-db` | 1 replica, `Recreate` strategy (RWO volume). `postgres:17` with user/db `denisurl`, password from secret. `PGDATA=/var/lib/postgresql/data/pgdata`. Readiness via `pg_isready`. Requests 128Mi/50m, limit 256Mi. |
| Service | `denisurl-db` | ClusterIP, port 5432 |
| Deployment | `denisurl` | 1 replica, image `ghcr.io/antonov-denis/denisurl:latest`, pulled with `imagePullSecrets: ghcr-secret`. Env: `BASE_URL=https://go301.link`, `DATABASE_URL` from secret, `GOMEMLIMIT=100MiB`. Liveness and readiness probes on `GET /health:8000`. Requests 32Mi/10m, limits 128Mi/100m. |
| Service | `denisurl` | ClusterIP, port 80 -> 8000 |
| Ingress | `denisurl` | Host `go301.link`, path `/`, TLS via Traefik annotations with cert resolver `letsencrypt` |

The manifest does not create these secrets. Create them by hand before applying:

| Secret | Keys | Used by |
|---|---|---|
| `denisurl-db` | `POSTGRES_PASSWORD`, `DATABASE_URL` | Postgres deployment, app deployment |
| `ghcr-secret` | docker registry credentials | pulling the app image from GHCR |

The `DATABASE_URL` in the secret should point at the in-cluster service `denisurl-db:5432`. The header comment in the manifest has the `kubectl create secret` command. The app creates the schema on startup, so there is no manual migration step.

### Deployment

`.github/workflows/deploy.yml` runs on every push to `main`:

1. Logs in to GHCR with `GITHUB_TOKEN`, builds the Dockerfile and pushes `ghcr.io/antonov-denis/denisurl:<sha>` and `:latest`.
2. SSHes to the server using the repository secrets `HETZNER_HOST`, `HETZNER_USER` and `HETZNER_SSH_KEY`, then runs:
   ```sh
   sudo kubectl set image deployment/denisurl denisurl=ghcr.io/antonov-denis/denisurl:<sha>
   sudo kubectl rollout status deployment/denisurl
   ```

CI does not apply the manifest itself. The first setup, and any later manifest change, is a manual step:

```sh
sudo kubectl apply -f k8s/denisurl.yaml
```
