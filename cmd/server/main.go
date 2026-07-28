package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-flight-tracker/graph"
	"go-flight-tracker/internal/config"
	"go-flight-tracker/internal/flight"
	"go-flight-tracker/internal/redis"
	"go-flight-tracker/internal/store"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
)

func main() {
	// Load env and app settings
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Cancel on Ctrl+C / SIGTERM so everything shuts down cleanly
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Connect to Redis (Pub/Sub + cache)
	redisClient, err := redis.NewClient(ctx, *cfg)
	if err != nil {
		log.Fatalf("failed to initialize redis:: %v", err)
	}

	// Shared HTTP client for OpenSky requests
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// OpenSky client + poller that fetches aircraft on an interval
	client := flight.NewClient(httpClient, cfg)
	poller := flight.NewPoller(client, cfg.PollerInterval)

	// In-memory store for the latest aircraft state
	aircraftStore := store.NewAircraftStore()

	// Start polling OpenSky in the background
	go poller.Start(ctx)

	log.Println("Server started! Waiting for data from OpenSky (Press Ctrl+C to exit)...")

	// Listen for aircraft updates published on Redis Pub/Sub
	subChan, err := redisClient.SubscribeAircrafts(ctx)
	if err != nil {
		log.Fatalf("failed to subscribe to redis: %v", err)
	}

	// Background worker that logs each Redis broadcast
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

	// Wire GraphQL schema with the in-memory store and Redis client
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{Store: aircraftStore, RedisClient: redisClient}}))

	// Enable WebSocket (subscriptions) plus standard HTTP transports
	srv.AddTransport(transport.Websocket{KeepAlivePingInterval: 10 * time.Second})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	// Playground UI at / and GraphQL API at /query
	http.Handle("/", playground.Handler("GraphQL Playground", "/query"))
	http.Handle("/query", srv)

	// Serve GraphQL Playground and the /query endpoint
	go func() {
		log.Println("🚀 GraphQL Playground available at http://localhost:8080/")
		if err := http.ListenAndServe(":8080", nil); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	// Main loop: consume poller updates until shutdown
	for {
		select {
		case aircrafts, ok := <-poller.Output():
			if !ok {
				return
			}

			// Refresh local state with the latest OpenSky snapshot
			aircraftStore.Update(aircrafts)

			// Cache a short-lived snapshot, then broadcast to subscribers
			_ = redisClient.SetCache(ctx, "aircrafts:snapshot", aircrafts, 30*time.Second)
			_ = redisClient.PublishAircrafts(ctx, aircrafts)

			// Quick sanity log of how many aircraft we currently track
			stored := aircraftStore.GetAll()
			log.Printf("Received update: %d aircrafts from API | Total active in memory store: %d\n",
				len(aircrafts), len(stored))

			// Print one sample aircraft from the store
			if len(stored) > 0 {
				a := stored[0]
				log.Printf("   -> Store sample: ICAO24: %s | Callsign: %s | Alt: %.2fm\n",
					a.Icao24, a.Callsign, a.BaroAltitude)
			}

		case <-ctx.Done():
			return
		}
	}
}
