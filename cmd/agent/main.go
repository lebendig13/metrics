package main

import (
	config "github.com/lebendig13/metrics/internal/config"
)

func main() {
	configFlags := config.ParseAgentFlags()

	agent := NewAgent(&configFlags)
	agent.Process()
}
