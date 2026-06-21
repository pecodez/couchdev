FROM node:lts-alpine AS web
WORKDIR /src
COPY hub/web/package*.json ./
RUN npm ci
COPY hub/web/ ./
RUN npm run build

FROM golang:latest AS go-build
WORKDIR /src
COPY hub/go.mod hub/go.sum ./
RUN go mod download
COPY hub/ ./
COPY --from=web /src/dist ./web/dist
RUN go build -trimpath -o /couchdev ./cmd/couchdev

FROM ubuntu:24.04
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
  && apt-get install -y --no-install-recommends \
       ca-certificates curl gnupg tmux \
  && curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - \
  && apt-get install -y --no-install-recommends nodejs \
  && npm install -g @anthropic-ai/claude-code \
  && rm -rf /var/lib/apt/lists/*
COPY --from=go-build /couchdev /usr/local/bin/couchdev
COPY hub/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
