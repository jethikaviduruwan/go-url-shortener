# go-url-shortener

A full-featured URL shortener service in Go — custom codes, expiry, click analytics, and a stats dashboard. Zero external dependencies.

## Features
- Shorten any URL to a random 7-character code
- Custom alias support (`/r/my-link`)
- Optional TTL — links auto-expire after N hours
- Click tracking with referer + user-agent logging
- Analytics endpoint: top links by clicks, total clicks, total links
- UTM/tracking parameter stripping
- CORS-enabled for frontend integration

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/shorten` | Shorten a URL |
| GET | `/api/links` | List all links |
| DELETE | `/api/links/:code` | Delete a link |
| GET | `/api/stats` | Click analytics |
| GET | `/r/:code` | Redirect to original URL |
| GET | `/health` | Health check |

## Usage

```bash
go run .
```

```bash
# Shorten a URL
curl -X POST http://localhost:9090/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/very/long/path"}'

# With custom code and 24h expiry
curl -X POST http://localhost:9090/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","custom_code":"mylink","ttl_hours":24}'

# Stats
curl http://localhost:9090/api/stats

# Follow a short link
curl -L http://localhost:9090/r/mylink
```

## Configuration

| Env var | Default | Description |
|---------|---------|-------------|
| `PORT` | `9090` | HTTP port |
| `BASE_URL` | `http://localhost:9090` | Public base URL for short links |
// v4-1
