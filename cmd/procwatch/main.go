package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/procwatch/internal/config"
	"github.com/user/procwatch/internal/logger"
	"github.com/user/procwatch/internal/supervisor"
)

func main() {
	cfgPath := flag.String("config", "procwatch.yaml", "path to config file")
	logLevel := flag.String("log-level", "info", "log level (debug, info, warn, error)")
	flag.Parse()

	log := logger.New(os.Stdout, *logLevel)

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log.Info("procwatch starting", map[string]any{
		"processes": len(cfg.Processes),
		"config":    *cfgPath,
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	sv := supervisor.New(cfg, log)
	if err := sv.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "supervisor error: %v\n", err)
		os.Exit(1)
	}

	log.Info("procwatch stopped", nil)
}
