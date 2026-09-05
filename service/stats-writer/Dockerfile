FROM golang:1.25.3-bookworm AS builder

WORKDIR /app

# Copy dependency manifests first (cache friendly)
COPY shared/go.mod shared/go.sum ./shared/
COPY pb/go.mod pb/go.sum ./pb/
COPY pkg/go.mod pkg/go.sum ./pkg/
# Copy all service go.mod files because they are inter-dependent via replacements
COPY service/ ./service/
RUN find service -type f ! -name "go.mod" ! -name "go.sum" -delete

WORKDIR /app/service/stats-writer

RUN --mount=type=cache,target=/go/pkg/mod,id=gomod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild,sharing=locked \
    go mod download

WORKDIR /app

# Copy source
COPY shared/ ./shared/
COPY pb/ ./pb/
COPY pkg/ ./pkg/
COPY service/ ./service/

WORKDIR /app/service/stats-writer

RUN --mount=type=cache,target=/go/pkg/mod,id=gomod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild,sharing=shared \
    GOMAXPROCS=2 CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/stats-writer_bin ./cmd/main.go


FROM alpine:3.21

# Copy ca-certificates from builder (avoids apk network dependency)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

RUN addgroup -g 1000 appuser && \
    adduser -D -s /bin/sh -u 1000 -G appuser appuser

WORKDIR /app

RUN mkdir -p /var/log/app && chown -R 1000:1000 /var/log/app

COPY --from=builder /app/stats-writer_bin ./stats-writer

RUN chown 1000:1000 ./stats-writer

USER appuser

EXPOSE 8082

CMD ["./stats-writer"]
