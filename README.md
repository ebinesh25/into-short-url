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
api/
├── main.go          # Entry point, route setup
├── routes/
│   ├── shorten.go   # Shorten URL handler
│   └── resolve.go   # Redirect handler
├── helpers/
│   ├── shorten.go   # Short code generation
│   └── redis.go     # Redis operations
└── database/
    └── database.go  # Redis connection
```

## Running Locally

```bash
cd api
go run main.go
```

## Docker

```bash
docker build -t intolink api/
docker run -p 8080:8080 --env-file api/.env intolink
```

## Environment Variables

- `PORT`: Server port (default: 8080)
- `REDIS_ADDR`: Redis connection string
- `REDIS_PASSWORD`: Redis password (optional)
