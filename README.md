# Distributed Rate Limiter (Go)

Distributed token-bucket rate limiter using Redis as the central state store.

## Features
* Token-bucket algorithm (Lua for atomicity)
* Configurable per-IP / user limits
* Minimal HTTP middleware
* CLI & tests

## Quick start
```bash
git clone …
cd rate-limiter
go run .
```

Ensure Redis is running locally at the configured address (default `localhost:6379`). Then visit http://localhost:8080/ping

## Configuration
Env vars (prefix `RATE_`) or `config.yaml`:
* `REDIS_ADDR` (default `localhost:6379`)
* `CAPACITY` tokens (default 100)
* `WINDOW` duration (default `1m`)
* `SERVER_PORT` (default `:8080`)

## Benchmarks
Run `go test ./... -bench=.` (TODO: add benchmark script). 