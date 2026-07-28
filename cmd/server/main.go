package main

import (
	"context"
	"go-flight-tracker/internal/config"
	"go-flight-tracker/internal/flight"
	"go-flight-tracker/internal/redis"
	"go-flight-tracker/internal/store"
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

	redisClient, err := redis.NewClient(ctx, *cfg)
	if err != nil {
		log.Fatalf("failed to initialize redis:: %v", err)
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	client := flight.NewClient(httpClient, cfg)
	poller := flight.NewPoller(client, cfg.PollerInterval)

	aircraftStore := store.NewAircraftStore()

	go poller.Start(ctx)

	log.Println("Server started! Waiting for data from OpenSky (Press Ctrl+C to exit)...")

	subChan, err := redisClient.SubscribeAircrafts(ctx)
	if err != nil {
		log.Fatalf("failed to subscribe to redis: %v", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case aircrafts, ok := <-subChan:
				if !ok {
					return
				}
				log.Printf("📡 [REDIS PUB/SUB] Event received! Broadcasted %d aircrafts to subscribers\n", len(aircrafts))
			}
		}
	}()

	for {
		select {
		case aircrafts, ok := <-poller.Output():
			if !ok {
				return
			}
			aircraftStore.Update(aircrafts)

			stored := aircraftStore.GetAll()
			log.Printf("Received update: %d aircrafts from API | Total active in memory store: %d\n",
				len(aircrafts), len(stored))

			if len(stored) > 0 {
				a := stored[0]
				log.Printf("   -> Store sample: ICAO24: %s | Callsign: %s | Alt: %.2fm\n",
					a.Icao24, a.Callsign, a.BaroAltitude)
			}

		case <-ctx.Done():
			log.Println("Closing application...")
			return
		}
	}
}
