package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"rate-limiter/config"
	"rate-limiter/limiter"
	"rate-limiter/middleware"
)

func main() {
	config.Load()

	rdb, err := limiter.NewRedisClient(config.Cfg.RedisAddr, config.Cfg.RedisDB)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	// Hard-coded rate-limit parameters
	const capacity = 10        // tokens (requests)
	const window = time.Minute // refill window

	l := limiter.New(rdb, capacity, window)

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	})

	// Protected endpoint example
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello world")
	})

	handler := middleware.New(l, "ratelimit:")(mux)

	log.Printf("server listening on %s", config.Cfg.ServerPort)
	if err := http.ListenAndServe(config.Cfg.ServerPort, handler); err != nil {
		log.Fatalf("http: %v", err)
	}
}
