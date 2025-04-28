package api

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				stack := debug.Stack()

				slog.Error("Error recovered from panic",
					slog.String("stack", string(stack)),
					slog.Any("error", err),
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
