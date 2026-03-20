package handlers

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var openapi string

// RobotsTXTHandler handles the robots.txt endpoint.
func RobotsTXTHandler(w http.ResponseWriter, _ *http.Request) {
	const robotsTxt = `User-agent: *
Allow: /v1/openapi.yaml
Disallow: /`
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(robotsTxt))
}

// OpenAPIHandler handles the OpenAPI spec endpoint.
func OpenAPIHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/yaml")
	_, _ = w.Write([]byte(openapi))
}
