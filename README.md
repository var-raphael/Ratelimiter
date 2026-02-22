# Go Rate Limiter

A lightweight rate limiting service built in Go. Drop it in front of any API and it handles the throttling for you. No external dependencies, pure standard library.

Built on the **token bucket algorithm**: tokens refill continuously over time, so traffic is smoothed rather than hard-cutoff. Burst-friendly by design.

---

## How it works

Every unique key (user ID, IP, API key, whatever you choose) gets its own bucket. Each request costs one token. Tokens refill at a constant rate based on your `limit` and `window` settings. If the bucket is empty, the request is rejected until tokens refill.

```
limit=10, window=60  →  refill rate of 1 token every 6 seconds
```

Inactive buckets (no activity for 10+ minutes) are cleaned up automatically so memory doesn't grow forever.

---

## Quickstart

```bash
git clone https://github.com/var-raphael/Ratelimiter.git
cd Ratelimiter
go run main.go
# Listening on http://localhost:8080
```

To use a different port:

```bash
PORT=9000 go run main.go
```

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go build -o rate-limiter main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/rate-limiter .
EXPOSE 8080
CMD ["./rate-limiter"]
```

```bash
docker build -t rate-limiter .
docker run -p 8080:8080 rate-limiter
```

---

## API

### `POST /check`

Check whether a request should be allowed.

**Request**
```json
{
  "key": "user:123",
  "limit": 100,
  "window": 60
}
```

| Field | Type | Description |
|-------|------|-------------|
| `key` | string | Unique identifier: user ID, IP, API key, etc. |
| `limit` | int | Max requests allowed in the window |
| `window` | int | Time window in seconds |

**Allowed (`200 OK`)**
```json
{
  "allowed": true,
  "remaining": 99,
  "reset_at": 1704564789
}
```

**Rate limited (`429 Too Many Requests`)**
```json
{
  "allowed": false,
  "remaining": 0,
  "reset_at": 1704564789
}
```

`reset_at` is a Unix timestamp for when the next token becomes available.

---

### `GET /health`

```json
{
  "status": "healthy",
  "time": "2026-02-22T01:42:00Z"
}
```

---

## Usage examples

### cURL
```bash
curl -X POST http://localhost:8080/check \
  -H "Content-Type: application/json" \
  -d '{"key": "user:dave", "limit": 10, "window": 60}'
```

### Node.js
```javascript
async function checkRateLimit(key, limit = 100, window = 60) {
  const res = await fetch('http://localhost:8080/check', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ key, limit, window }),
  });

  const data = await res.json();

  if (data.allowed) {
    console.log(`Allowed. Remaining: ${data.remaining}`);
    return true;
  }

  const retryIn = data.reset_at - Math.floor(Date.now() / 1000);
  console.log(`Rate limited. Retry in ${retryIn}s`);
  return false;
}
```

### Python
```python
import requests
import time

def check_rate_limit(key, limit=100, window=60):
    res = requests.post('http://localhost:8080/check', json={
        'key': key, 'limit': limit, 'window': window
    })
    data = res.json()

    if data['allowed']:
        print(f"Allowed. Remaining: {data['remaining']}")
        return True

    retry_in = data['reset_at'] - int(time.time())
    print(f"Rate limited. Retry in {retry_in}s")
    return False
```

### PHP
```php
function checkRateLimit($key, $limit = 100, $window = 60) {
    $payload = json_encode(['key' => $key, 'limit' => $limit, 'window' => $window]);

    $ch = curl_init('http://localhost:8080/check');
    curl_setopt_array($ch, [
        CURLOPT_CUSTOMREQUEST  => 'POST',
        CURLOPT_POSTFIELDS     => $payload,
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_HTTPHEADER     => ['Content-Type: application/json'],
    ]);

    $result = json_decode(curl_exec($ch), true);
    curl_close($ch);

    if ($result['allowed']) {
        echo "Allowed. Remaining: {$result['remaining']}\n";
        return true;
    }

    $retryIn = $result['reset_at'] - time();
    echo "Rate limited. Retry in {$retryIn}s\n";
    return false;
}
```

---

## Key design patterns

The `key` field is how you define what gets rate limited. Some examples:

```
user:123                         →  per user
ip:192.168.1.1                   →  per IP address
apikey:abc123                    →  per API key
user:123:endpoint:/api/search    →  per user per endpoint
```

For tiered limits, just use different keys:

```
user:123:free      limit=100,  window=3600   (free tier)
user:123:pro       limit=5000, window=3600   (pro tier)
```

---

## HTTP status codes

| Code | Meaning |
|------|---------|
| `200` | Request allowed |
| `400` | Missing or invalid parameters |
| `405` | Wrong HTTP method |
| `429` | Rate limit exceeded |

---

## Performance

Tested on modest hardware:

- **10,000–50,000 req/s** throughput
- **Sub-millisecond** response times
- **~100 bytes** memory per active key
- Concurrent-safe via `sync.Map` and per-bucket mutexes

---

## Production notes

- **CORS**: Currently set to `*`. Lock it down to specific origins before deploying.
- **HTTPS**: Run behind nginx or Caddy with SSL. This service speaks plain HTTP.
- **Key hashing**: Don't use raw API keys as the `key` field. Hash them first.
- **DDoS**: Put this behind a load balancer or CDN. It rate limits at the app layer, not the network layer.

---

## License

MIT. Use it however you want.
