package consts

import "time"

// HTTP server timeout constants.
const (
	// ReadTimeout is the maximum duration for reading the entire request, including the body.
	ReadTimeout = 5 * time.Second
	// WriteTimeout is the maximum duration before timing out writes of the response.
	WriteTimeout = 10 * time.Second
	// IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled.
	IdleTimeout = 15 * time.Second
)

// RequestIDFormat is the generated request ID format (YYYYMMDD-HHMMSS.mmm).
const RequestIDFormat = "20060102-150405.000"
