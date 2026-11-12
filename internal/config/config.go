package config

import (
	"flag"
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
)

type ServerConfig struct {
	RunAddress      string `env:"ADDRESS"`
	LogLevel        string `env:"LOG_LEVEL"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	HashKey         string `env:"KEY"`
}

type AgentConfigIntervals struct {
	ReportInterval int `env:"REPORT_INTERVAL"`
	PollInterval   int `env:"POLL_INTERVAL"`
}

type AgentConfig struct {
	ServerAddress string `env:"ADDRESS"`
	Intervals     AgentConfigIntervals
	HashKey       string `env:"KEY"`
	RateLimit     int    `env:"RATE_LIMIT"`
}

func GetServerConfig() ServerConfig {
	var result ServerConfig
	flag.StringVar(&result.RunAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&result.LogLevel, "l", "info", "log level")
	flag.IntVar(&result.StoreInterval, "i", 300, "store interval")
	flag.StringVar(&result.FileStoragePath, "f", "/home/user/metrics/metrics_values.txt", "file storage path")
	flag.BoolVar(&result.Restore, "r", false, "restore saved metrics when server starts")
	flag.StringVar(&result.DatabaseDSN, "d", fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", `localhost`, `metricsuser`, `metricsuser`, `metrics`), "database connection string")
	flag.StringVar(&result.HashKey, "k", "", "hash key")
	flag.Parse()

	if err := env.Parse(&result); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	return result
}

func GetAgentConfig() AgentConfig {
	var result AgentConfig
	flag.StringVar(&result.ServerAddress, "a", "localhost:8080", "address and port of the server")
	flag.IntVar(&result.Intervals.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&result.Intervals.PollInterval, "p", 2, "poll interval")
	flag.StringVar(&result.HashKey, "k", "", "hash key")
	flag.IntVar(&result.RateLimit, "l", 1, "rate limit")
	flag.Parse()

	if err := env.Parse(&result); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	if result.RateLimit < 1 {
		log.Printf("Rate limit must be positive, got %d; using 1", result.RateLimit)
		result.RateLimit = 1
	}

	return result
}
