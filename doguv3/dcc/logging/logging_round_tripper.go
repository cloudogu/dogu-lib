package logging

import (
	"log/slog"
	"net/http"
	"time"
)

// LoggingRoundTripper intercepts HTTP requests to log metrics
type LoggingRoundTripper struct {
	Proxied http.RoundTripper
}

func (l *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	startTime := time.Now()

	// Execute the actual HTTP request
	resp, err := l.Proxied.RoundTrip(req)

	duration := time.Since(startTime)

	if err != nil {
		// Log the failure if the network request fails completely
		slog.Error("DCC Http Client Request failed: ",
			"Method", req.Method,
			"URL", req.URL.String(),
			"Duration", duration,
			"Error", err)
		return nil, err
	}

	// Log success with URL, status code, and execution time
	slog.Info("DCC Http Client Request information: ",
		"Method", req.Method,
		"URL", req.URL.String(),
		"Duration", duration,
		"Status", resp.StatusCode)
	return resp, nil
}
