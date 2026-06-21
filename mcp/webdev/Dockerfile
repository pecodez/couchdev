FROM node:lts-alpine
RUN apk add --no-cache chromium nss freetype harfbuzz ca-certificates ttf-freefont
ENV PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
ENV PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH=/usr/bin/chromium-browser
WORKDIR /app
COPY mcp/webdev/package*.json ./
RUN npm ci
COPY mcp/webdev/ ./
RUN npm run build
CMD ["tail", "-f", "/dev/null"]
