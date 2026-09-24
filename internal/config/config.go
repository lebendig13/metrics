package config

import (
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v11"
)

type ServerConfig struct {
	RunAddress      string `env:"ADDRESS"`
	LogLevel        string `env:"LOG_LEVEL"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

type AgentConfigIntervals struct {
	ReportInterval int `env:"REPORT_INTERVAL"`
	PollInterval   int `env:"POLL_INTERVAL"`
}

type AgentConfig struct {
	ServerAddress string `env:"ADDRESS"`
	Intervals     AgentConfigIntervals
}

func GetServerConfig() ServerConfig {
	var result ServerConfig
	flag.StringVar(&result.RunAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&result.LogLevel, "l", "info", "log level")
	flag.IntVar(&result.StoreInterval, "i", 300, "store interval")
	flag.StringVar(&result.FileStoragePath, "f", "/home/user/metrics/metrics_values.txt", "file storage path")
	flag.BoolVar(&result.Restore, "r", false, "restore saved metrics when server starts")
	flag.Parse()

	var envConfig ServerConfig
	err := env.Parse(&envConfig)
	if err != nil {
		log.Fatal("Cannot get environment variables \"ADDRESS\", \"LOG_LEVEL\", \"STORE_INTERVAL\", \"FILE_STORAGE_PATH\", \"RESTORE\": ", err)
	}
	if envConfig.RunAddress != "" {
		result.RunAddress = envConfig.RunAddress
	}
	if envConfig.LogLevel != "" {
		result.LogLevel = envConfig.LogLevel
	}
	if os.Getenv("STORE_INTERVAL") != "" {
		result.StoreInterval = envConfig.StoreInterval
	}
	if envConfig.FileStoragePath != "" {
		result.FileStoragePath = envConfig.FileStoragePath
	}
	if envConfig.Restore {
		result.Restore = envConfig.Restore
	}
	return result
}

func GetAgentConfig() AgentConfig {
	var result AgentConfig

	flag.StringVar(&result.ServerAddress, "a", "localhost:8080", "address and port of the server")
	flag.IntVar(&result.Intervals.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&result.Intervals.PollInterval, "p", 2, "poll interval")
	flag.Parse()

	var envConfig AgentConfig
	err := env.Parse(&envConfig)
	if err != nil {
		log.Fatal("Cannot get environment variables \"ADDRESS\", \"REPORT_INTERVAL\", \"POLL_INTERVAL\": ", err)
	}

	if envConfig.ServerAddress != "" {
		result.ServerAddress = envConfig.ServerAddress
	}
	if envConfig.Intervals.ReportInterval != 0 {
		result.Intervals.ReportInterval = envConfig.Intervals.ReportInterval
	}
	if envConfig.Intervals.PollInterval != 0 {
		result.Intervals.PollInterval = envConfig.Intervals.PollInterval
	}

	log.Println("Server address: ", result.ServerAddress,
		"; pollInterval = ", result.Intervals.PollInterval,
		"; reportInterval = ", result.Intervals.ReportInterval)
	return result
}
