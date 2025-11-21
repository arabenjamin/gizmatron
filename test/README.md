# Gizmatron Test Suite

This directory contains the test suite for Gizmatron, organized into unit tests and integration tests.

## Test Organization

### Unit Tests
Unit tests are located alongside the code they test:
- `robot/*_test.go` - Tests for robot hardware abstraction layer
- `server/*_test.go` - Tests for HTTP server and handlers

Unit tests:
- Do NOT require hardware
- Do NOT require the service to be running
- Test individual functions and components in isolation
- Run fast (< 1 second)

### Integration Tests
Integration tests are in `test/integration/`:
- `hardware_test.go` - API endpoint availability tests
- `device_test.go` - Hardware device initialization tests

Integration tests:
- REQUIRE the gizmatron service to be running on localhost:8080
- REQUIRE hardware to be connected (for full validation)
- Test the complete system end-to-end
- May run slower due to network calls

## Running Tests

### Run All Unit Tests
```bash
# From gizmatron directory
go test ./robot ./server -v
```

### Run Unit Tests Only (Fast)
```bash
go test -short ./...
```

### Run Integration Tests (Service Must Be Running)
```bash
# Ensure gizmatron is running first:
docker compose up -d

# Then run integration tests:
go test ./test/integration -v

# Or run all tests including integration:
go test ./... -v
```

### Run Specific Test
```bash
# Run specific test by name
go test ./robot -run TestNewServo -v

# Run integration test for devices
go test ./test/integration -run TestDeviceInitialization -v
```

### Run Tests with Coverage
```bash
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## CI/CD Integration

### GitHub Actions Workflow

The CI/CD pipeline runs tests in two stages:

**1. Test Stage (GitHub-hosted runner)**
- Runs unit tests only (`-short` flag)
- Fast feedback on code quality
- No hardware required
- Must pass before deployment

**2. Deploy Stage (Self-hosted Raspberry Pi)**
- Deploys to Raspberry Pi
- Runs integration tests
- Verifies hardware components
- Tests complete system functionality

### Test Flags

- `-short` - Skip integration tests (for CI)
- `-v` - Verbose output
- `-cover` - Show code coverage
- `-race` - Enable race detector (unit tests only)

## Writing New Tests

### Unit Test Example
```go
// robot/example_test.go
package robot

import "testing"

func TestExample(t *testing.T) {
    // Arrange
    servo := NewServo(true, 0, 10.3)

    // Act
    result := servo.SomeFunction()

    // Assert
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### Integration Test Example
```go
// test/integration/example_test.go
package integration

import "testing"

func TestExample(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Test code that requires running service
    resp, err := http.Get("http://localhost:8080/ping")
    // ...
}
```

## Test Best Practices

1. **Isolation**: Unit tests should not depend on external systems
2. **Naming**: Use descriptive test names (TestFunctionName_Scenario_ExpectedBehavior)
3. **Table Tests**: Use table-driven tests for multiple scenarios
4. **Cleanup**: Use `defer` for cleanup in tests
5. **Short Flag**: Mark integration tests with `if testing.Short()`
6. **Clear Failures**: Provide clear error messages explaining what failed

## Troubleshooting

### Integration Tests Failing
- Ensure gizmatron service is running: `docker ps`
- Check service is accessible: `curl http://localhost:8080/ping`
- Verify port 8080 is not blocked by firewall

### Build Failures
- Run `go mod tidy` to sync dependencies
- Check for syntax errors: `go vet ./...`
- Format code: `go fmt ./...`

### Hardware Tests Failing on Pi
- Verify devices are connected: `ls /dev/i2c-* /dev/video*`
- Check permissions: User must have access to hardware devices
- Run with privileges: `sudo` may be required for hardware access
