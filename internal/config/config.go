package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	OpenSkyClientID     string
	OpenSkyClientSecret string
	OpenSkyBaseURL      string
	PollerInterval      time.Duration
	BoundingBox         *BoundingBox
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
