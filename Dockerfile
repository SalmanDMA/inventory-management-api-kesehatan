# ---------- Build Stage ----------
FROM golang:1.23 AS builder

WORKDIR /app

# Copy mod files and download dependencies first (better cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy all source
COPY . .

# pastikan folder public ikut ke builder
COPY public ./public

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o backend main.go


# ---------- Runtime Stage ----------
FROM ubuntu:22.04

# Install needed packages: tzdata for timezone, postgresql-client for pg_dump
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    postgresql-client tzdata ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary
COPY --from=builder /app/backend /app/backend

# Copy static/public folder
COPY --from=builder /app/public /app/public

EXPOSE 8000
CMD ["./backend"]
