package main

import (
	config "github.com/lebendig13/metrics/internal/config"
)

func main() {
	configP := config.GetAgentConfig()

	agent := NewAgent(&configP)
	agent.Process()
}
