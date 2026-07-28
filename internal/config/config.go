package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// OpenSky / Poller
	OpenSkyClientID     string
	OpenSkyClientSecret string
	OpenSkyBaseURL      string
	PollerInterval      time.Duration
	BoundingBox         *BoundingBox

	// Redis
	RedisAddr    string
	RedisPass    string
	RedisDB      int
	RedisChannel string
}

type BoundingBox struct {
	Lamin float64
	Lomin float64
	Lamax float64
	Lomax float64
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		OpenSkyClientID:     os.Getenv("OPENSKY_CLIENT_ID"),
		OpenSkyClientSecret: os.Getenv("OPENSKY_CLIENT_SECRET"),
		OpenSkyBaseURL:      "https://opensky-network.org/api",
		PollerInterval:      10 * time.Second,

		RedisAddr:    os.Getenv("REDIS_ADDR"),
		RedisPass:    os.Getenv("REDIS_PASSWORD"),
		RedisDB:      getEnvAsInt(os.Getenv("REDIS_DB"), 0),
		RedisChannel: os.Getenv("REDIS_CHANNEL"),
	}

	if laminStr := os.Getenv("OPENSKY_LAMIN"); laminStr != "" {
		lamin, _ := strconv.ParseFloat(laminStr, 64)
		lomin, _ := strconv.ParseFloat(os.Getenv("OPENSKY_LOMIN"), 64)
		lamax, _ := strconv.ParseFloat(os.Getenv("OPENSKY_LAMAX"), 64)
		lomax, _ := strconv.ParseFloat(os.Getenv("OPENSKY_LOMAX"), 64)

		cfg.BoundingBox = &BoundingBox{
			Lamin: lamin,
			Lomin: lomin,
			Lamax: lamax,
			Lomax: lomax,
		}
	}

	return cfg, nil
}

func (c *Config) BuildStatesURL() string {
	baseURL := fmt.Sprintf("%s/states/all", c.OpenSkyBaseURL)
	if c.BoundingBox == nil {
		return baseURL
	}
	return fmt.Sprintf("%s?lamin=%.4f&lomin=%.4f&lamax=%.4f&lomax=%.4f",
		baseURL, c.BoundingBox.Lamin, c.BoundingBox.Lomin, c.BoundingBox.Lamax, c.BoundingBox.Lomax)
}

func getEnvAsInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}
