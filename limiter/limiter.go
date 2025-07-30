package limiter

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// Limiter implements a distributed token-bucket rate limiter backed by Redis.
// The algorithm maintains a bucket of size `capacity` that refills at `rate` tokens per second.
// Each request tries to consume one token (or N tokens) from the bucket.
// When the bucket is empty the request is rejected.
//
// The state for each key is kept in Redis as a hash with two fields:
//   tokens    – current token count (float stored as string)
//   timestamp – unix time of the last refill event (seconds)
//
// All updates are performed by a Lua script to guarantee atomicity under concurrent access.
// The script also sets a TTL on the key (2× burst window) so idle buckets disappear automatically.
//
// A typical key could be "ratelimit:<ip>" or "ratelimit:<userId>".
//
// Usage:
//   l := limiter.New(client, 100, time.Minute) // 100 requests per minute → rate = 100/60
//   ok, err := l.Allow(ctx, "ratelimit:"+ip)
//   if !ok { /* return 429 */ }
//
// The limiter is safe for concurrent use.

type Limiter struct {
	redis     *redis.Client
	capacity  float64       // maximum number of tokens in the bucket
	rate      float64       // tokens regenerated per second
	ttl       time.Duration // TTL to set for idle buckets
	luaScript *redis.Script // cached Lua script
}

// New creates a new Limiter.
//
//	capacity – maximum burst size (max tokens)
//	window   – logical window that defines the refill speed (capacity/window tokens per second).
//	          For example capacity=100 and window=1m → ~1.67 tokens per second.
//
// The TTL for Redis keys is set to 2×window so buckets for inactive keys eventually expire.
func New(client *redis.Client, capacity int, window time.Duration) *Limiter {
	if capacity <= 0 {
		panic("limiter: capacity must be > 0")
	}
	if window <= 0 {
		panic("limiter: window must be > 0")
	}

	rate := float64(capacity) / window.Seconds()

	lua := redis.NewScript(luaSource)

	return &Limiter{
		redis:     client,
		capacity:  float64(capacity),
		rate:      rate,
		ttl:       window * 2,
		luaScript: lua,
	}
}

// Allow attempts to consume <tokens> (default 1) from the bucket identified by key.
// It returns true if the request should be allowed, false otherwise.
func (l *Limiter) Allow(ctx context.Context, key string, tokens ...int) (bool, error) {
	n := 1
	if len(tokens) > 0 {
		n = tokens[0]
	}
	if n <= 0 {
		return true, nil // zero or negative consumption trivially allowed
	}

	now := time.Now().Unix()

	// Run the Lua script atomically.
	res, err := l.luaScript.Run(ctx, l.redis, []string{key},
		l.rate,               // ARGV[1] rate (tokens per second)
		l.capacity,           // ARGV[2] capacity
		now,                  // ARGV[3] current unix time (seconds)
		n,                    // ARGV[4] tokens requested
		int(l.ttl.Seconds()), // ARGV[5] TTL seconds
	).Int64()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

// Lua script implementing token bucket logic atomically.
// Input:
//
//	KEYS[1] – bucket key
//	ARGV[1] – rate (tokens per second, float)
//	ARGV[2] – capacity (max tokens, float)
//	ARGV[3] – now (current unix time, integer seconds)
//	ARGV[4] – requested tokens (integer)
//	ARGV[5] – ttl (key expiration in seconds)
//
// Returns 1 if allowed, 0 if limited.
const luaSource = `
local rate       = tonumber(ARGV[1])
local capacity   = tonumber(ARGV[2])
local now        = tonumber(ARGV[3])
local requested  = tonumber(ARGV[4])
local ttl        = tonumber(ARGV[5])

-- fetch existing bucket state
local data = redis.call('HMGET', KEYS[1], 'tokens', 'timestamp')
local tokens = tonumber(data[1])
local ts     = tonumber(data[2])

if tokens == nil then
  tokens = capacity
  ts = now
end

-- refill based on time elapsed
local delta = math.max(0, now - ts)
local new_tokens = math.min(capacity, tokens + delta * rate)

local allowed = new_tokens >= requested
if allowed then
  new_tokens = new_tokens - requested
end

-- persist state and TTL
redis.call('HMSET', KEYS[1], 'tokens', new_tokens, 'timestamp', now)
redis.call('EXPIRE', KEYS[1], ttl)

if allowed then
  return 1
else
  return 0
end
`
