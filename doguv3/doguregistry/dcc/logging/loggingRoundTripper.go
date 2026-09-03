package logging

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// LoggingRoundTripper intercepts HTTP requests to log metrics
type LoggingRoundTripper struct {
	originalTransport http.RoundTripper
}

func NewLoggingRoundTripper(transport http.RoundTripper) *LoggingRoundTripper {
	return &LoggingRoundTripper{
		originalTransport: transport,
	}
}

func (l *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	startTime := time.Now()

	// Ensure there is an original transport; throw an error when nil
	originalTransport := l.originalTransport
	if originalTransport == nil {
		slog.Error("Ensure there is an original Transport")
		return nil, fmt.Errorf("original transport is not provided to LoggingRoundTripper")
	}

	// Execute the actual HTTP request
	resp, err := originalTransport.RoundTrip(req)

	duration := time.Since(startTime)

	// remove sensitive fields from a copy of the URL
	safeURL := *req.URL
	safeURL.User = nil

	if err != nil {
		// Log the failure if the network request fails completely
		slog.Error("DCC Http Client Request failed: ",
			"Method", req.Method,
			"URL", safeURL.String(),
			"Duration", duration,
			"Error", err)
		return nil, err
	}

	// Log success with URL, status code, and execution time
	slog.Info("DCC Http Client Request information: ",
		"Method", req.Method,
		"URL", safeURL.String(),
		"Duration", duration,
		"Status", resp.StatusCode)
	return resp, nil
}
