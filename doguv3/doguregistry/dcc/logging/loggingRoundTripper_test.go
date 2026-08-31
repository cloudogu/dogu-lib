package logging

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggingRoundTripper_LogsRequestData(t *testing.T) {
	// 1. Setup a local mock server to avoid hitting the live internet
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot) // 418 I'm a teapot (Distinctive status to look for)
	}))
	defer mockServer.Close()

	// 2. Capture slog output using a local bytes buffer
	var logBuffer bytes.Buffer

	// Create a test logger.
	// Switch to slog.NewJSONHandler if your production app outputs JSON lines.
	testLogger := slog.New(slog.NewTextHandler(&logBuffer, nil))

	// Set the test logger as default so your LoggingRoundTripper captures it
	slog.SetDefault(testLogger)

	// 3. Initialize your exact LoggingRoundTripper injecting the mock transport
	client := &http.Client{
		Transport: &LoggingRoundTripper{
			Proxied: http.DefaultTransport,
		},
	}

	// 4. Fire the HTTP request against the mock server
	resp, err := client.Get(mockServer.URL)
	if err != nil {
		t.Fatalf("Failed to execute request through LoggingRoundTripper: %v", err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			errClose := resp.Body.Close()
			if errClose != nil {
				slog.Error("failed to close body: ", "error", errClose)
			}
		}
	}()

	// 5. Assertions on the generated text logs
	logOutput := logBuffer.String()

	logSuccess := "DCC Http Client Request information: "
	if !strings.Contains(logOutput, logSuccess) {
		t.Errorf("Expected logs to contain logSuccess message %q, but got:\n%s", logSuccess, logOutput)
	}
	// Verify that the requested URL path is recorded
	if !strings.Contains(logOutput, mockServer.URL) {
		t.Errorf("Expected logs to contain URL %q, but got:\n%s", mockServer.URL, logOutput)
	}

	// Verify that the HTTP status code is recorded
	if !strings.Contains(logOutput, "418") {
		t.Errorf("Expected logs to contain status code 418, but got:\n%s", logOutput)
	}

	// Verify that the HTTP request method is recorded
	if !strings.Contains(logOutput, "GET") {
		t.Errorf("Expected logs to contain HTTP method 'GET', but got:\n%s", logOutput)
	}

	// Verify that the HTTP request method is recorded
	if !strings.Contains(logOutput, "Duration") {
		t.Errorf("Expected logs to contain Duration, but got:\n%s", logOutput)
	}
}
func TestLoggingRoundTripper_LogsErrorOnFailure(t *testing.T) {
	// 1. Capture slog output using a local bytes buffer
	var logBuffer bytes.Buffer
	testLogger := slog.New(slog.NewTextHandler(&logBuffer, nil))
	slog.SetDefault(testLogger)

	// 2. Initialize your exact LoggingRoundTripper
	client := &http.Client{
		Transport: &LoggingRoundTripper{
			Proxied: http.DefaultTransport,
		},
	}

	// 3. Create a request with an already canceled context to force a RoundTrip error
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 👈 Cancel immediately!

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot) // 418 I'm a teapot (Distinctive status to look for)
	}))
	defer mockServer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mockServer.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// 4. Fire the request (this will fail instantly because the context is canceled)
	_, err = client.Do(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	// 5. Assertions on the generated error logs
	logOutput := logBuffer.String()

	// Verify that the log captured the standard context cancellation text

	expectedErrorString := "context canceled"
	if !strings.Contains(logOutput, expectedErrorString) {
		t.Errorf("Expected logs to contain error message %q, but got:\n%s", expectedErrorString, logOutput)
	}

	logError := "DCC Http Client Request failed: "
	if !strings.Contains(logOutput, logError) {
		t.Errorf("Expected logs to contain logError message %q, but got:\n%s", logError, logOutput)
	}
	// Verify that the log still tracks metadata like the HTTP method
	if !strings.Contains(logOutput, "GET") {
		t.Errorf("Expected logs to contain HTTP method 'GET', but got:\n%s", logOutput)
	}

	// Verify that the requested URL path is recorded
	if !strings.Contains(logOutput, mockServer.URL) {
		t.Errorf("Expected logs to contain URL %q, but got:\n%s", mockServer.URL, logOutput)
	}

	// Verify that the HTTP request method is recorded
	if !strings.Contains(logOutput, "Duration") {
		t.Errorf("Expected logs to contain Duration, but got:\n%s", logOutput)
	}

	// Verify that the HTTP request method is recorded
	if !strings.Contains(logOutput, "Error") {
		t.Errorf("Expected logs to contain Error, but got:\n%s", logOutput)
	}
}

func TestLoggingRoundTripper_LogsErrorOnEmptyTransport(t *testing.T) {
	// 1. Capture slog output using a local bytes buffer
	var logBuffer bytes.Buffer
	testLogger := slog.New(slog.NewTextHandler(&logBuffer, nil))
	slog.SetDefault(testLogger)

	// 2. Initialize your exact LoggingRoundTripper
	client := &http.Client{
		Transport: &LoggingRoundTripper{
			Proxied: nil,
		},
	}

	// 3. Create a request with an already canceled context to force a RoundTrip error
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 👈 Cancel immediately!

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot) // 418 I'm a teapot (Distinctive status to look for)
	}))
	defer mockServer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mockServer.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// 4. Fire the request (this will fail instantly because the context is canceled)
	resp, err := client.Do(req)
	defer func() {
		if resp != nil && resp.Body != nil {
			errClose := resp.Body.Close()
			if errClose != nil {
				slog.Error("failed to close body: ", "error", errClose)
			}
		}
	}()

	// 5. Assertions on the generated error logs
	logOutput := logBuffer.String()

	// Verify that the log captured the standard context cancellation text

	expectedErrorString := "DCC Http Client Request failed"
	assert.Contains(t, logOutput, expectedErrorString)

}
