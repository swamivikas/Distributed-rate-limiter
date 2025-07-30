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

## default setting is 10 req per min per IP address 

<img width="1591" height="707" alt="Screenshot 2025-07-30 202450" src="https://github.com/user-attachments/assets/b5097df1-0fc7-4397-b28a-f6c433dd1c78" />


<img width="1526" height="733" alt="Screenshot 2025-07-30 202511" src="https://github.com/user-attachments/assets/aba97728-176d-457c-b693-3fda45653cf8" />
