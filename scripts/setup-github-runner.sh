#!/bin/bash
# GitHub Actions Self-Hosted Runner Setup Script
# For Raspberry Pi (Native Installation)
#
# This script installs and configures a GitHub Actions self-hosted runner
# on the Raspberry Pi for the gizmatron project.

set -e

echo "=========================================="
echo "GitHub Actions Runner Setup"
echo "For: gizmatron robotics project"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running on Raspberry Pi
echo "🔍 Checking system..."
if ! grep -q "Raspberry Pi\|BCM" /proc/cpuinfo 2>/dev/null; then
    echo -e "${YELLOW}⚠️  Warning: This doesn't appear to be a Raspberry Pi${NC}"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Detect architecture
ARCH=$(uname -m)
if [[ "$ARCH" == "aarch64" ]]; then
    RUNNER_ARCH="arm64"
    echo "✓ Detected ARM64 architecture"
elif [[ "$ARCH" == "armv7l" ]]; then
    RUNNER_ARCH="arm"
    echo "✓ Detected ARM32 architecture"
else
    echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"
    exit 1
fi

# Get GitHub token
echo ""
echo "📝 GitHub Personal Access Token Required"
echo ""
echo "To get your token:"
echo "  1. Go to: https://github.com/settings/tokens/new"
echo "  2. Token name: 'gizmatron-runner'"
echo "  3. Select scopes: ✅ repo (Full control of private repositories)"
echo "  4. Click 'Generate token'"
echo "  5. Copy the token (starts with 'ghp_')"
echo ""
read -p "Enter your GitHub token: " GITHUB_TOKEN

if [ -z "$GITHUB_TOKEN" ]; then
    echo -e "${RED}❌ Token is required${NC}"
    exit 1
fi

# Check if runner already exists
RUNNER_DIR="$HOME/actions-runner"
if [ -d "$RUNNER_DIR" ]; then
    echo ""
    echo -e "${YELLOW}⚠️  Runner directory already exists at: $RUNNER_DIR${NC}"
    read -p "Remove and reinstall? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "🗑️  Removing existing runner..."
        cd "$RUNNER_DIR"
        if [ -f "./svc.sh" ]; then
            sudo ./svc.sh stop 2>/dev/null || true
            sudo ./svc.sh uninstall 2>/dev/null || true
        fi
        cd "$HOME"
        rm -rf "$RUNNER_DIR"
        echo "✓ Removed existing runner"
    else
        echo "Exiting..."
        exit 0
    fi
fi

# Install dependencies
echo ""
echo "📦 Installing dependencies..."
sudo apt-get update -qq
sudo apt-get install -y curl jq git

# Install Go if not present
if ! command -v go &> /dev/null; then
    echo "📦 Installing Go..."
    GO_VERSION="1.23.5"
    if [[ "$RUNNER_ARCH" == "arm64" ]]; then
        GO_ARCH="arm64"
    else
        GO_ARCH="armv6l"
    fi
    wget -q https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-${GO_ARCH}.tar.gz
    rm go${GO_VERSION}.linux-${GO_ARCH}.tar.gz

    # Add to PATH if not already there
    if ! grep -q "/usr/local/go/bin" "$HOME/.bashrc"; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> "$HOME/.bashrc"
        export PATH=$PATH:/usr/local/go/bin
    fi
    echo "✓ Go installed: $(go version)"
else
    echo "✓ Go already installed: $(go version)"
fi

# Ensure user is in docker group
echo ""
echo "👤 Checking user permissions..."
if groups | grep -q docker; then
    echo "✓ User is in docker group"
else
    echo "Adding user to docker group..."
    sudo usermod -aG docker $USER
    echo -e "${YELLOW}⚠️  You'll need to log out and back in for docker group to take effect${NC}"
fi

# Add user to hardware access groups (best effort)
for group in i2c gpio video; do
    if getent group $group > /dev/null 2>&1; then
        sudo usermod -aG $group $USER 2>/dev/null || true
    fi
done

# Create runner directory
echo ""
echo "📁 Creating runner directory..."
mkdir -p "$RUNNER_DIR"
cd "$RUNNER_DIR"

# Get latest runner version
echo "🔍 Finding latest runner version..."
RUNNER_VERSION=$(curl -s https://api.github.com/repos/actions/runner/releases/latest | jq -r '.tag_name' | sed 's/v//')
echo "✓ Latest version: $RUNNER_VERSION"

# Download runner
echo ""
echo "⬇️  Downloading GitHub Actions runner..."
RUNNER_FILE="actions-runner-linux-${RUNNER_ARCH}-${RUNNER_VERSION}.tar.gz"
RUNNER_URL="https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/${RUNNER_FILE}"

if ! curl -o "$RUNNER_FILE" -L "$RUNNER_URL"; then
    echo -e "${RED}❌ Failed to download runner${NC}"
    exit 1
fi

echo "✓ Downloaded runner"

# Verify checksum (optional but recommended)
echo "🔐 Verifying checksum..."
curl -o "checksums.txt" -L "https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/checksums.txt"
if sha256sum -c checksums.txt 2>&1 | grep -q "$RUNNER_FILE: OK"; then
    echo "✓ Checksum verified"
else
    echo -e "${YELLOW}⚠️  Checksum verification failed (continuing anyway)${NC}"
fi

# Extract runner
echo ""
echo "📦 Extracting runner..."
tar xzf "$RUNNER_FILE"
rm "$RUNNER_FILE" checksums.txt
echo "✓ Extracted runner"

# Configure runner
echo ""
echo "⚙️  Configuring runner..."
echo "This will register the runner with GitHub..."
echo ""

./config.sh \
    --url https://github.com/arabenjamin/gizmatron \
    --token "$GITHUB_TOKEN" \
    --name "gizmatron-pi" \
    --labels "self-hosted,Linux,ARM64,raspberry-pi" \
    --work "_work" \
    --unattended \
    --replace

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Runner configured successfully${NC}"
else
    echo -e "${RED}❌ Runner configuration failed${NC}"
    exit 1
fi

# Install as systemd service
echo ""
echo "🔧 Installing runner as systemd service..."
sudo ./svc.sh install

if [ $? -eq 0 ]; then
    echo "✓ Service installed"
else
    echo -e "${RED}❌ Service installation failed${NC}"
    exit 1
fi

# Start service
echo ""
echo "▶️  Starting runner service..."
sudo ./svc.sh start

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Service started${NC}"
else
    echo -e "${RED}❌ Service start failed${NC}"
    exit 1
fi

# Check status
echo ""
echo "📊 Checking runner status..."
sleep 2
sudo ./svc.sh status

# Final instructions
echo ""
echo "=========================================="
echo -e "${GREEN}✅ GitHub Actions Runner Setup Complete!${NC}"
echo "=========================================="
echo ""
echo "Runner Details:"
echo "  Name: gizmatron-pi"
echo "  Location: $RUNNER_DIR"
echo "  Service: actions.runner.arabenjamin-gizmatron.gizmatron-pi.service"
echo ""
echo "Verify the runner is online:"
echo "  🌐 https://github.com/arabenjamin/gizmatron/settings/actions/runners"
echo "  You should see 'gizmatron-pi' with a green 'Idle' status"
echo ""
echo "Useful commands:"
echo "  View logs:    sudo journalctl -u actions.runner.* -f"
echo "  Stop runner:  cd $RUNNER_DIR && sudo ./svc.sh stop"
echo "  Start runner: cd $RUNNER_DIR && sudo ./svc.sh start"
echo "  Status:       cd $RUNNER_DIR && sudo ./svc.sh status"
echo "  Uninstall:    cd $RUNNER_DIR && sudo ./svc.sh uninstall"
echo ""
echo "Next steps:"
echo "  1. Verify runner shows as 'Idle' on GitHub"
echo "  2. Push a commit to trigger the CI/CD pipeline"
echo "  3. Watch the pipeline run on your Pi!"
echo ""

if groups | grep -q docker; then
    echo -e "${GREEN}✓ Setup complete - ready to run CI/CD jobs!${NC}"
else
    echo -e "${YELLOW}⚠️  IMPORTANT: Log out and back in for docker group to take effect${NC}"
fi
