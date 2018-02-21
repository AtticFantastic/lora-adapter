package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mainflux/mainflux/lora-adapter/adapter"
	"go.uber.org/zap"
)

const (
	port               int    = 6070
	defMainfluxMqttURL string = "tcp://localhost:1883"
	envMainfluxMqttURL string = "LORA_ADAPTER_MAINFLUX_URL"
	defLoraMqttURL     string = "tcp://localhost:1884"
	envLoraMqttURL     string = "LORA_ADAPTER_LORASERVER_URL"
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
	if mainfluxAdapter, err := NewBackend(cfg.MainfluxURL, "", "", false); err != nil {
		println("Cannot create the Mainflux backend")
	}

	if loraAdapter, err = NewBackend(cfg.LoraURL, "", "", true); err != nil {
		println("Cannot create LoRa Server backend")
	}

	errs := make(chan error, 3)

	go func() {
		errs <- MainfluxAdapter.Sub()
	}()

	go func() {
		errs <- loraAdapter.Sub()
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errs)
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
