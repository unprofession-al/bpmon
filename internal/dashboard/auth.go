package dashboard

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/unprofession-al/bpmon/internal/bpmon"
)

type key int

// TokenAuth is a middleware that allows to read a particular part of the query
// string a request as token, checks if the token is known and matches a user
// and then stores the user in the cotext of the request.
type TokenAuth struct {
	// ContextKey is used as key to store to users/recipients in the Context
	// of the Request.
	ContextKey interface{}

	// Param specifies which parameter of the URLs query string should be
	// considered to hold the token.
	Param string

	// Tokens is a map where k=[API token] and v=[user/recipient]. If the token
	// provided in the url param matches with a token in the map the
	// corresponding user/recipient will be stored in the context of the
	// request.
	Tokens map[string]string
}

// Wrap returns the the middleware as http.Handler.
func (m TokenAuth) Wrap(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var authToken string
		authTokens := r.URL.Query()[m.Param]
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

// HeaderAuth provides a middleware which allows to read a user (aka recipient)
// name from a HTTP header. The user/recipient is then stored in the context of
// the request.
//
// This is for example useful when used in combination
// with keycloak/keycloak-gatekeeper. See...
//   https://www.keycloak.org/
//   https://github.com/keycloak/keycloak-gatekeeper
// ... for details.
type HeaderAuth struct {
	// ContextKey is used as key to store to users/recipients in the context
	// of the request.
	ContextKey interface{}

	// HeaderName specifies the name of the HTTP header in which the users/
	// recipients are provided.
	HeaderName string
}

// Wrap returns the the middleware as http.Handler.
func (m HeaderAuth) Wrap(next http.Handler) http.Handler {
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

// Authorization provides a middleware that reads the user/recipient name from
// the request context and decides based on the '{bp}' URL param if the
// recipient is allowed to access the ressource.
type Authorization struct {
	// Context key of the user/recipient value, eg. where to find the
	// user/recipient in the context of the request.
	RecipientContextKey interface{}

	// ProtectPattern that matches all paths to be protected.
	ProtectPattern *regexp.Regexp

	// HTTP Stus Code that will be returned in case of unauthorized requests.
	OnAuthErrorReturn int

	// BP holds the BP definitions including the allowed recipients per BP.
	BP bpmon.BusinessProcesses
}

// Wrap returns the the middleware as http.Handler.
// TODO: This should be improved...
func (m Authorization) Wrap(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if !m.ProtectPattern.MatchString(path) {
			next.ServeHTTP(w, r)
			return
		}
		if recipients := r.Context().Value(m.RecipientContextKey); recipients != nil {
			for _, bp := range m.BP.GetByRecipients(recipients.([]string)) {
				if strings.Contains(path, bp.ID) {
					next.ServeHTTP(w, r)
					return
				}
			}
		}
		http.Error(w, "", m.OnAuthErrorReturn)
	}

	return http.HandlerFunc(fn)
}
