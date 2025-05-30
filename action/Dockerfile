# ── builder ─────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder
WORKDIR /app

# copy your Go module files from the root context
COPY go.mod go.sum ./
RUN go mod download

# copy and build your CLI
COPY scanner-cli/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /keysweep-scanner .

# ── gitleaks image ─────────────────────────────────────────────────
FROM ghcr.io/gitleaks/gitleaks:v8.23.0 AS gitleaks

# ── final image ────────────────────────────────────────────────────
FROM alpine:latest
RUN apk add --no-cache ca-certificates git curl

# make /workspace the place entrypoint will run from
WORKDIR /workspace

# bring in your baked-in rules file
COPY gitleaks.toml /workspace/gitleaks.toml

# pull in the scanner + gitleaks binaries
COPY --from=builder /keysweep-scanner /bin/keysweep-scanner
COPY --from=gitleaks /usr/bin/gitleaks /bin/gitleaks

# bring in your wrapper
COPY action/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
