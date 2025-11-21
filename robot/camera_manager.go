package robot

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"gocv.io/x/gocv"
)

// CameraBackend represents different camera backend options
type CameraBackend string

const (
	BackendLibCamera CameraBackend = "libcamera"
	BackendV4L2      CameraBackend = "v4l2"
	BackendGStreamer CameraBackend = "gstreamer"
	BackendNone      CameraBackend = "none"
)

// CameraConfig holds camera configuration parameters
type CameraConfig struct {
	Backend  CameraBackend
	Device   string
	Width    int
	Height   int
	FPS      int
	Format   string
	Pipeline string
}

// DefaultCameraConfig returns default camera configuration
func DefaultCameraConfig() *CameraConfig {
	return &CameraConfig{
		Backend: detectCameraBackend(),
		Device:  "/dev/video0",
		Width:   640,
		Height:  480,
		FPS:     30,
		Format:  "MJPG",
	}
}

// detectCameraBackend automatically detects the best available camera backend
func detectCameraBackend() CameraBackend {
	// Check for libcamera availability
	if _, err := exec.LookPath("libcamera-hello"); err == nil {
		log.Println("CAMERA: libcamera detected")
		return BackendLibCamera
	}

	// Check for GStreamer with libcamera plugin
	if hasGStreamerLibCamera() {
		log.Println("CAMERA: GStreamer with libcamera detected")
		return BackendGStreamer
	}

	// Check for V4L2 device
	if _, err := os.Stat("/dev/video0"); err == nil {
		log.Println("CAMERA: V4L2 device detected")
		return BackendV4L2
	}

	log.Println("CAMERA: No camera backend detected")
	return BackendNone
}

// hasGStreamerLibCamera checks if GStreamer has libcamera plugin
func hasGStreamerLibCamera() bool {
	cmd := exec.Command("gst-inspect-1.0", "libcamerasrc")
	return cmd.Run() == nil
}

// CameraManager manages different camera backends
type CameraManager struct {
	config   *CameraConfig
	webcam   *gocv.VideoCapture
	pipeline string
}

// NewCameraManager creates a new camera manager with auto-detection
func NewCameraManager() *CameraManager {
	return &CameraManager{
		config: DefaultCameraConfig(),
	}
}

// NewCameraManagerWithConfig creates a camera manager with specific config
func NewCameraManagerWithConfig(config *CameraConfig) *CameraManager {
	return &CameraManager{
		config: config,
	}
}

// Initialize sets up the camera based on the detected/configured backend
func (cm *CameraManager) Initialize() error {
	log.Printf("CAMERA: Initializing camera with backend: %s", cm.config.Backend)

	switch cm.config.Backend {
	case BackendLibCamera:
		return cm.initLibCamera()
	case BackendGStreamer:
		return cm.initGStreamer()
	case BackendV4L2:
		return cm.initV4L2()
	default:
		return fmt.Errorf("no camera backend available")
	}
}

// initLibCamera initializes camera using libcamera with GStreamer
func (cm *CameraManager) initLibCamera() error {
	// Create GStreamer pipeline for libcamera
	cm.pipeline = fmt.Sprintf(
		"libcamerasrc ! video/x-raw,width=%d,height=%d,framerate=%d/1 ! videoconvert ! appsink drop=1",
		cm.config.Width, cm.config.Height, cm.config.FPS,
	)

	log.Printf("CAMERA: Using libcamera pipeline: %s", cm.pipeline)

	var err error
	cm.webcam, err = gocv.OpenVideoCapture(cm.pipeline)
	if err != nil {
		return fmt.Errorf("failed to open libcamera pipeline: %v", err)
	}

	return cm.validateCapture()
}

// initGStreamer initializes camera using pure GStreamer pipeline
func (cm *CameraManager) initGStreamer() error {
	// Try different GStreamer sources
	pipelines := []string{
		fmt.Sprintf("libcamerasrc ! video/x-raw,width=%d,height=%d,framerate=%d/1 ! videoconvert ! appsink",
			cm.config.Width, cm.config.Height, cm.config.FPS),
		fmt.Sprintf("v4l2src device=%s ! video/x-raw,width=%d,height=%d,framerate=%d/1 ! videoconvert ! appsink",
			cm.config.Device, cm.config.Width, cm.config.Height, cm.config.FPS),
	}

	for _, pipeline := range pipelines {
		log.Printf("CAMERA: Trying GStreamer pipeline: %s", pipeline)

		var err error
		cm.webcam, err = gocv.OpenVideoCapture(pipeline)
		if err != nil {
			log.Printf("CAMERA: Pipeline failed: %v", err)
			continue
		}

		if err = cm.validateCapture(); err != nil {
			log.Printf("CAMERA: Pipeline validation failed: %v", err)
			cm.webcam.Close()
			continue
		}

		cm.pipeline = pipeline
		return nil
	}

	return fmt.Errorf("all GStreamer pipelines failed")
}

// initV4L2 initializes camera using traditional V4L2
func (cm *CameraManager) initV4L2() error {
	log.Printf("CAMERA: Opening V4L2 device: %s", cm.config.Device)

	// Extract device number from device path
	deviceNum := 0
	if strings.HasPrefix(cm.config.Device, "/dev/video") {
		fmt.Sscanf(cm.config.Device, "/dev/video%d", &deviceNum)
	}

	var err error
	cm.webcam, err = gocv.OpenVideoCapture(deviceNum)
	if err != nil {
		return fmt.Errorf("failed to open V4L2 device %s: %v", cm.config.Device, err)
	}

	// Set camera properties
	cm.webcam.Set(gocv.VideoCaptureFrameWidth, float64(cm.config.Width))
	cm.webcam.Set(gocv.VideoCaptureFrameHeight, float64(cm.config.Height))
	cm.webcam.Set(gocv.VideoCaptureFPS, float64(cm.config.FPS))

	return cm.validateCapture()
}

// validateCapture tests if the camera can actually capture frames
func (cm *CameraManager) validateCapture() error {
	if cm.webcam == nil {
		return fmt.Errorf("webcam is nil")
	}

	testMat := gocv.NewMat()
	defer testMat.Close()

	// Try to read a frame with timeout
	for attempts := 0; attempts < 5; attempts++ {
		if ok := cm.webcam.Read(&testMat); ok && !testMat.Empty() {
			log.Printf("CAMERA: Successfully captured test frame (%dx%d)",
				testMat.Cols(), testMat.Rows())
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("failed to capture test frame after 5 attempts")
}

// GetVideoCapture returns the underlying VideoCapture object
func (cm *CameraManager) GetVideoCapture() *gocv.VideoCapture {
	return cm.webcam
}

// GetConfig returns the current camera configuration
func (cm *CameraManager) GetConfig() *CameraConfig {
	return cm.config
}

// GetPipeline returns the current GStreamer pipeline (if applicable)
func (cm *CameraManager) GetPipeline() string {
	return cm.pipeline
}

// Close releases camera resources
func (cm *CameraManager) Close() error {
	if cm.webcam != nil {
		cm.webcam.Close()
		cm.webcam = nil
	}
	return nil
}

// GetCameraInfo returns information about the camera setup
func (cm *CameraManager) GetCameraInfo() map[string]interface{} {
	info := map[string]interface{}{
		"backend":  string(cm.config.Backend),
		"device":   cm.config.Device,
		"width":    cm.config.Width,
		"height":   cm.config.Height,
		"fps":      cm.config.FPS,
		"pipeline": cm.pipeline,
	}

	if cm.webcam != nil {
		info["is_opened"] = cm.webcam.IsOpened()
		if cm.webcam.IsOpened() {
			info["actual_width"] = cm.webcam.Get(gocv.VideoCaptureFrameWidth)
			info["actual_height"] = cm.webcam.Get(gocv.VideoCaptureFrameHeight)
			info["actual_fps"] = cm.webcam.Get(gocv.VideoCaptureFPS)
		}
	}

	return info
}
