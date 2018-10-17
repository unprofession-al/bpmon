package dashboard

import (
	"context"
	"net/http"
	"strings"
)

type key int

type TokenAuth struct {
	ContextKey key
	Tokens     map[string]string
}

func (m TokenAuth) Inject(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var authToken string
		authTokens := r.URL.Query()["authtoken"]
		if len(authTokens) > 0 {
			authToken = authTokens[0]
		}
		recipient, ok := m.Tokens[authToken]
		if !ok {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), m.ContextKey, []string{recipient})
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

type HeaderAuth struct {
	ContextKey key
	HeaderName string
}

func (m HeaderAuth) Inject(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var groups []string

		header := r.Header.Get(m.HeaderName)
		if header != "" {
			groups = strings.Split(header, ",")
		}

		if len(groups) < 1 {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), m.ContextKey, groups)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}
