package cors

import (
	"net/http"
	"net/url"
	"strings"
)

// Config holds allowed browser origins for HTTP CORS and WebSocket upgrades.
type Config struct {
	// Origins are full origins, e.g. "http://localhost:4200".
	// Use "*" to allow any origin (dev only).
	Origins []string
}

// ParseOrigins splits a comma-separated CORS_ORIGINS env value.
func ParseOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{"http://localhost:4200", "http://127.0.0.1:4200"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, strings.TrimRight(p, "/"))
		}
	}
	return out
}

func (c Config) AllowAll() bool {
	for _, o := range c.Origins {
		if o == "*" {
			return true
		}
	}
	return false
}

// Allows reports whether origin is permitted (empty origin = non-browser / same-tab).
func (c Config) Allows(origin string) bool {
	if origin == "" || c.AllowAll() {
		return true
	}
	origin = strings.TrimRight(origin, "/")
	for _, o := range c.Origins {
		if strings.EqualFold(o, origin) {
			return true
		}
	}
	return false
}

// OriginPatterns returns host patterns for coder/websocket AcceptOptions.
// Example: "http://localhost:4200" → "localhost:4200"
func (c Config) OriginPatterns() []string {
	if c.AllowAll() {
		return nil
	}
	patterns := make([]string, 0, len(c.Origins))
	for _, o := range c.Origins {
		u, err := url.Parse(o)
		if err != nil || u.Host == "" {
			continue
		}
		patterns = append(patterns, u.Host)
	}
	return patterns
}

// Middleware adds CORS headers and handles OPTIONS preflight.
func Middleware(cfg Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if cfg.Allows(origin) {
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			} else if cfg.AllowAll() {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			w.Header().Set("Access-Control-Allow-Headers",
				"Content-Type, Authorization, GraphQL-Client-Name, GraphQL-Client-Version, Apollo-Require-Preflight")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
