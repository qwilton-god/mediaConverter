package middleware

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					traceID := GetTraceID(r.Context())
					logger.Error("Panic recovered",
						zap.String("trace_id", traceID),
						zap.Any("error", err),
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{
						"error":    "Internal server error",
						"trace_id": traceID,
					})
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
