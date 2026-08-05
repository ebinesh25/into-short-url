# IntoLink

A simple URL shortener service built with Go and Gin framework. Append any URL to the domain to get a shortened link that redirects to the original destination.

## How It Works

1. **Shorten a URL**: Visit `intolink.site/https://long-url-site.com`
2. **Get short code**: Returns a shortened URL like `intolink.site/suwodj`
3. **Redirect**: Visiting the short URL redirects to the original long URL

## Features

- URL shortening with random 10-character alphanumeric codes
- Redis-backed storage for fast lookups
- 301 permanent redirects
- Health check endpoint (`/ping`)

## Project Structure

```
.
├── cmd/server/       # Application entry point
├── internal/         # Database, URL, and route packages
├── web/templates/    # Runtime HTML templates
├── Dockerfile        # Production multi-stage image
└── compose.yaml      # Application and Redis services
```

## Run with Docker

```bash
docker compose up --build
```

The API is available at `http://localhost:8080`; verify it with
`curl http://localhost:8080/ping`. Redis data is persisted in the
`redis-data` Docker volume. Set `PORT` before starting Compose to use a
different host port.

To stop the services, run `docker compose down`. Add `--volumes` only when
you also want to delete the persisted Redis data.

## Running without Docker

Start Redis, set `REDIS_STRING` to its URL, then run:

```bash
go run ./cmd/server
```

## Environment Variables

- `PORT`: Server port (default: 8080)
- `REDIS_STRING`: Redis connection URL (for example, `redis://localhost:6379/0`)
