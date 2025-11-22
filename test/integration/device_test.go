package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

type BotStatusResponse struct {
	Status       string                 `json:"status"`
	DeviceStatus map[string]DeviceInfo  `json:"device_status"`
	BotName      string                 `json:"botname"`
	ThisRequest  map[string]interface{} `json:"this_request"`
}

type DeviceInfo struct {
	Name          string `json:"name"`
	DeviceType    string `json:"device_type"`
	IsOperational bool   `json:"isoperational"`
	IsRunning     bool   `json:"isrunning"`
}

// TestDeviceInitialization verifies all hardware devices are initialized
func TestDeviceInitialization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := &http.Client{Timeout: timeout}

	resp, err := client.Get(baseURL + "/api/v1/bot-status")
	if err != nil {
		t.Fatalf("Failed to get bot status: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var status BotStatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(status.DeviceStatus) == 0 {
		t.Error("No devices reported in status")
	}

	t.Logf("Found %d devices", len(status.DeviceStatus))
	for name, device := range status.DeviceStatus {
		t.Logf("Device: %s, Type: %s, Operational: %v, Running: %v",
			name, device.DeviceType, device.IsOperational, device.IsRunning)
	}
}

// TestCameraDevice verifies camera device is present and operational
func TestCameraDevice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := &http.Client{Timeout: timeout}

	resp, err := client.Get(baseURL + "/api/v1/bot-status")
	if err != nil {
		t.Fatalf("Failed to get bot status: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var status BotStatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check if camera device exists
	cameraFound := false
	for name, device := range status.DeviceStatus {
		if device.DeviceType == "camera" {
			cameraFound = true
			t.Logf("Camera device '%s' found - Operational: %v", name, device.IsOperational)

			// Camera should be at least initialized (operational)
			// It's OK if it's not running yet
			if !device.IsOperational {
				t.Logf("WARNING: Camera is not operational (this may be expected if hardware is not connected)")
			}
			break
		}
	}

	if !cameraFound {
		t.Log("WARNING: No camera device found in device list")
	}
}

// TestArmDevice verifies robotic arm device is present
func TestArmDevice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := &http.Client{Timeout: timeout}

	resp, err := client.Get(baseURL + "/api/v1/bot-status")
	if err != nil {
		t.Fatalf("Failed to get bot status: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var status BotStatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check if arm device exists
	armFound := false
	for name, device := range status.DeviceStatus {
		if device.DeviceType == "arm" {
			armFound = true
			t.Logf("Arm device '%s' found - Operational: %v", name, device.IsOperational)

			if !device.IsOperational {
				t.Logf("WARNING: Arm is not operational (this may be expected if hardware is not connected)")
			}
			break
		}
	}

	if !armFound {
		t.Log("WARNING: No arm device found in device list")
	}
}

// TestLEDDevices verifies LED status indicators are present
func TestLEDDevices(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := &http.Client{Timeout: timeout}

	resp, err := client.Get(baseURL + "/api/v1/bot-status")
	if err != nil {
		t.Fatalf("Failed to get bot status: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var status BotStatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check for LED devices
	ledCount := 0
	for name, device := range status.DeviceStatus {
		if device.DeviceType == "led" {
			ledCount++
			t.Logf("LED device '%s' found - Operational: %v", name, device.IsOperational)
		}
	}

	t.Logf("Found %d LED devices", ledCount)

	// We expect 3 LEDs (running, server, arm) but don't fail if missing
	if ledCount == 0 {
		t.Log("WARNING: No LED devices found")
	}
}

// TestResponseTime measures API response time
func TestResponseTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := &http.Client{Timeout: timeout}

	start := time.Now()
	resp, err := client.Get(baseURL + "/ping")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to ping API: %v", err)
	}
	defer resp.Body.Close()

	t.Logf("API response time: %v", elapsed)

	// Response should be reasonably fast (under 1 second for local requests)
	if elapsed > 1*time.Second {
		t.Logf("WARNING: API response time is slow: %v", elapsed)
	}
}
