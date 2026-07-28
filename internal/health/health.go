package health

import (
	"context"
	"go-flight-tracker/internal/redis"
	"net/http"
	"time"
)

func RegisterHandlers(redisClient *redis.Client) {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx); err != nil {
			http.Error(w, "Redis connection failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("READY"))
	})
}
