package proxy

import "net/http"

// NewFarmProxy forwards /api/devices and /api/plants requests to the farm service.
func NewFarmProxy(targetURL string) (http.Handler, error) {
	return NewReverseProxy(targetURL)
}
