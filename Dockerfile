# syntax=docker/dockerfile:1

# ---- Builder ----
FROM golang:1.22-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux

RUN --mount=type=cache,target=/root/.cache/go-build \
	--mount=type=cache,target=/go/pkg/mod \
	go build -o /out/server .

# ---- Runtime ----
FROM gcr.io/distroless/base-debian12 AS runner

ENV PORT=8080 \
	ALLOW_CORS=true \
	DATABASE_URL=postgres://postgres:postgres@db:5432/social?sslmode=disable \
	JWT_SECRET=change-me-in-prod \
	DEFAULT_ADMIN_EMAIL=admin@example.com \
	DEFAULT_ADMIN_PASSWORD=admin123

WORKDIR /

COPY --from=builder /out/server /server

USER 65532:65532

EXPOSE 8080

# no shell available; rely on container orchestration healthchecks
ENTRYPOINT ["/server"]