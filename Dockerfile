# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/intolink ./cmd/server

FROM alpine:3.22 AS runtime
RUN apk add --no-cache ca-certificates \
    && addgroup -S intolink \
    && adduser -S -G intolink intolink

WORKDIR /app
COPY --from=build /out/intolink ./intolink
COPY web ./web

USER intolink
EXPOSE 8080
ENTRYPOINT ["/app/intolink"]
