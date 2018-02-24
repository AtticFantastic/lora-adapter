package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	adapter "github.com/mainflux/lora-adapter"
	"go.uber.org/zap"
)

const (
	port           int    = 6070
	defMainfluxURL string = "tcp://localhost:1883"
	envMainfluxURL string = "LORA_ADAPTER_MAINFLUX_URL"
	defLoraURL     string = "tcp://localhost:1884"
	envLoraURL     string = "LORA_ADAPTER_LORASERVER_URL"
)

type config struct {
	Port        int
	MainfluxURL string
	LoraURL     string
}

func main() {
	cfg := config{
		Port:        port,
		MainfluxURL: getenv(envMainfluxURL, defMainfluxURL),
		LoraURL:     getenv(envLoraURL, defLoraURL),
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	adapter.InitLogger(logger)

	// Create adapters that connect as MQTT clients to brokers of Mainflux and LoRa Server
	mainfluxAdapter, err := adapter.NewAdapter(cfg.MainfluxURL, "", "", false)
	if err != nil {
		logger.Error("Failed to Mainflux adapter", zap.Error(err))
		return
	}

	loraAdapter, err := adapter.NewAdapter(cfg.LoraURL, "", "", true)
	if err != nil {
		logger.Error("Failed to LoRa Server adapter", zap.Error(err))
		return
	}

	errs := make(chan error, 3)

	go func() {
		errs <- mainfluxAdapter.Sub()
	}()

	go func() {
		errs <- loraAdapter.Sub()
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	c := <-errs
	logger.Info("terminated", zap.String("error", c.Error()))
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
