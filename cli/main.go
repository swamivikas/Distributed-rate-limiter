package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"rate-limiter/config"
	"rate-limiter/limiter"
)

func main() {
	root := &cobra.Command{Use: "ratecli", Short: "Rate limiter CLI"}

	pingCmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping Redis server",
		RunE: func(cmd *cobra.Command, args []string) error {
			config.Load()
			rdb, err := limiter.NewRedisClient(config.Cfg.RedisAddr, config.Cfg.RedisDB)
			if err != nil {
				return err
			}
			res := rdb.Ping(context.Background())
			fmt.Println("Redis response:", res.Val())
			return nil
		},
	}

	root.AddCommand(pingCmd)
	_ = root.Execute()
}
