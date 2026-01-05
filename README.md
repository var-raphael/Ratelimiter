# 🚀 Token Bucket Rate Limiter Service

A high-performance, lightweight rate limiting service built in Go using the token bucket algorithm. Perfect for API rate limiting, request throttling, and traffic control across distributed systems.

## ✨ Features

- **Token Bucket Algorithm**: Smooth, predictable rate limiting with burst capacity
- **High Performance**: Handles 10,000+ requests/second on modest hardware
- **Zero Dependencies**: Pure Go implementation with standard library only
- **Automatic Cleanup**: Memory-efficient with automatic bucket expiration
- **CORS Enabled**: Ready for browser-based applications
- **RESTful API**: Simple JSON-based interface
- **Thread-Safe**: Concurrent request handling with sync.Map

## 📊 Performance

- **Throughput**: 10,000-50,000 requests/second (depending on hardware)
- **Latency**: Sub-millisecond response times
- **Memory**: ~1-2 MB baseline + ~100 bytes per active rate limit key
- **Concurrency**: Handles thousands of concurrent connections

## 🎯 How It Works

The service uses a **token bucket algorithm**:
1. Each unique key gets its own bucket with a specified capacity (limit)
2. Tokens refill at a constant rate over the time window
3. Each request consumes 1 token
4. Requests are allowed if tokens are available, rejected otherwise
5. Inactive buckets are automatically cleaned up after 10 minutes

## 🚀 Quick Start

### Running Locally

```bash
# Clone or download the service
git clone https://github.com/var-raphael/Ratelimiter.git
cd rate-limiter

# Run the service
go run main.go

# The service starts on http://localhost:8080
```

### Using Docker

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

## 📡 API Reference

### Check Rate Limit

**Endpoint**: `POST /check`

**Request Body**:
```json
{
  "key": "user:123",
  "limit": 100,
  "window": 60
}
```

**Parameters**:
- `key` (string, required): Unique identifier for the rate limit (e.g., user ID, IP address, API key)
- `limit` (int, required): Maximum number of requests allowed
- `window` (int, required): Time window in seconds

**Response** (200 OK):
```json
{
  "allowed": true,
  "remaining": 99,
  "reset_at": 1704564789
}
```

**Response** (429 Too Many Requests):
```json
{
  "allowed": false,
  "remaining": 0,
  "reset_at": 1704564789
}
```

### Health Check

**Endpoint**: `GET /health`

**Response**:
```json
{
  "status": "healthy",
  "time": "2024-01-06T15:04:05Z"
}
```

## 💻 Usage Examples

### Python

```python
import requests
import time

def check_rate_limit(key, limit=100, window=60):
    url = "http://localhost:8080/check"
    payload = {
        "key": key,
        "limit": limit,
        "window": window
    }
    
    response = requests.post(url, json=payload)
    data = response.json()
    
    if data["allowed"]:
        print(f"Request allowed. Remaining: {data['remaining']}")
        return True
    else:
        wait_time = data["reset_at"] - int(time.time())
        print(f"Rate limited. Retry in {wait_time} seconds")
        return False

# Example usage
for i in range(5):
    check_rate_limit("user:alice", limit=3, window=10)
    time.sleep(1)
```

### Node.js

```javascript
const axios = require('axios');

async function checkRateLimit(key, limit = 100, window = 60) {
  try {
    const response = await axios.post('http://localhost:8080/check', {
      key: key,
      limit: limit,
      window: window
    });
    
    const data = response.data;
    
    if (data.allowed) {
      console.log(`Request allowed. Remaining: ${data.remaining}`);
      return true;
    }
  } catch (error) {
    if (error.response && error.response.status === 429) {
      const data = error.response.data;
      const waitTime = data.reset_at - Math.floor(Date.now() / 1000);
      console.log(`Rate limited. Retry in ${waitTime} seconds`);
    }
    return false;
  }
}

// Example usage
(async () => {
  for (let i = 0; i < 5; i++) {
    await checkRateLimit('user:bob', 3, 10);
    await new Promise(resolve => setTimeout(resolve, 1000));
  }
})();
```

### PHP

```php
<?php

function checkRateLimit($key, $limit = 100, $window = 60) {
    $url = "http://localhost:8080/check";
    $data = json_encode([
        "key" => $key,
        "limit" => $limit,
        "window" => $window
    ]);
    
    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_CUSTOMREQUEST, "POST");
    curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        'Content-Type: application/json',
        'Content-Length: ' . strlen($data)
    ]);
    
    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);
    
    $result = json_decode($response, true);
    
    if ($result['allowed']) {
        echo "Request allowed. Remaining: {$result['remaining']}\n";
        return true;
    } else {
        $waitTime = $result['reset_at'] - time();
        echo "Rate limited. Retry in {$waitTime} seconds\n";
        return false;
    }
}

// Example usage
for ($i = 0; $i < 5; $i++) {
    checkRateLimit("user:charlie", 3, 10);
    sleep(1);
}
```

### cURL

```bash
# Check rate limit
curl -X POST http://localhost:8080/check \
  -H "Content-Type: application/json" \
  -d '{
    "key": "user:dave",
    "limit": 10,
    "window": 60
  }'

# Health check
curl http://localhost:8080/health
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type RateLimitRequest struct {
    Key    string `json:"key"`
    Limit  int    `json:"limit"`
    Window int    `json:"window"`
}

type RateLimitResponse struct {
    Allowed   bool   `json:"allowed"`
    Remaining int    `json:"remaining"`
    ResetAt   int64  `json:"reset_at"`
}

func checkRateLimit(key string, limit, window int) (bool, error) {
    url := "http://localhost:8080/check"
    
    payload, _ := json.Marshal(RateLimitRequest{
        Key:    key,
        Limit:  limit,
        Window: window,
    })
    
    resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
    
    var result RateLimitResponse
    json.NewDecoder(resp.Body).Decode(&result)
    
    if result.Allowed {
        fmt.Printf("Request allowed. Remaining: %d\n", result.Remaining)
        return true, nil
    } else {
        waitTime := result.ResetAt - time.Now().Unix()
        fmt.Printf("Rate limited. Retry in %d seconds\n", waitTime)
        return false, nil
    }
}

func main() {
    for i := 0; i < 5; i++ {
        checkRateLimit("user:eve", 3, 10)
        time.Sleep(1 * time.Second)
    }
}
```

## 🎨 Common Use Cases

### API Key Rate Limiting
```json
{
  "key": "apikey:abc123",
  "limit": 1000,
  "window": 3600
}
```

### IP-Based Throttling
```json
{
  "key": "ip:192.168.1.1",
  "limit": 100,
  "window": 60
}
```

### User-Specific Limits
```json
{
  "key": "user:12345:endpoint:/api/search",
  "limit": 10,
  "window": 60
}
```

### Tiered Rate Limits
```python
# Free tier: 100 req/hour
check_rate_limit("user:123:free", limit=100, window=3600)

# Premium tier: 10000 req/hour
check_rate_limit("user:456:premium", limit=10000, window=3600)
```

## ⚙️ Configuration

### Changing the Port

Edit `main.go`:
```go
port := ":8080"  // Change to your preferred port
```

Or use environment variables:
```go
port := os.Getenv("PORT")
if port == "" {
    port = ":8080"
}
```

### Adjusting Cleanup Interval

Modify the cleanup ticker in `cleanup()`:
```go
ticker := time.NewTicker(5 * time.Minute)  // Adjust as needed
```

### Bucket Expiration Time

Change the expiration threshold in `cleanup()`:
```go
if elapsed > 10*time.Minute {  // Adjust as needed
    rl.buckets.Delete(key)
}
```

## 🔒 Security Considerations

- **Key Design**: Use unpredictable keys to prevent abuse (e.g., `hash(apikey)` instead of raw API keys)
- **CORS**: Update CORS settings in production to whitelist specific origins
- **HTTPS**: Deploy behind a reverse proxy (nginx, Caddy) with SSL/TLS
- **Authentication**: Add API key validation before rate limit checks
- **DDoS Protection**: Use this service behind a load balancer or CDN

## 🐛 Error Handling

The service returns appropriate HTTP status codes:
- `200 OK`: Request allowed
- `400 Bad Request`: Invalid parameters
- `405 Method Not Allowed`: Wrong HTTP method
- `429 Too Many Requests`: Rate limit exceeded

## 📈 Monitoring

Monitor your rate limiter with:

```bash
# Check health
watch -n 1 curl -s http://localhost:8080/health

# Load testing with Apache Bench
ab -n 10000 -c 100 -p request.json -T application/json http://localhost:8080/check
```

## 🤝 Contributing

Contributions are welcome! Feel free to submit issues or pull requests.

## 📄 License

MIT License - feel free to use in your projects!

---

