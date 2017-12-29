package webhelpers

import (
	"context"
	"net/http"
)

type TokenAuth struct {
	Tokens map[string]string
}

func (ta TokenAuth) Create(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var authToken string
		authTokens := r.URL.Query()["authtoken"]
		if len(authTokens) > 0 {
			authToken = authTokens[0]
		}
		recipient, ok := ta.Tokens[authToken]
		if !ok {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "recipient", recipient)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}
