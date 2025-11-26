# GitHub Actions Runner Setup (Native Installation)

This guide shows how to install the GitHub Actions self-hosted runner natively on the Raspberry Pi.

**Why Native Installation?**
- ✅ Official GitHub approach (fully supported)
- ✅ Better security (no Docker-in-Docker)
- ✅ Simpler troubleshooting
- ✅ Direct hardware access for robotics
- ✅ Lower overhead

## Quick Start (Automated Setup)

SSH into your Raspberry Pi and run:

```bash
ssh ara@192.168.0.37
cd ~/gizmatron
git pull
./scripts/setup-github-runner.sh
```

The script will:
1. ✓ Check system compatibility
2. ✓ Install dependencies (Go, git, etc.)
3. ✓ Download the latest runner
4. ✓ Register with GitHub
5. ✓ Install as systemd service
6. ✓ Start the runner

## Prerequisites

Before running the setup script:

### 1. Create GitHub Personal Access Token

1. Go to: https://github.com/settings/tokens/new
2. Token name: `gizmatron-runner`
3. Expiration: 90 days (or No expiration)
4. Select scopes:
   - ✅ **repo** (Full control of private repositories)
5. Click "Generate token"
6. **Copy the token** - it starts with `ghp_` and you won't see it again!

Keep this token ready - the setup script will ask for it.

## Manual Installation (If Needed)

If the automated script doesn't work, follow these steps:

### Step 1: Install Dependencies

```bash
sudo apt-get update
sudo apt-get install -y curl jq git
```

### Step 2: Download Runner

```bash
# Create runner directory
mkdir -p ~/actions-runner && cd ~/actions-runner

# Download latest runner for ARM64
curl -o actions-runner-linux-arm64-2.311.0.tar.gz -L \
  https://github.com/actions/runner/releases/download/v2.311.0/actions-runner-linux-arm64-2.311.0.tar.gz

# Extract
tar xzf ./actions-runner-linux-arm64-2.311.0.tar.gz
```

### Step 3: Configure Runner

```bash
./config.sh \
  --url https://github.com/arabenjamin/gizmatron \
  --token YOUR_GITHUB_TOKEN \
  --name gizmatron-pi \
  --labels self-hosted,Linux,ARM64,raspberry-pi \
  --work _work \
  --unattended \
  --replace
```

### Step 4: Install as Service

```bash
# Install service
sudo ./svc.sh install

# Start service
sudo ./svc.sh start

# Check status
sudo ./svc.sh status
```

## Verify Installation

### Check Service Status

```bash
cd ~/actions-runner
sudo ./svc.sh status
```

You should see:
```
● actions.runner.arabenjamin-gizmatron.gizmatron-pi.service
   Active: active (running)
```

### View Logs

```bash
# Real-time logs
sudo journalctl -u actions.runner.* -f

# Last 50 lines
sudo journalctl -u actions.runner.* -n 50
```

### Check on GitHub

1. Go to: https://github.com/arabenjamin/gizmatron/settings/actions/runners
2. You should see **gizmatron-pi** with:
   - ✅ Green "Idle" status
   - Labels: self-hosted, Linux, ARM64, raspberry-pi

## Managing the Runner

### Service Commands

```bash
cd ~/actions-runner

# Start
sudo ./svc.sh start

# Stop
sudo ./svc.sh stop

# Restart
sudo ./svc.sh restart

# Status
sudo ./svc.sh status

# Uninstall
sudo ./svc.sh uninstall
```

### Update Runner

```bash
cd ~/actions-runner

# Stop service
sudo ./svc.sh stop

# Download new version
# (Check https://github.com/actions/runner/releases)
curl -o actions-runner-linux-arm64-VERSION.tar.gz -L \
  https://github.com/actions/runner/releases/download/vVERSION/actions-runner-linux-arm64-VERSION.tar.gz

# Extract (overwrites old files)
tar xzf ./actions-runner-linux-arm64-VERSION.tar.gz

# Restart service
sudo ./svc.sh start
```

### Remove Runner

```bash
cd ~/actions-runner

# Stop and uninstall
sudo ./svc.sh stop
sudo ./svc.sh uninstall

# Remove directory
cd ~
rm -rf actions-runner
```

Then remove it from GitHub:
https://github.com/arabenjamin/gizmatron/settings/actions/runners

## Troubleshooting

### Runner shows "Offline" on GitHub

```bash
# Check if service is running
sudo systemctl status actions.runner.*

# Restart service
cd ~/actions-runner
sudo ./svc.sh restart

# Check logs for errors
sudo journalctl -u actions.runner.* -n 100
```

### "Permission denied" accessing Docker

```bash
# Verify user is in docker group
groups

# If not, add user to docker group
sudo usermod -aG docker $USER

# Log out and back in for changes to take effect
```

### "Permission denied" accessing hardware (I2C, GPIO)

```bash
# Add user to hardware groups
sudo usermod -aG i2c,gpio,video $USER

# Log out and back in
```

### Runner keeps crashing

```bash
# Check system resources
free -h
df -h

# Check logs for errors
sudo journalctl -u actions.runner.* -n 200

# Try running manually to see errors
cd ~/actions-runner
./run.sh
```

### Token expired

If you see "HTTP 401 Unauthorized":

1. Generate a new token (same as Step 1 above)
2. Reconfigure runner:
```bash
cd ~/actions-runner
sudo ./svc.sh stop
./config.sh remove --token OLD_TOKEN
./config.sh --url https://github.com/arabenjamin/gizmatron --token NEW_TOKEN --name gizmatron-pi --unattended
sudo ./svc.sh start
```

## Security Best Practices

1. **Token Management**
   - Use tokens with minimal scope (repo only)
   - Set expiration dates (90 days)
   - Rotate tokens regularly
   - Never commit tokens to git

2. **Runner Security**
   - Only run on trusted networks or behind VPN (Twingate)
   - Keep Pi OS updated: `sudo apt-get update && sudo apt-get upgrade`
   - Monitor runner logs for suspicious activity
   - The runner has full access to your repository code

3. **Network Security**
   - Use Twingate VPN for remote access
   - Don't expose runner to internet directly
   - Consider firewall rules: `sudo ufw status`

## How It Works

### Pipeline Flow

1. You push code to GitHub
2. GitHub Actions triggers workflow
3. Jobs with `runs-on: self-hosted` are assigned to your runner
4. Runner on Pi:
   - Checks out code
   - Runs build steps
   - Executes tests
   - Has access to real hardware (I2C, GPIO, camera)
   - Reports results back to GitHub

### Directory Structure

```
~/actions-runner/
├── config.sh           # Configuration script
├── run.sh             # Run runner manually
├── svc.sh             # Service management
├── bin/               # Runner binaries
├── externals/         # Dependencies
└── _work/             # Job workspace
    └── gizmatron/     # Your repo (during jobs)
```

### Service Details

- **Service Name**: `actions.runner.arabenjamin-gizmatron.gizmatron-pi.service`
- **User**: Runs as your user (e.g., `ara`)
- **Auto-start**: Starts on boot
- **Restart**: Automatically restarts if it crashes

## Testing the Pipeline

After setup, test the complete pipeline:

1. Make a small change to code:
```bash
cd ~/gizmatron
echo "# Test" >> README.md
git add README.md
git commit -S -m "Test CI/CD pipeline"
git push origin main
```

2. Watch on GitHub:
   - https://github.com/arabenjamin/gizmatron/actions
   - You should see the workflow run through all stages

3. Monitor on Pi:
```bash
# Watch runner logs
sudo journalctl -u actions.runner.* -f

# Or watch job execution
watch -n 1 'ls -la ~/actions-runner/_work/gizmatron/gizmatron/'
```

## Resources

- [GitHub Self-Hosted Runners Documentation](https://docs.github.com/en/actions/hosting-your-own-runners)
- [Runner Releases](https://github.com/actions/runner/releases)
- [Self-Hosted Runner Security](https://docs.github.com/en/actions/hosting-your-own-runners/managing-self-hosted-runners/about-self-hosted-runners#self-hosted-runner-security)
