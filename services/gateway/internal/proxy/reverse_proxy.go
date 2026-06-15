package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewReverseProxy forwards requests to another service.
func NewReverseProxy(targetURL string) (http.Handler, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("parse service url: %w", err)
	}

	return httputil.NewSingleHostReverseProxy(target), nil
}
