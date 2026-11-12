# 226os web image (ADR-0006): Vite build served by an unprivileged nginx on 8080.
# Build from the repository root: docker build -f src/web/Dockerfile -t 226os-web .

# node:26-alpine (Node 26.13.1; .nvmrc says 26). Node 26 no longer bundles corepack, so a pinned
# corepack is installed, then pnpm itself is resolved from the packageManager field.
FROM node:26-alpine@sha256:6bd9ab43c0ec0190ca7b2d193906cb8b7edf6e7f4afa43ddeccd6452710bffb5 AS build
RUN npm install --global --ignore-scripts corepack@0.36.0 && corepack enable
ENV COREPACK_ENABLE_DOWNLOAD_PROMPT=0 CI=true
WORKDIR /repo
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY src/web/package.json src/web/
RUN pnpm install --frozen-lockfile --filter web...
COPY src/web src/web
COPY api api
RUN pnpm --filter web build

# nginxinc/nginx-unprivileged:stable-alpine (nginx 1.30.6, alpine 3.24.2, uid 101, no root at runtime).
# Digest re-resolved 2026-11-12: upstream had not republished (same digest), and no variant was clean
# (stable-alpine-slim and mainline-alpine-slim still had 3 fixable HIGH findings). The base therefore
# carries fixable HIGH/CRITICAL OS-package CVEs (brotli, curl, expat, libxslt, musl, zlib), fixed by the
# `apk upgrade` layer below. Trade-off: that layer pulls the then-current alpine packages at build time, so
# two builds of the same commit can differ until upstream republishes the base; the final image is
# recorded by digest at build, and image-scan.sh gates it. Drop the layer when the base is clean.
FROM nginxinc/nginx-unprivileged:stable-alpine@sha256:f5116ed4606f3eb015dd4a8e28cb9b351476dc0ad4bc6e9f93b19cec130cdf03 AS web
USER root
RUN apk upgrade --no-cache
ARG VCS_REF=unknown
LABEL org.opencontainers.image.title="226os-web" org.opencontainers.image.revision="${VCS_REF}"
COPY src/web/nginx/nginx.conf /etc/nginx/nginx.conf
COPY src/web/nginx/security-headers.conf /etc/nginx/security-headers.conf
COPY --from=build /repo/src/web/dist /usr/share/nginx/html
USER 101:101
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
  CMD ["wget", "-q", "-O", "/dev/null", "http://127.0.0.1:8080/"]
