# Camera Usage Guide

## Overview

Gizmatron now supports multiple camera backends with automatic detection:
- **V4L2**: USB webcams and integrated laptop cameras (development and Pi)
- **GStreamer + libcamera**: Raspberry Pi Camera Module via CSI port (production)

The system automatically detects and uses the best available camera backend.

## Quick Start

### Development (Laptop/Desktop)

Just run gizmatron - it will automatically detect your USB or integrated webcam:

```bash
./gizmatron
```

The system will try V4L2 devices in order: `/dev/video0`, then `/dev/video1`.

### Raspberry Pi with USB Webcam

Same as development - automatic detection will find the USB webcam:

```bash
./gizmatron
```

### Raspberry Pi with Camera Module (CSI)

Ensure GStreamer and libcamera tools are installed (included in Docker image):

```bash
./gizmatron
```

The system will automatically detect libcamera support and use the CSI camera module.

## Configuration

### Environment Variables

Override automatic detection with these optional variables:

```bash
# Force specific backend
GIZMATRON_CAMERA_BACKEND=auto    # Options: auto, gstreamer, v4l2

# V4L2 device number (default: 0)
GIZMATRON_CAMERA_DEVICE=0        # /dev/video0, /dev/video1, etc.

# Frame dimensions (default: 640x480)
GIZMATRON_CAMERA_WIDTH=640
GIZMATRON_CAMERA_HEIGHT=480

# Frame rate (default: 30 fps)
GIZMATRON_CAMERA_FPS=30
```

### Example: Force USB Webcam on Pi

```bash
GIZMATRON_CAMERA_BACKEND=v4l2 GIZMATRON_CAMERA_DEVICE=1 ./gizmatron
```

### Example: Force Pi Camera Module

```bash
GIZMATRON_CAMERA_BACKEND=gstreamer ./gizmatron
```

### Example: Custom Resolution

```bash
GIZMATRON_CAMERA_WIDTH=1280 GIZMATRON_CAMERA_HEIGHT=720 ./gizmatron
```

## How Auto-Detection Works

When `GIZMATRON_CAMERA_BACKEND=auto` (default):

1. **Check for GStreamer + libcamera** (Pi Camera Module)
   - Looks for `libcamera-hello` command
   - Looks for `gst-launch-1.0` command
   - If found, tries GStreamer pipeline

2. **Check for V4L2 device 0** (primary USB/integrated camera)
   - Checks if `/dev/video0` exists
   - If found, tries to open and read test frame

3. **Check for V4L2 device 1** (secondary USB camera)
   - Checks if `/dev/video1` exists
   - If found, tries to open and read test frame

4. **Report failure** if all backends fail

## Docker Configuration

The Dockerfile includes all necessary dependencies:

```dockerfile
# GStreamer for camera pipeline support
gstreamer1.0-tools
gstreamer1.0-plugins-base
gstreamer1.0-plugins-good
gstreamer1.0-plugins-bad
libgstreamer1.0-dev

# libcamera for Pi Camera Module
libcamera-dev
libcamera-tools

# V4L2 utilities
v4l-utils
```

## Troubleshooting

### Camera Not Detected

Check available cameras:

```bash
# List V4L2 devices
v4l2-ctl --list-devices

# Test Pi Camera Module
libcamera-hello --list-cameras

# Test GStreamer pipeline
gst-launch-1.0 libcamerasrc ! autovideosink
```

### Enable Debug Logging

The camera system logs all detection and initialization steps:

```bash
./gizmatron 2>&1 | grep CAMERA
```

Expected log output:

```
CAMERA: Initializing Camera ...
CAMERA: Auto-detecting camera backend...
CAMERA: libcamera tools detected
CAMERA: Attempting to open with GStreamer + libcamera...
CAMERA: Using GStreamer pipeline: libcamerasrc ! video/x-raw,width=640,height=480,framerate=30/1 ! videoconvert ! appsink
CAMERA: Successfully opened with GStreamer + libcamera
CAMERA: Camera Ready ...
```

### Permission Issues

Ensure user has access to video devices:

```bash
# Add user to video group
sudo usermod -aG video $USER

# Logout and login for changes to take effect
```

### Pi Camera Module Not Working

1. Enable camera in raspi-config:
   ```bash
   sudo raspi-config
   # Interface Options -> Camera -> Enable
   ```

2. Verify libcamera installation:
   ```bash
   libcamera-hello --list-cameras
   ```

3. Test camera outside container:
   ```bash
   libcamera-still -o test.jpg
   ```

### USB Webcam Not Working

1. Check device exists:
   ```bash
   ls -la /dev/video*
   ```

2. Test with v4l2:
   ```bash
   v4l2-ctl --device=/dev/video0 --all
   ```

3. Verify Docker has access:
   ```bash
   docker run --rm --device /dev/video0 gizmatron:latest ls -la /dev/video0
   ```

## API Endpoints

### Check Camera Status

```bash
curl http://localhost:8080/api/v1/bot-status
```

Response includes camera operational status:

```json
{
  "name": "Gizmatron",
  "status": "operational",
  "devices": {
    "camera": {
      "operational": true,
      "running": false
    }
  }
}
```

### Start Camera Stream

```bash
curl -X POST http://localhost:8080/api/v1/camera/start
```

### Stop Camera Stream

```bash
curl -X POST http://localhost:8080/api/v1/camera/stop
```

### Take Picture

```bash
curl -X POST http://localhost:8080/api/v1/camera/picture
```

## Development Workflow

1. **Develop on laptop** with built-in or USB webcam
2. **Test on Pi** with USB webcam for quick iteration
3. **Deploy to production** with Pi Camera Module for final hardware

All three scenarios work automatically without code changes!

## Future Enhancements

- [ ] Support for multiple simultaneous cameras
- [ ] Camera selection via API
- [ ] Hot-plugging USB cameras
- [ ] RTSP streaming support
- [ ] Hardware-accelerated encoding on Pi
