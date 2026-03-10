# ── Build stage ────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o emp-api ./cmd/main.go

# ── Runtime stage ───────────────────────────────────────────
FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/emp-api .
# Copy .env if present (optional; prefer real env vars in production)
COPY .env* ./

EXPOSE 8080
ENTRYPOINT ["./emp-api"]
