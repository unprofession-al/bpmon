package webhelpers

import (
	"context"
	"net/http"
	"strings"
)

type TokenAuth struct {
	Tokens map[string]string
}

type key int

const (
	KeyRecipients key = iota
)

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

		ctx := context.WithValue(r.Context(), KeyRecipients, []string{recipient})
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

func RecipientsHeaderAuth(next http.Handler, RecipientsHeaderName string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var groups []string

		header := r.Header.Get(RecipientsHeaderName)
		if header != "" {
			groups = strings.Split(r.Header.Get(RecipientsHeaderName), ",")
		}

		if len(groups) < 1 {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), KeyRecipients, groups)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}
