package server

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"time"
)

func LoggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration", time.Since(start).String(),
			"client", r.RemoteAddr,
		)
	})
}

func APIKeyAuthMiddleware(validAPIKey string, next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientAPIKey := r.Header.Get("X-API-Key")
		if clientAPIKey == "" {
			authHeader := r.Header.Get("Authorization")
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				clientAPIKey = authHeader[7:]
			}
		}

		if subtle.ConstantTimeCompare([]byte(clientAPIKey), []byte(validAPIKey)) != 1 {
			if logger != nil {
				logger.Warn("authentication failed",
					"reason", "invalid key",
					"client", r.RemoteAddr,
					"path", r.URL.Path,
				)
			}
			w.Header().Set("WWW-Authenticate", `Bearer realm="API", error="invalid_token"`)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": "Unauthorized", "message": "API key required. Use X-API-Key header or Authorization: Bearer <key>"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
