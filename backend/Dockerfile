# syntax=docker/dockerfile:1.7
# ponytail: multi-stage dev/prod in single Dockerfile; split when CI matrices diverge
FROM golang:1.24-alpine AS base
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Target: dev (Hot reload with Air)
FROM base AS dev
RUN go install github.com/air-verse/air@v1.61.7
CMD ["air", "-c", ".air.toml"]

# Target: build (Static Go compilation)
FROM base AS build
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/server ./cmd/server

# Target: prod (Minimal Alpine runtime)
FROM alpine:3.21 AS prod
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g 10001 -S amba && \
    adduser -u 10001 -S amba -G amba

USER 10001:10001
WORKDIR /app
COPY --from=build --chown=10001:10001 /bin/server /app/server

ENV SERVER_HOST=0.0.0.0 \
    SERVER_PORT=8080 \
    ENVIRONMENT=development \
    LLM_MOCK=true

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/app/server"]
