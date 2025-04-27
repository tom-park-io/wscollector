package main

import (
	"wscollector/config"
	"wscollector/internal/bybit/collector"
	"wscollector/logger"

	"go.uber.org/zap"
)

func main() {
	// viper config
	cfg := config.Load()

	// zap logger
	log, err := logger.New(cfg.Log)
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}
	defer log.Sync()

	// run collector
	if err := collector.StartCollector(cfg, log); err != nil {
		log.Fatal("collector failed", zap.Error(err))
	}

	select {}
}
