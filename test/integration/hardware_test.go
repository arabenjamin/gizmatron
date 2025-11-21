package integration

import (
	"net/http"
	"testing"
	"time"
)

const (
	baseURL = "http://localhost:8080"
	timeout = 5 * time.Second
)

// TestAPIHealthCheck verifies the API server is responding
func TestAPIHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := &http.Client{Timeout: timeout}

	resp, err := client.Get(baseURL + "/ping")
	if err != nil {
		t.Fatalf("Failed to connect to API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestBotStatus verifies we can get robot status
func TestBotStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := &http.Client{Timeout: timeout}

	resp, err := client.Get(baseURL + "/api/v1/bot-status")
	if err != nil {
		t.Fatalf("Failed to get bot status: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestAPIEndpoints tests that all critical endpoints are accessible
func TestAPIEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	endpoints := []struct {
		name   string
		path   string
		method string
	}{
		{"Health Check", "/ping", "GET"},
		{"Bot Status", "/api/v1/bot-status", "GET"},
	}

	client := &http.Client{Timeout: timeout}

	for _, endpoint := range endpoints {
		t.Run(endpoint.name, func(t *testing.T) {
			var resp *http.Response
			var err error

			if endpoint.method == "GET" {
				resp, err = client.Get(baseURL + endpoint.path)
			} else {
				resp, err = client.Post(baseURL+endpoint.path, "application/json", nil)
			}

			if err != nil {
				t.Fatalf("Failed to access %s: %v", endpoint.path, err)
			}
			defer resp.Body.Close()

			// We accept 200 or 4xx/5xx (service is responding)
			// Just ensuring we get a response
			if resp.StatusCode == 0 {
				t.Errorf("No response from %s", endpoint.path)
			}
		})
	}
}

// TestServiceAvailability checks if the container is running and ports are accessible
func TestServiceAvailability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := &http.Client{Timeout: 2 * time.Second}

	// Just check if port 8080 is listening
	_, err := client.Get(baseURL + "/ping")
	if err != nil {
		t.Fatalf("Service not available on port 8080: %v", err)
	}
}
