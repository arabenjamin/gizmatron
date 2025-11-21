# GitHub Actions Self-Hosted Runner Setup

This guide explains how to set up and verify the GitHub Actions self-hosted runner on the Raspberry Pi for automated deployments.

## Why Self-Hosted Runner?

The CI/CD pipeline uses a self-hosted runner on the Raspberry Pi because:
1. **Hardware Access**: Integration tests need access to physical hardware (I2C, GPIO, camera)
2. **Local Deployment**: Deploys directly to the target device
3. **Network Access**: The Pi may be on a private network (via Twingate)

## Checking if Runner is Installed

SSH into your Raspberry Pi and check:

```bash
# Check if runner service is running
sudo systemctl status actions.runner.*

# OR check for runner process
ps aux | grep Runner.Listener

# Check if runner directory exists
ls -la ~/actions-runner/
```

## Installing the Runner

If the runner is not installed, follow these steps:

### 1. Go to GitHub Repository Settings

1. Navigate to your GitHub repository
2. Go to **Settings** → **Actions** → **Runners**
3. Click **New self-hosted runner**
4. Select **Linux** and **ARM64** (for Raspberry Pi)

### 2. SSH into Raspberry Pi

```bash
ssh pi@192.168.0.29
# OR
ssh ara@gizmatron.local
```

### 3. Download and Extract Runner

```bash
# Create a folder
mkdir -p ~/actions-runner && cd ~/actions-runner

# Download the latest runner package (check GitHub for current version)
curl -o actions-runner-linux-arm64-2.311.0.tar.gz -L \
  https://github.com/actions/runner/releases/download/v2.311.0/actions-runner-linux-arm64-2.311.0.tar.gz

# Extract the installer
tar xzf ./actions-runner-linux-arm64-2.311.0.tar.gz
```

### 4. Configure the Runner

```bash
# Create the runner and start the configuration
./config.sh --url https://github.com/YOUR_USERNAME/gizmatron \
  --token YOUR_TOKEN_FROM_GITHUB

# Follow the prompts:
# - Enter runner name: gizmatron-pi (or similar)
# - Enter runner labels: self-hosted,Linux,ARM64,raspberry-pi
# - Enter work folder: _work (default)
```

**Note**: Replace `YOUR_USERNAME` and `YOUR_TOKEN_FROM_GITHUB` with values from the GitHub UI.

### 5. Install as a Service

```bash
# Install the service
sudo ./svc.sh install

# Start the service
sudo ./svc.sh start

# Check status
sudo ./svc.sh status
```

### 6. Verify Runner is Connected

1. Go to GitHub → Settings → Actions → Runners
2. You should see your runner listed as "Idle" (green dot)

## Managing the Runner Service

```bash
# Start the runner
sudo systemctl start actions.runner.*

# Stop the runner
sudo systemctl stop actions.runner.*

# Restart the runner
sudo systemctl restart actions.runner.*

# View logs
sudo journalctl -u actions.runner.* -f

# Check status
sudo systemctl status actions.runner.*
```

## Running in Docker (Alternative)

You can also run the runner in Docker:

```bash
# Create docker-compose.yml for runner
cat > ~/runner-compose.yml << 'EOF'
version: '3.8'
services:
  runner:
    image: myoung34/github-runner:latest
    environment:
      - RUNNER_NAME=gizmatron-pi-docker
      - RUNNER_WORKDIR=/tmp/runner
      - RUNNER_SCOPE=repo
      - REPO_URL=https://github.com/YOUR_USERNAME/gizmatron
      - ACCESS_TOKEN=YOUR_PERSONAL_ACCESS_TOKEN
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    restart: unless-stopped
EOF

# Start the runner
docker compose -f ~/runner-compose.yml up -d
```

## Troubleshooting

### Runner Shows as Offline

```bash
# Check if service is running
sudo systemctl status actions.runner.*

# Restart the service
sudo systemctl restart actions.runner.*

# Check logs for errors
sudo journalctl -u actions.runner.* --no-pager -n 100
```

### Runner Can't Access Docker

The runner needs permission to use Docker:

```bash
# Add runner user to docker group
sudo usermod -aG docker $(whoami)

# Restart runner service
sudo systemctl restart actions.runner.*
```

### Runner Can't Access Hardware Devices

The runner user needs access to hardware:

```bash
# Add to necessary groups
sudo usermod -aG i2c,gpio,video $(whoami)

# Create udev rules for device access
sudo nano /etc/udev/rules.d/99-gizmatron.rules
```

Add:
```
KERNEL=="i2c-[0-9]*", GROUP="i2c", MODE="0660"
KERNEL=="gpiomem", GROUP="gpio", MODE="0660"
SUBSYSTEM=="video4linux", GROUP="video", MODE="0660"
```

Then:
```bash
# Reload udev rules
sudo udevadm control --reload-rules
sudo udevadm trigger

# Restart runner
sudo systemctl restart actions.runner.*
```

### Runner Workflow Fails with Permission Errors

If you see "permission denied" errors in workflows:

```bash
# Check runner user
ps aux | grep Runner.Listener

# Verify docker permissions
sudo -u RUNNER_USER docker ps

# Verify hardware device permissions
ls -la /dev/i2c-* /dev/video* /dev/gpiomem
```

## Testing the Runner

Create a test workflow to verify the runner works:

```yaml
# .github/workflows/test-runner.yml
name: Test Self-Hosted Runner

on: workflow_dispatch

jobs:
  test:
    runs-on: self-hosted
    steps:
      - name: Check environment
        run: |
          echo "Hostname: $(hostname)"
          echo "User: $(whoami)"
          echo "Docker: $(docker --version)"
          echo "Go: $(go version)"

      - name: Check hardware access
        run: |
          echo "I2C devices:"
          ls -la /dev/i2c-* || echo "No I2C devices"
          echo "Video devices:"
          ls -la /dev/video* || echo "No video devices"
          echo "GPIO:"
          ls -la /dev/gpiomem || echo "No GPIO access"
```

Run this workflow manually from GitHub Actions tab to verify everything works.

## Security Considerations

1. **Use PAT with minimal scope**: Only grant repo access
2. **Use runner groups**: Limit which workflows can use the runner
3. **Regular updates**: Keep the runner software updated
4. **Network security**: Use Twingate VPN for secure access
5. **Monitor logs**: Regularly check runner logs for suspicious activity

## Updating the Runner

```bash
cd ~/actions-runner

# Stop the service
sudo ./svc.sh stop

# Download new version
curl -o actions-runner-linux-arm64-VERSION.tar.gz -L \
  https://github.com/actions/runner/releases/download/vVERSION/actions-runner-linux-arm64-VERSION.tar.gz

# Extract
tar xzf ./actions-runner-linux-arm64-VERSION.tar.gz

# Start the service
sudo ./svc.sh start
```

## Resources

- [GitHub Self-Hosted Runners Documentation](https://docs.github.com/en/actions/hosting-your-own-runners)
- [Raspberry Pi Setup Guide](https://github.com/actions/runner/blob/main/docs/start/envlinux.md)
