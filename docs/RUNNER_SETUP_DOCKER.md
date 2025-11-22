# GitHub Actions Runner Setup (Docker Compose)

This guide shows how to run the GitHub Actions self-hosted runner using Docker Compose alongside gizmatron.

## Prerequisites

- Raspberry Pi with Docker and Docker Compose installed
- GitHub Personal Access Token (PAT)

## Step 1: Create GitHub Personal Access Token

1. Go to: https://github.com/settings/tokens/new
2. Token name: `gizmatron-runner`
3. Select scopes:
   - ✅ **repo** (Full control of private repositories)
4. Click "Generate token"
5. **Copy the token** (you won't see it again!)

## Step 2: Set Up Environment Variables on Pi

SSH into your Raspberry Pi:
```bash
ssh ara@192.168.0.37
cd ~/gizmatron
```

Create `.env` file from template:
```bash
cp .env.example .env
nano .env
```

Edit the file and add your token:
```bash
GITHUB_RUNNER_TOKEN=ghp_YOUR_ACTUAL_TOKEN_HERE
```

Save and exit (Ctrl+X, Y, Enter).

## Step 3: Start the Runner

```bash
docker compose up -d github-runner
```

## Step 4: Verify Runner is Connected

Check the logs:
```bash
docker compose logs -f github-runner
```

You should see:
```
✓ Runner successfully added
✓ Connected to GitHub
```

Verify on GitHub:
1. Go to: https://github.com/arabenjamin/gizmatron/settings/actions/runners
2. You should see **gizmatron-pi** with a green "Idle" status

## Managing the Runner

**View logs:**
```bash
docker compose logs -f github-runner
```

**Restart runner:**
```bash
docker compose restart github-runner
```

**Stop runner:**
```bash
docker compose stop github-runner
```

**Remove runner:**
```bash
docker compose down github-runner
```

## Running the Full Stack

To run gizmatron and the runner together:
```bash
docker compose up -d
```

This starts:
- **gizmatron** - Main robotics service
- **github-runner** - CI/CD runner
- **twingate-connector** - VPN (if configured)

## Troubleshooting

### Runner shows "Offline"
```bash
# Check if container is running
docker compose ps

# Check logs for errors
docker compose logs github-runner

# Restart the runner
docker compose restart github-runner
```

### "Permission denied" errors in pipeline
```bash
# Ensure the runner user has Docker access
# The runner container already has /var/run/docker.sock mounted
# Check if it can run Docker:
docker compose exec github-runner docker ps
```

### Token expired
1. Generate a new PAT (Step 1)
2. Update `.env` file
3. Restart: `docker compose restart github-runner`

## How It Works

The runner container:
- Uses `myoung34/github-runner:latest` image (ARM64 compatible)
- Mounts Docker socket for running CI/CD jobs
- Registers with GitHub automatically using your PAT
- Has privileged access for hardware interaction
- Automatically restarts if it crashes

## Security Notes

- ⚠️ Never commit `.env` file (already in `.gitignore`)
- ⚠️ Use a token with minimal scope (repo only)
- ⚠️ Rotate token periodically
- ⚠️ The runner has Docker access (can run containers)
- ⚠️ Only run on trusted, private networks or behind VPN

## Pipeline Flow

Once the runner is set up:
1. Push code to GitHub
2. GitHub Actions triggers
3. Jobs that require `runs-on: self-hosted` execute on your Pi
4. Runner builds, deploys, and tests on actual hardware
5. Results reported back to GitHub

## Resources

- [GitHub Self-Hosted Runners Docs](https://docs.github.com/en/actions/hosting-your-own-runners)
- [myoung34/github-runner Image](https://github.com/myoung34/docker-github-actions-runner)
