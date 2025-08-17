# syntax=docker/dockerfile:1

# ---- Builder ----
FROM golang:1.22-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
	build-essential ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1 GOOS=linux

# Use build cache mounts for faster rebuilds
RUN --mount=type=cache,target=/root/.cache/go-build \
	--mount=type=cache,target=/go/pkg/mod \
	go build -o /out/server .

# ---- Runtime ----
FROM debian:bookworm-slim AS runner

RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates tzdata curl && rm -rf /var/lib/apt/lists/*

ENV PORT=8080 \
	ALLOW_CORS=true \
	DATABASE_URL=file:/data/app.db?_foreign_keys=on \
	JWT_SECRET=change-me-in-prod \
	DEFAULT_ADMIN_EMAIL=admin@example.com \
	DEFAULT_ADMIN_PASSWORD=admin123

WORKDIR /

COPY --from=builder /out/server /server

# Non-root user
RUN useradd -m -u 10001 appuser && \
	mkdir -p /data && chown -R appuser:appuser /data

USER appuser

VOLUME ["/data"]

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --retries=5 --start-period=10s \
	CMD curl -fsS http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/server"]