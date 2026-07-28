package graph

import (
	"go-flight-tracker/internal/redis"
	"go-flight-tracker/internal/store"
)

type Resolver struct {
	Store       *store.AircraftStore
	RedisClient *redis.Client
}
