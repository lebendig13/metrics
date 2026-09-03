package config

import (
	"flag"
)

type ServerConfig struct {
	RunAddress string
}

type AgentConfigIntervals struct {
	ReportInterval int
	PollInterval   int
}

type AgentConfig struct {
	ServerAddress string
	Intervals     AgentConfigIntervals
}

func ParseServerFlags() ServerConfig {
	var result ServerConfig
	flag.StringVar(&result.RunAddress, "a", "localhost:8080", "address and port to run server")
	flag.Parse()
	return result
}

func ParseAgentFlags() AgentConfig {
	var result AgentConfig
	flag.StringVar(&result.ServerAddress, "a", "localhost:8080", "address and port of the server")
	flag.IntVar(&result.Intervals.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&result.Intervals.PollInterval, "p", 2, "poll interval")
	flag.Parse()
	return result
}
