# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN GOPROXY=https://mirror-go.runflare.com go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/unitly ./cmd

# Runtime stage
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /usr/local/bin/unitly /usr/local/bin/unitly

EXPOSE ${SERVER_PORT}
ENTRYPOINT ["/usr/local/bin/unitly"]