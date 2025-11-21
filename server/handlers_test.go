package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPing(t *testing.T) {
	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ping)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Could not parse response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", response["status"])
	}

	if response["message"] != "pong!" {
		t.Errorf("Expected message 'pong!', got '%v'", response["message"])
	}

	// Check that this_request exists
	if _, ok := response["this_request"]; !ok {
		t.Error("Response missing 'this_request' field")
	}
}

func TestClientHash(t *testing.T) {
	req1, _ := http.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	req1.Header.Set("User-Agent", "TestAgent/1.0")

	req2, _ := http.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	req2.Header.Set("User-Agent", "TestAgent/1.0")

	hash1 := clientHash(req1)
	hash2 := clientHash(req2)

	if hash1 != hash2 {
		t.Error("Same client should produce same hash")
	}

	// Different user agent should produce different hash
	req3, _ := http.NewRequest("GET", "/", nil)
	req3.RemoteAddr = "192.168.1.1:12345"
	req3.Header.Set("User-Agent", "DifferentAgent/1.0")

	hash3 := clientHash(req3)

	if hash1 == hash3 {
		t.Error("Different user agents should produce different hashes")
	}
}

func TestPingResponseStructure(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ping", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("User-Agent", "TestAgent/1.0")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ping)
	handler.ServeHTTP(rr, req)

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Could not parse response: %v", err)
	}

	// Check required fields
	requiredFields := []string{"status", "message", "this_request"}
	for _, field := range requiredFields {
		if _, ok := response[field]; !ok {
			t.Errorf("Response missing required field: %s", field)
		}
	}

	// Check this_request structure
	thisRequest, ok := response["this_request"].(map[string]interface{})
	if !ok {
		t.Fatal("this_request is not a map")
	}

	requestFields := []string{"time", "client_address", "resource", "user_agent", "client"}
	for _, field := range requestFields {
		if _, ok := thisRequest[field]; !ok {
			t.Errorf("this_request missing required field: %s", field)
		}
	}
}

func TestContentType(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ping", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ping)

	handler.ServeHTTP(rr, req)

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
}
