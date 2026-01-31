# Use the official Golang image
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy files from the api directory
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary
RUN go build -o server main.go

# Use a tiny image for the actual runner
FROM alpine:3.18
WORKDIR /root/
COPY --from=builder /app/server .

# Cloud Run expects the app to listen on the PORT env var
EXPOSE 8080

CMD ["./server"]