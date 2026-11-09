package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	config "github.com/lebendig13/metrics/internal/config"
)

func main() {
	configP := config.GetAgentConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	agent := NewAgent(&configP)
	agent.Process(ctx)
}
