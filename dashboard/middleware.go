package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
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
			groups = strings.Split(header, ",")
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

func HeaderAuthMatcher(next http.Handler, HeaderName string, HeaderValuesAllowed []string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var values []string

		headerContent := r.Header.Get(HeaderName)
		if headerContent != "" {
			values = strings.Split(headerContent, ",")
		}

		isAllowed := false

		for _, val := range values {
			for _, matchVal := range HeaderValuesAllowed {
				if val == matchVal {
					isAllowed = true
				}
			}
		}

		if !isAllowed {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

type logmgs struct {
	Timestamp      string        `json:"timestamp"`
	Status         int           `json:"status"`
	Size           int           `json:"size"`
	Method         string        `json:"method"`
	Request        string        `json:"request"`
	RequestHeaders http.Header   `json:"request_headers"`
	Latency        time.Duration `json:"latency"`
}

type customResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (c *customResponseWriter) WriteHeader(status int) {
	c.status = status
	c.ResponseWriter.WriteHeader(status)
}

func (c *customResponseWriter) Write(b []byte) (int, error) {
	size, err := c.ResponseWriter.Write(b)
	c.size += size
	return size, err
}

func newCustomResponseWriter(w http.ResponseWriter) *customResponseWriter {
	// When WriteHeader is not called, it's safe to assume the status will be 200.
	return &customResponseWriter{
		ResponseWriter: w,
		status:         200,
	}
}

func Logger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		crw := newCustomResponseWriter(w)
		next.ServeHTTP(crw, r)
		end := time.Now()

		msg := &logmgs{
			Timestamp:      end.Format("2006/01/02-15:04:05.000"),
			Status:         crw.status,
			Size:           crw.size,
			Latency:        end.Sub(start),
			Method:         r.Method,
			Request:        r.URL.Path,
			RequestHeaders: r.Header,
		}

		b, _ := json.Marshal(msg)
		fmt.Println(string(b))
	}

	return http.HandlerFunc(fn)
}
