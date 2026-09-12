package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

type ServerConfig struct {
	RunAddress string `env:"ADDRESS"`
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
	err := env.Parse(&result)
	if err != nil {
		log.Fatal("Cannot get environment variable \"ADDRESS\": ", err)
	}
	if result.RunAddress == "" {
		flag.StringVar(&result.RunAddress, "a", "localhost:8080", "address and port to run server")
		flag.Parse()
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
