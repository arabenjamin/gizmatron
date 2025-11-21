# Raspberry Pi Camera Fix Strategy for Gizmatron

## Problem Analysis

The Raspberry Pi camera system has transitioned from V4L2 to libcamera, breaking OpenCV/GoCV compatibility. This document outlines multiple strategies to resolve this issue.

## Strategy 1: V4L2 Compatibility Layer (Quick Fix)

### Approach
Re-enable the legacy V4L2 compatibility layer to work with existing GoCV code.

### Implementation Steps

#### 1. Enable Legacy Camera Support
Add to `/boot/config.txt` on the Raspberry Pi:
```bash
# Enable legacy camera support
camera_auto_detect=0
start_x=1
gpu_mem=128

# Force V4L2 driver
dtoverlay=ov5647  # For Pi Camera v1
# OR
dtoverlay=imx219  # For Pi Camera v2
# OR  
dtoverlay=imx477  # For Pi Camera HQ
```

#### 2. Load V4L2 Module
Modify Docker setup to load the V4L2 compatibility module:
```dockerfile
# In Dockerfile, uncomment and modify:
RUN modprobe bcm2835-v4l2
```

#### 3. Docker Compose Changes
```yaml
services:
  gizmatron:
    privileged: true
    devices:
      - "/dev/video0:/dev/video0"
    volumes:
      - "/dev/video0:/dev/video0"
      - "/opt/vc:/opt/vc"  # VideoCore libraries
    environment:
      - LD_LIBRARY_PATH=/opt/vc/lib:$LD_LIBRARY_PATH
```

### Pros
- Minimal code changes
- Quick to implement
- Maintains existing GoCV integration

### Cons
- Uses deprecated technology
- May not work on future Raspberry Pi OS versions
- Limited long-term viability

## Strategy 2: OpenCV with libcamera Backend (Recommended)

### Approach
Rebuild OpenCV with libcamera support and create a libcamera-to-V4L2 bridge.

### Implementation Steps

#### 1. Install libcamera Development Libraries
```dockerfile
# Add to Dockerfile
RUN apt-get update && apt-get install -y \
    libcamera-dev \
    libcamera-apps \
    libcamera-tools \
    python3-libcamera \
    python3-kms++ \
    python3-pyqt5 \
    python3-prctl \
    libatlas-base-dev \
    ffmpeg \
    python3-pip
```

#### 2. Rebuild OpenCV with libcamera Support
Create a new Dockerfile stage:
```dockerfile
# Build OpenCV with libcamera support
FROM ubuntu:latest AS opencv-libcamera-builder

# Install dependencies
RUN apt-get update && apt-get install -y \
    build-essential cmake git pkg-config \
    libcamera-dev libcamera-apps \
    libjpeg-dev libpng-dev libtiff-dev \
    libavcodec-dev libavformat-dev libswscale-dev \
    libgtk-3-dev libcanberra-gtk-module libcanberra-gtk3-module \
    python3-dev python3-numpy \
    libtbb2 libtbb-dev libdc1394-22-dev \
    libxine2-dev libv4l-dev \
    libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev \
    qt5-default libvtk6-dev \
    libtesseract-dev libxvidcore-dev libx264-dev libgtk-3-dev \
    libopenexr-dev libatlas-base-dev gfortran \
    libgphoto2-dev libeigen3-dev libhdf5-dev

# Clone and build OpenCV with libcamera
RUN git clone https://github.com/opencv/opencv.git && \
    git clone https://github.com/opencv/opencv_contrib.git && \
    cd opencv && git checkout 4.11.0 && \
    cd ../opencv_contrib && git checkout 4.11.0

WORKDIR /opencv/build
RUN cmake -D CMAKE_BUILD_TYPE=RELEASE \
    -D CMAKE_INSTALL_PREFIX=/usr/local \
    -D OPENCV_EXTRA_MODULES_PATH=/opencv_contrib/modules \
    -D WITH_LIBCAMERA=ON \
    -D WITH_V4L=ON \
    -D WITH_GSTREAMER=ON \
    -D WITH_FFMPEG=ON \
    -D BUILD_EXAMPLES=OFF \
    -D INSTALL_PYTHON_EXAMPLES=OFF \
    -D INSTALL_C_EXAMPLES=OFF \
    -D PYTHON_EXECUTABLE=$(which python3) \
    .. && \
    make -j$(nproc) && \
    make install && \
    ldconfig
```

#### 3. Create libcamera to V4L2 Bridge Service
Create a bridge service that exposes libcamera as a V4L2 device:
```bash
#!/bin/bash
# libcamera-bridge.sh
# Create a virtual V4L2 device from libcamera

# Install v4l2loopback
modprobe v4l2loopback video_nr=0 card_label="libcamera-bridge"

# Start libcamera-vid to pipe to virtual device
libcamera-vid --codec mjpeg --width 640 --height 480 --framerate 30 --timeout 0 --output - | \
    ffmpeg -f mjpeg -i - -f v4l2 -vcodec rawvideo -pix_fmt yuv420p /dev/video0
```

## Strategy 3: Native libcamera Integration (Future-Proof)

### Approach
Create a Go wrapper for libcamera and modify Gizmatron to use it directly.

### Implementation Steps

#### 1. Create CGO libcamera Wrapper
```go
// camera_libcamera.go
package robot

/*
#cgo pkg-config: libcamera
#include <libcamera/libcamera.h>
#include <stdlib.h>

// C wrapper functions for libcamera
extern int init_libcamera();
extern int capture_frame(unsigned char** data, int* size);
extern void cleanup_libcamera();
*/
import "C"
import (
    "unsafe"
    "errors"
)

type LibCamera struct {
    initialized bool
}

func NewLibCamera() (*LibCamera, error) {
    if result := C.init_libcamera(); result != 0 {
        return nil, errors.New("failed to initialize libcamera")
    }
    return &LibCamera{initialized: true}, nil
}

func (lc *LibCamera) CaptureFrame() ([]byte, error) {
    if !lc.initialized {
        return nil, errors.New("libcamera not initialized")
    }
    
    var data *C.uchar
    var size C.int
    
    if result := C.capture_frame(&data, &size); result != 0 {
        return nil, errors.New("failed to capture frame")
    }
    
    // Convert C data to Go slice
    frame := C.GoBytes(unsafe.Pointer(data), size)
    C.free(unsafe.Pointer(data))
    
    return frame, nil
}

func (lc *LibCamera) Close() {
    if lc.initialized {
        C.cleanup_libcamera()
        lc.initialized = false
    }
}
```

#### 2. Modify Camera Implementation
```go
// camera_new.go - Updated camera implementation
package robot

import (
    "log"
    "image"
    "gocv.io/x/gocv"
)

type CamNew struct {
    IsOperational bool
    IsRunning     bool
    DetectFaces   bool
    libcam        *LibCamera
    ImgMat        gocv.Mat
    StopStream    chan bool
}

func InitCamNew() (*CamNew, error) {
    log.Printf("CAMERA: Initializing libcamera...")
    
    libcam, err := NewLibCamera()
    if err != nil {
        return nil, err
    }
    
    c := &CamNew{
        libcam:        libcam,
        DetectFaces:   false,
        IsOperational: true,
        IsRunning:     false,
        StopStream:    make(chan bool),
    }
    
    c.ImgMat = gocv.NewMat()
    log.Printf("libcamera initialized successfully")
    return c, nil
}

func (c *CamNew) CaptureFrame() error {
    frameData, err := c.libcam.CaptureFrame()
    if err != nil {
        return err
    }
    
    // Convert raw frame data to OpenCV Mat
    // This requires implementing proper format conversion
    // from libcamera formats to OpenCV formats
    
    return nil
}
```

## Strategy 4: GStreamer Pipeline Integration (Robust)

### Approach
Use GStreamer with libcamera source and OpenCV GStreamer backend.

### Implementation Steps

#### 1. Install GStreamer with libcamera
```dockerfile
RUN apt-get install -y \
    gstreamer1.0-tools \
    gstreamer1.0-plugins-base \
    gstreamer1.0-plugins-good \
    gstreamer1.0-plugins-bad \
    gstreamer1.0-libcamera \
    libgstreamer1.0-dev \
    libgstreamer-plugins-base1.0-dev
```

#### 2. Modify GoCV Integration
```go
// camera_gstreamer.go
package robot

import (
    "gocv.io/x/gocv"
    "log"
)

func (c *Cam) openWebcamGStreamer() error {
    // GStreamer pipeline for libcamera
    pipeline := "libcamerasrc ! videoconvert ! appsink"
    
    var err error
    c.Webcam, err = gocv.OpenVideoCaptureWithAPI(pipeline, gocv.VideoCaptureGStreamer)
    if err != nil {
        log.Printf("Error opening GStreamer pipeline: %v", err)
        return err
    }
    
    c.IsOperational = true
    log.Println("Camera opened with GStreamer pipeline")
    return nil
}
```

## Recommended Implementation Plan

### Phase 1: Quick Fix (Strategy 1)
1. Enable V4L2 compatibility layer
2. Test existing functionality
3. Document limitations

### Phase 2: Robust Solution (Strategy 4 - GStreamer)
1. Implement GStreamer pipeline integration
2. Add fallback mechanisms
3. Update Docker configurations
4. Add comprehensive error handling

### Phase 3: Future-Proofing (Strategy 3)
1. Develop native libcamera integration
2. Create abstraction layer for camera backends
3. Add support for multiple camera types

## Testing Strategy

### Unit Tests
```go
func TestCameraInitialization(t *testing.T) {
    cam, err := InitCam()
    if err != nil {
        t.Fatalf("Camera initialization failed: %v", err)
    }
    defer cam.Stop()
    
    if !cam.IsOperational {
        t.Error("Camera should be operational after initialization")
    }
}

func TestFrameCapture(t *testing.T) {
    cam, err := InitCam()
    if err != nil {
        t.Skip("Camera not available for testing")
    }
    defer cam.Stop()
    
    // Test frame capture
    err = cam.CaptureFrame()
    if err != nil {
        t.Errorf("Frame capture failed: %v", err)
    }
}
```

### Integration Tests
- Test camera functionality in Docker container
- Verify hardware device access
- Test API endpoints with camera operations

## Configuration Management

### Environment Variables
```bash
# Camera configuration
GIZMATRON_CAMERA_BACKEND=libcamera  # libcamera, v4l2, gstreamer
GIZMATRON_CAMERA_DEVICE=/dev/video0
GIZMATRON_CAMERA_WIDTH=640
GIZMATRON_CAMERA_HEIGHT=480
GIZMATRON_CAMERA_FPS=30
```

### Runtime Detection
```go
func detectCameraBackend() string {
    // Check for libcamera availability
    if _, err := exec.LookPath("libcamera-hello"); err == nil {
        return "libcamera"
    }
    
    // Check for V4L2 device
    if _, err := os.Stat("/dev/video0"); err == nil {
        return "v4l2"
    }
    
    return "none"
}
```

This comprehensive strategy provides multiple paths forward, from quick fixes to future-proof solutions, ensuring Gizmatron's camera functionality works reliably on modern Raspberry Pi systems.