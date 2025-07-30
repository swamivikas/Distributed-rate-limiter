package limiter

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// NewRedisClient creates a redis.Client with sane defaults based on address & DB.
// It pings the server to verify connectivity.
func NewRedisClient(addr string, db int) (*redis.Client, error) {
	opts := &redis.Options{
		Addr:         addr,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
