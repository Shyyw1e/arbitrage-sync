FROM golang:1.24-bookworm AS builder

RUN apt-get update && apt-get install -y \
  git \
  gcc \
  g++ \
  make \
  libsqlite3-dev \
  ca-certificates

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=1 go build -o arbitrage-sync ./cmd

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
  chromium \
  ca-certificates \
  tzdata \
  libsqlite3-0 \
  fonts-liberation \
  libappindicator3-1 \
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
  && apt-get clean

WORKDIR /app
COPY --from=builder /app/arbitrage-sync .
COPY .env .
COPY data.db .

CMD ["./arbitrage-sync"]
