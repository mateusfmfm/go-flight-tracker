package flight

import (
	"context"
	"encoding/json"
	"fmt"
	"go-flight-tracker/internal/config"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	cfg        *config.Config
}

func NewClient(httpClient *http.Client, cfg *config.Config) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return &Client{
		httpClient: httpClient,
		cfg:        cfg,
	}
}

func (c *Client) GetActiveFlights(ctx context.Context) ([]*Aircraft, error) {
	url := c.cfg.BuildStatesURL()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.cfg.OpenSkyClientID != "" && c.cfg.OpenSkyClientSecret != "" {
		req.SetBasicAuth(c.cfg.OpenSkyClientID, c.cfg.OpenSkyClientSecret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limit reached (429): too many requests to OpenSky Network")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, resp.Status)
	}

	var openSkyResponse OpenSkyResponse
	if err := json.NewDecoder(resp.Body).Decode(&openSkyResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	aircrafts := ParseOpenSkyResponse(&openSkyResponse)
	return aircrafts, nil
}
