package photos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL   = "https://api.planespotters.net/pub/photos/hex"
	defaultUserAgent = "go-flight-tracker/1.0 (+https://github.com/mateusfmfm/go-flight-tracker)"
	maxPhotos        = 3
)

// Photo is a provider-agnostic aircraft image.
type Photo struct {
	ID           string
	ThumbnailURL string
	URL          string
	Photographer string
	Link         string
}

// Cache stores serialized photo lists (e.g. Redis). Optional.
type Cache interface {
	GetBytes(ctx context.Context, key string) ([]byte, error)
	SetBytes(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

// Service fetches aircraft photos by ICAO24 hex.
type Service struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
	cache      Cache
	cacheTTL   time.Duration
}

func NewService(httpClient *http.Client, cache Cache) *Service {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 8 * time.Second}
	}
	return &Service{
		httpClient: httpClient,
		baseURL:    defaultBaseURL,
		userAgent:  defaultUserAgent,
		cache:      cache,
		cacheTTL:   6 * time.Hour,
	}
}

// ByICAO returns up to 3 photos for the given ICAO24.
// On any failure (network, status, decode) it returns an empty slice and nil error
// so GraphQL clients can show a placeholder without handling API errors.
func (s *Service) ByICAO(ctx context.Context, icao24 string) ([]Photo, error) {
	icao := strings.ToLower(strings.TrimSpace(icao24))
	if icao == "" {
		return []Photo{}, nil
	}

	cacheKey := "aircraft:photos:" + icao
	if s.cache != nil {
		if raw, err := s.cache.GetBytes(ctx, cacheKey); err == nil && len(raw) > 0 {
			var cached []Photo
			if json.Unmarshal(raw, &cached) == nil {
				return cached, nil
			}
		}
	}

	photos, err := s.fetch(ctx, icao)
	if err != nil {
		return []Photo{}, nil
	}
	if len(photos) > maxPhotos {
		photos = photos[:maxPhotos]
	}

	if s.cache != nil {
		if raw, err := json.Marshal(photos); err == nil {
			_ = s.cache.SetBytes(ctx, cacheKey, raw, s.cacheTTL)
		}
	}

	return photos, nil
}

func (s *Service) fetch(ctx context.Context, icao string) ([]Photo, error) {
	url := fmt.Sprintf("%s/%s", strings.TrimRight(s.baseURL, "/"), icao)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("photos api status %d", resp.StatusCode)
	}

	var payload planespottersResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	out := make([]Photo, 0, len(payload.Photos))
	for _, p := range payload.Photos {
		thumb := p.Thumbnail.Src
		large := p.ThumbnailLarge.Src
		if large == "" {
			large = thumb
		}
		if thumb == "" && large == "" {
			continue
		}
		if thumb == "" {
			thumb = large
		}
		out = append(out, Photo{
			ID:           p.ID,
			ThumbnailURL: thumb,
			URL:          large,
			Photographer: p.Photographer,
			Link:         p.Link,
		})
	}
	return out, nil
}

type planespottersResponse struct {
	Photos []planespottersPhoto `json:"photos"`
}

type planespottersPhoto struct {
	ID             string             `json:"id"`
	Thumbnail      planespottersImage `json:"thumbnail"`
	ThumbnailLarge planespottersImage `json:"thumbnail_large"`
	Link           string             `json:"link"`
	Photographer   string             `json:"photographer"`
}

type planespottersImage struct {
	Src string `json:"src"`
}
