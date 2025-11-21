#!/bin/bash
# migrate-camera.sh - Migration script for camera backend transition

set -e

echo "=== Gizmatron Camera Migration Script ==="
echo "This script helps migrate from V4L2 to libcamera support"
echo

# Check if running on Raspberry Pi
if ! grep -q "Raspberry Pi" /proc/device-tree/model 2>/dev/null; then
    echo "Warning: This script is designed for Raspberry Pi systems"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Function to check command availability
check_command() {
    if command -v "$1" >/dev/null 2>&1; then
        echo "✓ $1 is available"
        return 0
    else
        echo "✗ $1 is not available"
        return 1
    fi
}

# Function to check file existence
check_file() {
    if [ -e "$1" ]; then
        echo "✓ $1 exists"
        return 0
    else
        echo "✗ $1 does not exist"
        return 1
    fi
}

echo "=== System Analysis ==="

# Check OS version
if [ -f /etc/os-release ]; then
    . /etc/os-release
    echo "OS: $PRETTY_NAME"
    echo "Version: $VERSION"
else
    echo "Cannot determine OS version"
fi

# Check kernel version
echo "Kernel: $(uname -r)"

echo
echo "=== Camera System Analysis ==="

# Check for libcamera
echo "Checking libcamera availability:"
check_command "libcamera-hello"
check_command "libcamera-vid"
check_command "libcamera-still"

echo
echo "Checking V4L2 compatibility:"
check_file "/dev/video0"
check_command "v4l2-ctl"

# Check if V4L2 module is loaded
if lsmod | grep -q bcm2835_v4l2; then
    echo "✓ bcm2835-v4l2 module is loaded"
else
    echo "✗ bcm2835-v4l2 module is not loaded"
fi

echo
echo "Checking GStreamer:"
check_command "gst-launch-1.0"
check_command "gst-inspect-1.0"

# Check for libcamera GStreamer plugin
if gst-inspect-1.0 libcamerasrc >/dev/null 2>&1; then
    echo "✓ GStreamer libcamera plugin is available"
else
    echo "✗ GStreamer libcamera plugin is not available"
fi

echo
echo "=== Docker Environment Check ==="
check_command "docker"
check_command "docker-compose"

echo
echo "=== Recommended Actions ==="

# Determine the best strategy
libcamera_available=false
v4l2_available=false
gstreamer_libcamera_available=false

if command -v libcamera-hello >/dev/null 2>&1; then
    libcamera_available=true
fi

if [ -e /dev/video0 ]; then
    v4l2_available=true
fi

if gst-inspect-1.0 libcamerasrc >/dev/null 2>&1; then
    gstreamer_libcamera_available=true
fi

if $libcamera_available && $gstreamer_libcamera_available; then
    echo "✓ RECOMMENDED: Use libcamera with GStreamer backend"
    echo "  - Best performance and future compatibility"
    echo "  - Use: GIZMATRON_CAMERA_BACKEND=libcamera"
    recommended_backend="libcamera"
elif $gstreamer_libcamera_available; then
    echo "✓ RECOMMENDED: Use GStreamer with libcamera"
    echo "  - Good compatibility with modern Raspberry Pi OS"
    echo "  - Use: GIZMATRON_CAMERA_BACKEND=gstreamer"
    recommended_backend="gstreamer"
elif $v4l2_available; then
    echo "⚠ FALLBACK: Use V4L2 compatibility mode"
    echo "  - Legacy support, may not work on future OS versions"
    echo "  - Use: GIZMATRON_CAMERA_BACKEND=v4l2"
    recommended_backend="v4l2"
else
    echo "✗ ERROR: No camera backend available"
    echo "  - Check camera connection and enable camera in raspi-config"
    recommended_backend="none"
fi

echo
echo "=== Migration Options ==="
echo "1. Quick fix (V4L2 compatibility)"
echo "2. Recommended upgrade (libcamera + GStreamer)"
echo "3. Manual configuration"
echo "4. Exit"

read -p "Choose an option (1-4): " choice

case $choice in
    1)
        echo
        echo "=== Quick Fix: Enable V4L2 Compatibility ==="
        echo "This will enable the legacy camera driver."
        echo
        echo "Add to /boot/config.txt:"
        echo "camera_auto_detect=0"
        echo "start_x=1"
        echo "gpu_mem=128"
        echo
        echo "Then reboot and run:"
        echo "sudo modprobe bcm2835-v4l2"
        echo
        echo "Use original docker-compose.yml"
        ;;
    
    2)
        echo
        echo "=== Recommended: libcamera + GStreamer ==="
        echo "This uses the new camera stack for best compatibility."
        echo
        
        if [ -f "docker-compose.libcamera.yml" ]; then
            echo "Using enhanced Docker configuration..."
            
            # Create environment file
            cat > .env << EOF
# Gizmatron Camera Configuration
GIZMATRON_CAMERA_BACKEND=${recommended_backend}
GIZMATRON_CAMERA_WIDTH=640
GIZMATRON_CAMERA_HEIGHT=480
GIZMATRON_CAMERA_FPS=30

# Twingate configuration (optional)
# TWINGATE_NETWORK=your-network
# TWINGATE_ACCESS_TOKEN=your-access-token
# TWINGATE_REFRESH_TOKEN=your-refresh-token
EOF
            
            echo "✓ Created .env file with recommended settings"
            echo
            echo "To start Gizmatron with libcamera support:"
            echo "docker-compose -f docker-compose.libcamera.yml up -d"
            echo
            echo "To enable remote access:"
            echo "docker-compose -f docker-compose.libcamera.yml --profile remote-access up -d"
            
        else
            echo "✗ docker-compose.libcamera.yml not found"
            echo "Please ensure all migration files are present"
        fi
        ;;
    
    3)
        echo
        echo "=== Manual Configuration ==="
        echo "Environment variables you can set:"
        echo "GIZMATRON_CAMERA_BACKEND=auto|libcamera|gstreamer|v4l2"
        echo "GIZMATRON_CAMERA_WIDTH=640"
        echo "GIZMATRON_CAMERA_HEIGHT=480"
        echo "GIZMATRON_CAMERA_FPS=30"
        echo
        echo "Docker run example:"
        echo "docker run -e GIZMATRON_CAMERA_BACKEND=libcamera \\"
        echo "  --device /dev/video0:/dev/video0 \\"
        echo "  --privileged gizmatron:libcamera"
        ;;
    
    4)
        echo "Exiting..."
        exit 0
        ;;
    
    *)
        echo "Invalid option"
        exit 1
        ;;
esac

echo
echo "=== Additional Resources ==="
echo "- Camera troubleshooting: CAMERA_FIX_STRATEGY.md"
echo "- Architecture overview: ARCHITECTURE.md"
echo "- Project documentation: README.md"
echo
echo "For support, check the project issues on GitHub"
echo "=== Migration Complete ===