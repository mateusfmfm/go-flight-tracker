package main

import (
	"context"
	"go-flight-tracker/internal/config"
	"go-flight-tracker/internal/flight"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	client := flight.NewClient(httpClient, cfg)
	poller := flight.NewPoller(client, cfg.PollerInterval)

	go poller.Start(ctx)

	log.Println("Server started! Waiting for data from OpenSky (Press Ctrl+C to exit)...")

	for {
		select {
		case aircrafts, ok := <-poller.Output():
			if !ok {
				return
			}
			log.Printf("Received update: %d active aircrafts in radar\n", len(aircrafts))

		case <-ctx.Done():
			log.Println("Sinal de encerramento recebido. Finalizando aplicação graciosamente...")
			return
		}
	}
}
