package middleware

import (
	"net/http"
	"strings"
)

// BearerAuth validates Authorization: Bearer <token> against a set of valid
// tokens. Multiple tokens enable zero-downtime rotation.
func BearerAuth(tokens []string, next http.Handler) http.Handler {
	set := make(map[string]struct{}, len(tokens))
	for _, t := range tokens {
		set[t] = struct{}{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if _, ok := set[token]; !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
