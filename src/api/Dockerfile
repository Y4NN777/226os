# 226os API image (ADR-0006). Build from the repository root:
#   docker build -f src/api/Dockerfile --target api     -t 226os-api .
#   docker build -f src/api/Dockerfile --target migrate -t 226os-migrate .
# Every base image is pinned by digest; the tag is kept in the comment above it.

# golang:1.27.4-alpine (Go toolchain 1.27.4, matching go.mod)
FROM golang:1.27.4-alpine@sha256:0a5f7fe796599ff487605987432e7619ae45beeb2fba4c2ce0b907879932aebf AS build
ENV CGO_ENABLED=0 GOFLAGS=-mod=readonly GOTOOLCHAIN=local
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY src/api ./src/api
RUN go build -trimpath -ldflags="-s -w -buildid=" -o /out/api ./src/api/cmd/api
# Forward-only migration runner: goose pinned by go.mod, Postgres driver only.
RUN go build -trimpath -ldflags="-s -w -buildid=" \
      -tags 'no_sqlite3 no_mysql no_mssql no_clickhouse no_vertica no_ydb no_libsql no_turso' \
      -o /out/goose github.com/pressly/goose/v3/cmd/goose

# busybox:stable-musl (static busybox: only its wget applet is used, for health probes,
# because the distroless runtime has no shell and the API has no health subcommand)
FROM busybox:stable-musl@sha256:3c6ae8008e2c2eedd141725c30b20d9c36b026eb796688f88205845ef17aa213 AS busybox

# gcr.io/distroless/static-debian13:nonroot (no shell, no package manager, uid 65532)
FROM gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329 AS migrate
ARG VCS_REF=unknown
LABEL org.opencontainers.image.title="226os-migrate" org.opencontainers.image.revision="${VCS_REF}"
COPY --from=build /out/goose /goose
COPY db/migrations /migrations
# GOOSE_DRIVER / GOOSE_DBSTRING come from the environment (never from arguments, so the
# password is not visible in the process list). Only "up" is wired: forward-only.
ENV GOOSE_DRIVER=postgres GOOSE_MIGRATION_DIR=/migrations
USER 65532:65532
ENTRYPOINT ["/goose"]
CMD ["up"]

FROM gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329 AS api
ARG VCS_REF=unknown
LABEL org.opencontainers.image.title="226os-api" org.opencontainers.image.revision="${VCS_REF}"
COPY --from=build /out/api /api
COPY --from=busybox /bin/busybox /busybox/busybox
USER 65532:65532
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/busybox/busybox", "wget", "-q", "-O", "/dev/null", "http://127.0.0.1:8080/healthz"]
ENTRYPOINT ["/api"]
