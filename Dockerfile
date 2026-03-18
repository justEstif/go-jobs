# syntax=docker/dockerfile:1

# ─────────────────────────────────────────────
# Stage 1: Build
# ─────────────────────────────────────────────
FROM golang:1.26-bookworm AS builder

WORKDIR /app

# Install templ and migrate
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1001
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Download the standalone Tailwind binary (x86-64 Linux) with checksum verification
ADD --checksum=sha256:39e8d4e24b3c83b0a6e69e100a972fbc75d5fef8dce47b3ddac3cf92dea81fe3 \
    https://github.com/tailwindlabs/tailwindcss/releases/download/v4.2.1/tailwindcss-linux-x64 \
    /usr/local/bin/tailwindcss
RUN chmod +x /usr/local/bin/tailwindcss

# Download Go dependencies (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build: tailwind → templ → go
RUN tailwindcss -i styles/input.css -o static/css/tailwind.css --minify
RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/jobs ./cmd/jobs/

# ─────────────────────────────────────────────
# Stage 2: Run
# ─────────────────────────────────────────────
FROM debian:bookworm-slim AS runner

WORKDIR /app

# ca-certificates from the builder — needed for outbound HTTPS (scrapers, LLM calls)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary, migrate CLI, static assets, and migrations
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/bin/jobs ./jobs
COPY --from=builder /app/static ./static
COPY --from=builder /app/migrations ./migrations

EXPOSE 3000

ENTRYPOINT ["./jobs"]
