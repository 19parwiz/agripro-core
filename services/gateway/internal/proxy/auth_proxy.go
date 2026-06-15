package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewAuthProxy forwards /api/auth requests to the auth service.
func NewAuthProxy(targetURL string) (http.Handler, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("parse auth service url: %w", err)
	}

	return httputil.NewSingleHostReverseProxy(target), nil
}
