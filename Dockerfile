# ---------- Build Stage ----------
FROM golang:1.23 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o backend main.go


# ---------- Runtime Stage ----------
FROM ubuntu:22.04

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    postgresql-client tzdata ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/backend /app/backend
COPY --from=builder /app/public /app/public

ENV TZ=Asia/Jakarta

EXPOSE 8000

CMD ["./backend"]
