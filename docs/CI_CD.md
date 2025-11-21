# Gizmatron CI/CD Pipeline

This document describes the complete CI/CD pipeline for Gizmatron, from code push to deployment and testing.

## Pipeline Overview

```
┌─────────────────────────────────────────────────────────────┐
│ Developer: Push to GitHub                                    │
└───────────────────┬─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────┐
│ Stage 1: Unit Tests (GitHub-hosted runner, Ubuntu)          │
│ • Checkout code                                              │
│ • Setup Go 1.23.5                                            │
│ • Run go vet (static analysis)                               │
│ • Run go fmt check (code formatting)                         │
│ • Run unit tests (-short flag, no hardware needed)           │
│ • Generate coverage report                                   │
│ • Upload coverage artifact                                   │
│                                                              │
│ ✓ Fast feedback (~1-2 minutes)                              │
│ ✓ No hardware required                                       │
│ ✓ Catches basic issues early                                │
└───────────────────┬─────────────────────────────────────────┘
                    │
                    │ ✓ Tests Pass
                    ▼
┌─────────────────────────────────────────────────────────────┐
│ Stage 2: Build (GitHub-hosted runner, Ubuntu)                │
│ • Checkout code                                              │
│ • Build Docker image                                         │
│ • Verify image was created                                   │
│ • Test image sanity (files exist)                            │
│                                                              │
│ ✓ Validates Docker build works                              │
│ ✓ Catches build-time errors                                 │
└───────────────────┬─────────────────────────────────────────┘
                    │
                    │ ✓ Build Succeeds
                    ▼
┌─────────────────────────────────────────────────────────────┐
│ Stage 3: Deploy (Self-hosted runner, Raspberry Pi)          │
│ • Checkout code                                              │
│ • Stop existing gizmatron container                          │
│ • Clean up dangling images                                   │
│ • Build Docker image on Pi (ARM64)                           │
│ • Start container with hardware access                       │
│ • Wait for service to start (60s timeout)                    │
│ • Verify container is running                                │
│                                                              │
│ ✓ Deploys to actual hardware                                │
│ ✓ Automatic rollout                                          │
│ ✓ Health checks ensure successful start                     │
└───────────────────┬─────────────────────────────────────────┘
                    │
                    │ ✓ Deployment Succeeds
                    ▼
┌─────────────────────────────────────────────────────────────┐
│ Stage 4: Integration Tests (Self-hosted runner, Pi)         │
│ • Checkout code                                              │
│ • Wait for service stabilization                             │
│ • Run integration tests                                      │
│   - API endpoint tests                                       │
│   - Hardware device detection                                │
│   - Component communication                                  │
│ • Check container health                                     │
│ • Collect logs                                               │
│                                                              │
│ ✓ Validates complete system                                 │
│ ✓ Tests hardware integration                                │
│ ✓ Verifies robot ecosystem                                  │
└───────────────────┬─────────────────────────────────────────┘
                    │
                    │ ✓ Integration Tests Pass
                    ▼
┌─────────────────────────────────────────────────────────────┐
│ Stage 5: Summary (GitHub-hosted runner)                     │
│ • Aggregate results from all stages                          │
│ • Report overall pipeline status                             │
│ • Fail if any stage failed                                   │
│                                                              │
│ ✓ Clear success/failure indication                          │
└─────────────────────────────────────────────────────────────┘
```

## Pipeline Stages Explained

### Stage 1: Unit Tests

**Runner**: GitHub-hosted (Ubuntu latest)
**Duration**: ~1-2 minutes
**Purpose**: Fast feedback on code quality

**Steps**:
1. **Code Quality Checks**:
   - `go vet`: Static analysis for common mistakes
   - `go fmt`: Ensures code follows Go formatting standards

2. **Unit Testing**:
   - Runs with `-short` flag (skips integration tests)
   - Runs with `-race` flag (detects race conditions)
   - Generates coverage report

3. **Coverage Report**:
   - Uploaded as artifact
   - Can be downloaded from GitHub Actions

**What It Tests**:
- Servo initialization logic
- API response formatting
- Request handling
- Data structures
- Business logic

**What It Doesn't Test**:
- Hardware communication
- Docker deployment
- Full system integration

### Stage 2: Build

**Runner**: GitHub-hosted (Ubuntu latest)
**Duration**: ~2-3 minutes
**Purpose**: Validate Docker image builds correctly

**Steps**:
1. Build Docker image
2. Verify image exists
3. Quick sanity check (files present)

**Why Separate from Deploy**:
- Fails fast if Docker build is broken
- Doesn't waste Pi resources on bad builds
- Validates build process before deployment

### Stage 3: Deploy

**Runner**: Self-hosted (Raspberry Pi)
**Duration**: ~3-5 minutes
**Purpose**: Deploy to actual hardware

**Steps**:
1. **Graceful Shutdown**:
   - Stops existing container (if running)
   - Preserves other containers on Pi
   - Cleans up dangling images only

2. **Build on Target**:
   - Builds ARM64 image on Pi
   - Uses `--no-cache` for clean build
   - Ensures latest code is deployed

3. **Start Container**:
   ```bash
   docker run -d \
     --name gizmatron \
     --privileged \
     --device /dev/i2c-1 \      # Servo controller
     --device /dev/video0 \     # Camera
     --device /dev/gpiomem \    # GPIO LEDs
     -p 8080:8080 \
     --restart unless-stopped \
     gizmatron:latest
   ```

4. **Health Check**:
   - Polls `/ping` endpoint for 60 seconds
   - Exits if service doesn't start
   - Shows logs on failure

**Automatic Restart**:
- `--restart unless-stopped` ensures service survives reboots

### Stage 4: Integration Tests

**Runner**: Self-hosted (Raspberry Pi)
**Duration**: ~1-2 minutes
**Purpose**: Validate complete system

**Tests Run**:
1. **API Availability**:
   - `/ping` responds
   - `/api/v1/bot-status` responds
   - Response times acceptable

2. **Device Detection**:
   - Camera device detected
   - Arm device detected
   - LED devices detected
   - Devices report operational status

3. **System Health**:
   - Container is running
   - Logs show no critical errors
   - All expected components initialized

**Test Output**:
- Detailed logs of device status
- Container health information
- Recent application logs

### Stage 5: Summary

**Runner**: GitHub-hosted (Ubuntu latest)
**Duration**: ~10 seconds
**Purpose**: Report overall status

Aggregates results from all previous stages and provides clear success/failure indication.

## Triggering the Pipeline

### Automatic Triggers

1. **Push to main branch**:
   ```bash
   git push origin main
   ```
   Full pipeline runs automatically

2. **Pull Request to main**:
   ```bash
   gh pr create --base main
   ```
   Full pipeline runs for validation

### Manual Trigger

From GitHub UI:
1. Go to Actions tab
2. Select "Gizmatron CI/CD" workflow
3. Click "Run workflow"
4. Select branch
5. Click "Run workflow"

Or via CLI:
```bash
gh workflow run "Gizmatron CI/CD"
```

## Viewing Pipeline Results

### GitHub UI

1. Go to repository → Actions tab
2. Click on a workflow run
3. View each job's logs
4. Download artifacts (coverage reports)

### Job Dependencies

The pipeline uses `needs` to enforce order:
```yaml
test → build → deploy → integration-test → summary
```

If any stage fails, dependent stages are skipped.

## Local Development Workflow

### Before Pushing

1. **Run unit tests locally**:
   ```bash
   go test -short -v ./...
   ```

2. **Check formatting**:
   ```bash
   go fmt ./...
   gofmt -s -w .
   ```

3. **Run static analysis**:
   ```bash
   go vet ./...
   ```

4. **Build Docker image**:
   ```bash
   docker build -t gizmatron:test .
   ```

### Testing on Pi

If you have SSH access to the Pi:

```bash
# SSH to Pi
ssh pi@gizmatron.local

# Pull latest code
cd ~/gizmatron
git pull

# Run integration tests
go test ./test/integration -v
```

## Troubleshooting Pipeline Failures

### Stage 1: Unit Tests Fail

**Common Issues**:
- Code formatting not applied: Run `go fmt ./...`
- Static analysis errors: Fix issues reported by `go vet`
- Test failures: Check test output, fix code or tests
- Race conditions: Review concurrent code

**Fix**:
```bash
# Format code
go fmt ./...

# Check for issues
go vet ./...

# Run tests locally
go test -short -v -race ./...
```

### Stage 2: Build Fails

**Common Issues**:
- Dockerfile syntax error
- Missing dependencies in Dockerfile
- Base image unavailable

**Fix**:
```bash
# Test build locally
docker build -t gizmatron:test .

# Check Dockerfile syntax
docker build --check -t gizmatron:test .
```

### Stage 3: Deploy Fails

**Common Issues**:
- Self-hosted runner offline
- Docker daemon issues on Pi
- Hardware devices not accessible
- Port 8080 already in use

**Check Runner Status**:
```bash
# SSH to Pi
ssh pi@gizmatron.local

# Check runner service
sudo systemctl status actions.runner.*

# Check Docker
sudo docker ps -a

# Check port
sudo netstat -tulpn | grep 8080
```

**Fix Runner**:
```bash
sudo systemctl restart actions.runner.*
```

**Fix Docker**:
```bash
# Restart Docker
sudo systemctl restart docker

# Check logs
sudo journalctl -u docker -n 50
```

### Stage 4: Integration Tests Fail

**Common Issues**:
- Service not fully started
- Hardware devices not connected
- Permissions issues
- Port conflicts

**Check Service**:
```bash
# SSH to Pi
ssh pi@gizmatron.local

# Check container
sudo docker ps | grep gizmatron

# Check logs
sudo docker logs gizmatron

# Test endpoint manually
curl http://localhost:8080/ping
curl http://localhost:8080/api/v1/bot-status
```

**Check Hardware**:
```bash
# List devices
ls -la /dev/i2c-* /dev/video* /dev/gpiomem

# Check permissions
groups

# Test I2C
sudo i2cdetect -y 1
```

## Rolling Back

If deployment fails or causes issues:

```bash
# SSH to Pi
ssh pi@gizmatron.local

# Stop current container
sudo docker stop gizmatron
sudo docker rm gizmatron

# Find previous image
sudo docker images gizmatron

# Start previous version
sudo docker run -d \
  --name gizmatron \
  --privileged \
  --device /dev/i2c-1 \
  --device /dev/video0 \
  --device /dev/gpiomem \
  -p 8080:8080 \
  gizmatron:PREVIOUS_VERSION
```

## Security Considerations

1. **Self-Hosted Runner Security**:
   - Runner has full access to Pi
   - Only trusted contributors should have push access
   - Review all PR changes carefully

2. **Docker Privileged Mode**:
   - Required for hardware access
   - Containers can access host system
   - Keep images minimal and audited

3. **Network Security**:
   - Pi may be exposed via Twingate VPN
   - Ensure strong SSH credentials
   - Monitor runner logs regularly

4. **Secrets Management**:
   - Never commit secrets to repository
   - Use GitHub Secrets for sensitive data
   - Rotate credentials regularly

## Monitoring and Logs

### GitHub Actions Logs

- Available for 90 days
- Download logs: Actions → Workflow Run → Download logs
- View specific step: Click on step name

### Pi Logs

```bash
# Container logs
sudo docker logs gizmatron

# Follow logs live
sudo docker logs -f gizmatron

# Runner logs
sudo journalctl -u actions.runner.* -f

# System logs
sudo journalctl -xe
```

## Performance Optimization

### Caching

The workflow currently doesn't use caching, but could be optimized:

```yaml
- uses: actions/cache@v3
  with:
    path: |
      ~/.cache/go-build
      ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
```

### Parallel Jobs

For large test suites, consider splitting tests:
```yaml
test-robot:
  run: go test ./robot -short -v

test-server:
  run: go test ./server -short -v
```

## Future Enhancements

1. **Notifications**:
   - Slack/Discord notifications on failure
   - Email alerts for critical failures

2. **Staging Environment**:
   - Deploy to staging before production
   - Smoke tests on staging

3. **Performance Tests**:
   - Measure API response times
   - Track memory usage
   - Monitor CPU utilization

4. **Security Scanning**:
   - Container vulnerability scanning
   - Dependency audit
   - Static security analysis

5. **Release Automation**:
   - Semantic versioning
   - Changelog generation
   - GitHub releases

## Related Documentation

- [Test Suite Documentation](../test/README.md)
- [Runner Setup Guide](./SETUP_RUNNER.md)
- [Main README](../README.md)
- [Camera Fix Strategy](../CAMERA_FIX_STRATEGY.md)
