package main

import (
	"context"
	"go-flight-tracker/internal/flight"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	client := flight.NewClient(httpClient, "https://opensky-network.org/api")
	poller := flight.NewPoller(client, 10*time.Second)

	go poller.Start(ctx)

	log.Println("Server started! Waiting for data from OpenSky (Press Ctrl+C to exit)...")

	for {
		select {
		case aircrafts, ok := <-poller.Output():
			if !ok {
				// Canal foi fechado
				return
			}
			log.Printf("Received update: %d active aircrafts in radar\n", len(aircrafts))

			// Se quiser ver dados de alguma aeronave:
			if len(aircrafts) > 0 {
				a := aircrafts[0]
				log.Printf("   -> Exemplo: ICAO24: %s | Callsign: %s | País: %s | Altura: %.2fm\n",
					a.Icao24, a.Callsign, a.OriginCountry, a.BaroAltitude)
			}

		case <-ctx.Done():
			log.Println("Sinal de encerramento recebido. Finalizando aplicação graciosamente...")
			return
		}
	}
}
