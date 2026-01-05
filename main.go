package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	buckets sync.Map
	mu      sync.Mutex
}

type Bucket struct {
	tokens    float64
	capacity  float64
	refillRate float64
	lastRefill time.Time
	mu        sync.Mutex
}

type CheckRequest struct {
	Key    string `json:"key"`
	Limit  int    `json:"limit"`
	Window int    `json:"window"`
}

type CheckResponse struct {
	Allowed   bool   `json:"allowed"`
	Remaining int    `json:"remaining"`
	ResetAt   int64  `json:"reset_at"`
	Message   string `json:"message,omitempty"`
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) getBucket(key string, limit int, window int) *Bucket {
	if val, ok := rl.buckets.Load(key); ok {
		return val.(*Bucket)
	}

	refillRate := float64(limit) / float64(window)
	bucket := &Bucket{
		tokens:     float64(limit),
		capacity:   float64(limit),
		refillRate: refillRate,
		lastRefill: time.Now(),
	}

	rl.buckets.Store(key, bucket)
	return bucket
}

func (b *Bucket) allow() (bool, int, int64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastRefill = now

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		remaining := int(b.tokens)
		resetAt := now.Add(time.Duration((b.capacity - b.tokens) / b.refillRate * float64(time.Second))).Unix()
		return true, remaining, resetAt
	}

	resetAt := now.Add(time.Duration((1.0 - b.tokens) / b.refillRate * float64(time.Second))).Unix()
	return false, 0, resetAt
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.buckets.Range(func(key, value interface{}) bool {
			bucket := value.(*Bucket)
			bucket.mu.Lock()
			elapsed := time.Since(bucket.lastRefill)
			bucket.mu.Unlock()

			if elapsed > 10*time.Minute {
				rl.buckets.Delete(key)
			}
			return true
		})
	}
}

func (rl *RateLimiter) handleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CheckResponse{
			Allowed: false,
			Message: "Invalid request body",
		})
		return
	}

	if req.Key == "" || req.Limit <= 0 || req.Window <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CheckResponse{
			Allowed: false,
			Message: "Invalid parameters: key, limit and window are required",
		})
		return
	}

	bucket := rl.getBucket(req.Key, req.Limit, req.Window)
	allowed, remaining, resetAt := bucket.allow()

	w.Header().Set("Content-Type", "application/json")
	
	if !allowed {
		w.WriteHeader(http.StatusTooManyRequests)
	}

	json.NewEncoder(w).Encode(CheckResponse{
		Allowed:   allowed,
		Remaining: remaining,
		ResetAt:   resetAt,
	})
}

func (rl *RateLimiter) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	rl := NewRateLimiter()

	http.HandleFunc("/check", enableCORS(rl.handleCheck))
	http.HandleFunc("/health", enableCORS(rl.handleHealth))

	port := ":8080"
	log.Printf("Rate limiter service starting on port %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
