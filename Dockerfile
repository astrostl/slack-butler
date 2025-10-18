# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . ./

# Build static binary
# Note: VERSION is passed as build arg during release builds
# Example: docker build --build-arg VERSION=v1.2.1
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w -X main.Version=${VERSION}" -a -installsuffix cgo -o slack-butler .

# Final stage - minimal image with only the binary
FROM scratch

# Copy only the binary
COPY --from=builder /app/slack-butler /slack-butler

# Copy CA certificates for HTTPS (required for Slack API calls)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/slack-butler"]
