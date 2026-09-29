package photos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type memCache struct {
	data map[string][]byte
}

func (m *memCache) GetBytes(_ context.Context, key string) ([]byte, error) {
	return m.data[key], nil
}

func (m *memCache) SetBytes(_ context.Context, key string, value []byte, _ time.Duration) error {
	if m.data == nil {
		m.data = map[string][]byte{}
	}
	m.data[key] = value
	return nil
}

func TestByICAO_MapsAndLimits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"photos": []map[string]any{
				{
					"id": "1",
					"thumbnail": map[string]any{"src": "https://example.com/t1.jpg"},
					"thumbnail_large": map[string]any{"src": "https://example.com/l1.jpg"},
					"link":         "https://example.com/1",
					"photographer": "A",
				},
				{
					"id": "2",
					"thumbnail": map[string]any{"src": "https://example.com/t2.jpg"},
					"thumbnail_large": map[string]any{"src": "https://example.com/l2.jpg"},
					"link":         "https://example.com/2",
					"photographer": "B",
				},
				{
					"id": "3",
					"thumbnail": map[string]any{"src": "https://example.com/t3.jpg"},
					"thumbnail_large": map[string]any{"src": "https://example.com/l3.jpg"},
					"link":         "https://example.com/3",
					"photographer": "C",
				},
				{
					"id": "4",
					"thumbnail": map[string]any{"src": "https://example.com/t4.jpg"},
					"thumbnail_large": map[string]any{"src": "https://example.com/l4.jpg"},
					"link":         "https://example.com/4",
					"photographer": "D",
				},
			},
		})
	}))
	defer server.Close()

	svc := NewService(server.Client(), &memCache{})
	svc.baseURL = server.URL

	photos, err := svc.ByICAO(context.Background(), "E48B00")
	if err != nil {
		t.Fatal(err)
	}
	if len(photos) != 3 {
		t.Fatalf("expected max 3 photos, got %d", len(photos))
	}
	if photos[0].URL != "https://example.com/l1.jpg" {
		t.Fatalf("unexpected url: %s", photos[0].URL)
	}
}

func TestByICAO_FetchErrorReturnsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"blocked"}`))
	}))
	defer server.Close()

	svc := NewService(server.Client(), nil)
	svc.baseURL = server.URL

	photos, err := svc.ByICAO(context.Background(), "deadbeef")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(photos) != 0 {
		t.Fatalf("expected empty list, got %d", len(photos))
	}
}
