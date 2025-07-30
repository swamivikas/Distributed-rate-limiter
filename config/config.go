package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config holds all runtime configuration.
// Values can be supplied via YAML file (config.yaml) or env vars prefixed with RATE_ (e.g. RATE_REDIS_ADDR).
//
// Default limits are reusable for all keys unless overridden dynamically.
// These values are read once at startup. Live editing is possible via CLI / REST.

type Config struct {
	RedisAddr  string        `mapstructure:"redis_addr"`
	RedisDB    int           `mapstructure:"redis_db"`
	Capacity   int           `mapstructure:"capacity"` // tokens
	Window     time.Duration `mapstructure:"window"`   // e.g. 1m, 5s
	ServerPort string        `mapstructure:"server_port"`
}

var Cfg Config // global

// Load reads configuration from .env style env vars or config.{yaml|json|toml}.
// Call this at program start.
func Load() {
	v := viper.New()

	v.SetConfigName("config")
	v.AddConfigPath(".")
	v.SetConfigType("yaml")

	// Defaults
	v.SetDefault("redis_addr", "localhost:6379")
	v.SetDefault("redis_db", 0)
	v.SetDefault("capacity", 100)
	v.SetDefault("window", "1m")
	v.SetDefault("server_port", ":8080")

	v.SetEnvPrefix("rate")
	v.AutomaticEnv()

	_ = v.ReadInConfig() // ignore if file missing

	if err := v.Unmarshal(&Cfg); err != nil {
		log.Fatalf("config: %v", err)
	}

	// Duration comes as string if loaded from env; let viper handle but ensure non-zero
	if Cfg.Window == 0 {
		dur, err := time.ParseDuration(v.GetString("window"))
		if err != nil {
			log.Fatalf("config: invalid window duration: %v", err)
		}
		Cfg.Window = dur
	}
}
