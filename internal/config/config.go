package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

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

	var envConfig ServerConfig
	err := env.Parse(&envConfig)
	if err != nil {
		log.Fatal("Cannot get environment variables \"ADDRESS\", \"LOG_LEVEL\", \"STORE_INTERVAL\", \"FILE_STORAGE_PATH\", \"RESTORE\", \"DATABASE_DSN\": ", err)
	}
	if envConfig.RunAddress != "" {
		result.RunAddress = envConfig.RunAddress
	}
	if envConfig.LogLevel != "" {
		result.LogLevel = envConfig.LogLevel
	}
	siEnv, siEnvExists := os.LookupEnv("STORE_INTERVAL")
	if siEnvExists && siEnv != "" {
		siEnvValue, err := strconv.Atoi(siEnv)
		if err != nil {
			log.Fatalf("Invalid STORE_INTERVAL value '%v': %v", siEnv, err)
		}
		result.StoreInterval = siEnvValue
	}
	if envConfig.FileStoragePath != "" {
		result.FileStoragePath = envConfig.FileStoragePath
	}
	rEnv, rEnvExists := os.LookupEnv("RESTORE")
	if rEnvExists && rEnv != "" {
		rEnvValue, err := strconv.ParseBool(rEnv)
		if err != nil {
			log.Fatalf("Invalid RESTORE value '%v': %v", rEnv, err)
		}
		result.Restore = rEnvValue
	}
	if envConfig.DatabaseDSN != "" {
		result.DatabaseDSN = envConfig.DatabaseDSN
	}
	if envConfig.HashKey != "" {
		result.HashKey = envConfig.HashKey
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

	var envConfig AgentConfig
	err := env.Parse(&envConfig)
	if err != nil {
		log.Fatal("Cannot get environment variables \"ADDRESS\", \"REPORT_INTERVAL\", \"POLL_INTERVAL\": ", err)
	}

	if envConfig.ServerAddress != "" {
		result.ServerAddress = envConfig.ServerAddress
	}
	riEnv, riEnvExists := os.LookupEnv("REPORT_INTERVAL")
	if riEnvExists && riEnv != "" {
		riEnvValue, err := strconv.Atoi(riEnv)
		if err != nil {
			log.Fatalf("Invalid REPORT_INTERVAL value '%v': %v", riEnv, err)
		}
		result.Intervals.ReportInterval = riEnvValue
	}
	piEnv, piEnvExists := os.LookupEnv("POLL_INTERVAL")
	if piEnvExists && piEnv != "" {
		piEnvValue, err := strconv.Atoi(piEnv)
		if err != nil {
			log.Fatalf("Invalid POLL_INTERVAL value '%v': %v", piEnv, err)
		}
		result.Intervals.PollInterval = piEnvValue
	}
	if envConfig.HashKey != "" {
		result.HashKey = envConfig.HashKey
	}
	rlEnv, rlEnvExists := os.LookupEnv("RATE_LIMIT")
	if rlEnvExists && rlEnv != "" {
		rlEnvValue, err := strconv.Atoi(rlEnv)
		if err != nil {
			log.Fatalf("Invalid RATE_LIMIT value '%v': %v", rlEnv, err)
		}
		result.RateLimit = rlEnvValue
	}
	if result.RateLimit < 1 {
		log.Printf("Rate limit must be positive, got %d; using 1", result.RateLimit)
		result.RateLimit = 1
	}

	return result
}
