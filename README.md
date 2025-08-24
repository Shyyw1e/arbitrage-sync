# arbitrage-sync

Service for monitoring crypto arbitrage opportunities between **Rapira** and **Grinex** with Telegram notifications.  
Built with Go, `chromedp`, Redis-backed job queue, and a per-user worker/dispatcher model.

> Current pairs: **Rapira USDT/RUB**, **Grinex USDT/A7A5** (+ cross-combinations).  
> Note: Grinex USDT/RUB is temporarily disabled in code (can be enabled later when order book stabilizes).

---

## Features

- **Data sources & parsing**
  - Grinex: React pages via `chromedp` (headless Chromium), retries, anti-overlays/cookies.
  - Rapira: HTML/DOM via `chromedp`.
  - Single global ExecAllocator, prewarming and `EnsureAlive` helpers.
- **Order book caching**
  - Custom `OrderCache` (key: `Source|Pair|Side`), **TTL=60s**, `isUpdating` guard to avoid concurrent scrapes for the same key.
- **Queue & workers**
  - Redis queue (`BLPOP jobs:queue`), **job** format: `detect-as:<minDiff>:<maxSum>:<chatID>`.
  - **Dispatcher** spawns **one worker per chatID** and controls lifecycle by channel commands (`start/stop/update/shutdown`).
  - **Heartbeat** after each tick and **watchdog** that soft-restarts workers stale for `>90s`.
- **User state**
  - SQLite store: `minDiff`, `maxSum`, `step` (`waiting_for_input` → `ready_to_run` → `not_active`), survives restarts.
- **Analysis logic**
  - “Factual” (immediate) and “Potential/Reverse” signals.
  - Commission-aware (e.g., Grinex A7A5 — `0.0005`, Rapira — `0.0`), sum limit, rounding, **anti-duplicate** (per-chat hash), anti-spam.
- **Telegram bot**
  - `/start`, “Start/Stop analysis”, change params.
  - Messages via `go-telegram-bot-api`.
- **Production**
  - Docker multi-stage, headless Chromium (`CHROME_FLAGS`), `docker-compose` with `redis-internal` service, mounted `.env` and `data.db`, larger `/dev/shm`, `ulimits`.
- **Resilience**
  - Global key-level mutex for critical sections, race protection.
  - Channels + custom TTL cache reduce load and latency.
  - Graceful soft restarts by watchdog.

---

## Architecture (high level)

```
Telegram UI  →  Bot handler  →  Redis queue (jobs:queue)
                                 |
                                 v
                         Dispatcher (per-chat registry)
                          └── Worker(chatID)  —[20s tick]→  Parsers (chromedp)
                                               └→ Analysis → Telegram notify
                           ^
                           └— Watchdog (15s) monitors HB (90s stale → soft restart)
```

- **Dispatcher** keeps `map[chatID]*worker`, guarantees **one worker per chat**.  
- **Worker** keeps atomics (`min`, `max`, `bot`, `hb`) + command channel.  
- **Tick**: every `20s` (when running) → fetch/calc → send signals → set heartbeat.

---

## Getting Started

### Prerequisites
- Go 1.22+
- Docker & Docker Compose (recommended for prod-like run)
- A running Redis (compose provides one)
- Telegram Bot token

### Environment

Create `.env` in project root (example):
```ini
TELEGRAM_TOKEN=123456:ABC...
REDIS_ADDR=redis-internal:6379
REDIS_PASSWORD=
REDIS_DB=0

# chromedp/headless chrome
CHROME_FLAGS=--headless=new --disable-gpu --no-sandbox --disable-dev-shm-usage
```

### Run locally (simple)
```bash
go mod download
go run ./cmd/bot
```

### Run with Docker Compose
```bash
docker compose up -d --build
# logs
docker compose logs -f bot
```

Compose should start:
- `bot` (your Go service, headless Chromium inside image)
- `redis-internal` (queue storage)

> Ensure your container has enough shared memory (`/dev/shm`) and no GPU requirement.

---

## Bot flow (user side)

1. `/start` → bot asks for parameters (`minDiff`, `maxSum`).
2. “▶️ Начать анализ” →
   - user state becomes `ready_to_run`
   - job is enqueued: `detect-as:<minDiff>:<maxSum>:<chatID>`
   - dispatcher ensures worker(chatID) and starts it
3. Worker tick (20s):
   - pulls/caches order books, calculates **factual**/**potential** opportunities
   - sends Telegram messages
   - updates heartbeat (`hb`) at the end of tick
4. Watchdog (15s): if `now - hb > 90s`, soft-restart worker.
5. “⏹ Остановить анализ” → dispatcher stops worker, user step `not_active`, jobs cleaned.

---

## Redis Queue

- **Key**: `jobs:queue`
- **Enqueue**: `RPUSH jobs:queue detect-as:0.20:300000:123456789`
- **Dequeue**: worker loop `BLPOP jobs:queue 10`

> If using Redis Cluster/Sentinel, prefer `NewFailoverClient`/`NewClusterClient` so `BLPOP` always goes to master.

---

## Worker lifecycle (dispatcher)

- `start(chatID, min, max, bot)` → ensure worker exists → send `cmdStart`
- `stop(chatID)` → send `cmdStop`
- `isRunning(chatID)` → atomic bool from worker
- `list()` → snapshot of workers map

**Worker commands** (via channel):
- `cmdStart`: set params/bot, start ticker if not running
- `cmdUpdate`: update params/bot on the fly
- `cmdStop`: stop ticker and mark not running
- `cmdShutdown`: stop + return from goroutine

---

## Troubleshooting

### 1) BLPOP `READONLY` error
`BLPOP` is treated as a **write** operation by Redis (it modifies a list). On a **replica**, Redis returns `READONLY`.  
If you see errors like `READONLY You can't write against a read only replica`:

- Ensure your client is connected to **master**.
- In failover scenarios: re-init Redis client on READONLY, add small backoff.
- For Sentinel/Cluster use `NewFailoverClient` / `NewClusterClient`.

**Quick patch idea**:
- Detect READONLY in error handler → call `resetRedisClient()` → retry after `2s`.

### 2) Watchdog “soft restart loop”
If after (re)start watchdog immediately considers worker stale:
- Ensure **heartbeat is reset on `cmdStart`** or first tick runs successfully before 90s.
- Check that worker actually receives ticks (ticker channel isn’t `nil`).
- Verify user step is `ready_to_run` in SQLite (stop sets it to `not_active`).

### 3) Headless Chromium issues
- Add `--disable-dev-shm-usage` and `--no-sandbox` in containers.
- Increase `/dev/shm` size if pages are heavy.
- Prewarm and `EnsureAlive` the browser process.

---

## Roadmap

- Turn Redis queue into **worker pool** with backpressure.
- Add proper **rate limiting** for scrapers.
- Metrics/exporter for Prometheus.
- Enable Grinex USDT/RUB when the order book stabilizes.
- Graceful shutdown hooks (`cmdShutdown`) on process exit.

---

## Tech Stack

Go, `chromedp`, `goquery`, Redis, SQLite, Telegram Bot API, Docker, `logrus`.

---

## License

Private / TBD (choose a license before publishing).

