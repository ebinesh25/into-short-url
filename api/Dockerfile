FROM golang:1.16-alpine3.13
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o app .
CMD ["./app"]
EXPOSE 8080
