package httpapi

import (
	"net/http"
	"slices"
)

// CORS allows browser requests from the given origins. "*" allows any origin.
func CORS(allowedOrigins []string) Middleware {
	allowAll := slices.Contains(allowedOrigins, allOrigins)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get(headerOrigin)
			w.Header().Add(headerVary, headerOrigin)

			switch {
			case allowAll:
				w.Header().Set(headerAllowOrigin, allOrigins)
			case origin != "" && slices.Contains(allowedOrigins, origin):
				w.Header().Set(headerAllowOrigin, origin)
			}
			w.Header().Set(headerExposeHeaders, headerTraceID)

			if r.Method == http.MethodOptions {
				w.Header().Set(headerAllowMethods, "GET, POST, OPTIONS")
				w.Header().Set(headerAllowHeaders, headerContentType)
				w.Header().Set(headerMaxAge, "600")
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
