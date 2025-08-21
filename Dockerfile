# ===== build =====
FROM golang:1.24-bookworm AS builder

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
  git gcc g++ make libsqlite3-dev ca-certificates && \
  rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /out/arbitrage-sync ./cmd

FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
  chromium \
  ca-certificates \
  tzdata \
  libsqlite3-0 \
  fonts-liberation \
  libasound2 \
  libatk-bridge2.0-0 \
  libatk1.0-0 \
  libc6 \
  libcairo2 \
  libcups2 \
  libdbus-1-3 \
  libexpat1 \
  libfontconfig1 \
  libgcc1 \
  libgdk-pixbuf2.0-0 \
  libglib2.0-0 \
  libgtk-3-0 \
  libnspr4 \
  libnss3 \
  libpango-1.0-0 \
  libx11-6 \
  libxcomposite1 \
  libxdamage1 \
  libxrandr2 \
  xdg-utils \
  dumb-init && \
  rm -rf /var/lib/apt/lists/*

ENV CHROME_BIN=/usr/bin/chromium \
    TZ=Europe/Moscow

# Рекомендуемые флаги для headless в контейнере.
# Если у вас chromedp/puppeteer/playwright — просто читайте эти флаги из ENV в коде.
ENV CHROME_FLAGS="\
  --headless=new \
  --disable-gpu \
  --disable-software-rasterizer \
  --disable-dev-shm-usage \
  --no-zygote \
  --remote-debugging-pipe \
  --window-size=1920,1080 \
  --hide-scrollbars \
  --no-sandbox \
  "

# Если sandbox у вас ломается (часто в docker под root) — ДОБАВЬТЕ:
# ENV CHROME_FLAGS=\"$CHROME_FLAGS --no-sandbox\"
# (включайте осознанно, это компромисс по безопасности)

WORKDIR /app
COPY --from=builder /out/arbitrage-sync /usr/local/bin/arbitrage-sync

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/usr/local/bin/arbitrage-sync"]
