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

FROM debian:bookworm-slim
RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates tmux \
  && rm -rf /var/lib/apt/lists/*
COPY --from=go-build /couchdev /usr/local/bin/couchdev
CMD ["couchdev", "serve", "--config", "/etc/couchdev/config.json"]
