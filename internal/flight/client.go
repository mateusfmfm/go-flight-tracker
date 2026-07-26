package flight

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	HttpClient *http.Client
	BaseURL    string
}

func NewClient(httpClient *http.Client, baseURL string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	if baseURL == "" {
		baseURL = "https://opensky-network.org/api"
	}

	return &Client{
		HttpClient: httpClient,
		BaseURL:    baseURL,
	}
}

func (c *Client) GetActiveFlights(ctx context.Context) ([]*Aircraft, error) {
	url := fmt.Sprintf("%s/states/all", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("too many requests")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get active flights: %s", resp.Status)
	}

	var openSkyResponse OpenSkyResponse
	err = json.NewDecoder(resp.Body).Decode(&openSkyResponse)
	if err != nil {
		return nil, err
	}
	aircrafts := ParseOpenSkyResponse(&openSkyResponse)
	return aircrafts, nil
}
