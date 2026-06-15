package proxy

import "net/http"

// NewAuthProxy forwards /api/auth requests to the auth service.
func NewAuthProxy(targetURL string) (http.Handler, error) {
	return NewReverseProxy(targetURL)
}
