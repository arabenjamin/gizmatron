package robot

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"log"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

// CamV2 represents the updated camera implementation with backend support
type CamV2 struct {
	IsOperational  bool
	IsRunning      bool
	DetectFaces    bool
	err            error
	manager        *CameraManager
	ImgMat         gocv.Mat
	Stream         *mjpeg.Stream
	Buf            []byte
	mux            sync.Mutex
	StopStream     chan bool
	faceClassifier *gocv.CascadeClassifier
}

// InitCamV2 initializes the new camera implementation with automatic backend detection
func InitCamV2() (*CamV2, error) {
	log.Printf("CAMERA: Initializing Camera V2...")

	c := &CamV2{
		DetectFaces:   false,
		IsOperational: false,
		IsRunning:     false,
		StopStream:    make(chan bool),
	}

	// Initialize camera manager with auto-detection
	c.manager = NewCameraManager()
	if err := c.manager.Initialize(); err != nil {
		log.Printf("CAMERA: Failed to initialize camera manager: %v", err)
		return c, err
	}

	// Initialize image matrix
	c.ImgMat = gocv.NewMat()

	// Initialize face classifier for face detection
	c.initFaceClassifier()

	c.IsOperational = true
	log.Printf("CAMERA: Camera V2 initialized successfully with backend: %s",
		c.manager.GetConfig().Backend)

	return c, nil
}

// InitCamV2WithConfig initializes camera with specific configuration
func InitCamV2WithConfig(config *CameraConfig) (*CamV2, error) {
	log.Printf("CAMERA: Initializing Camera V2 with config...")

	c := &CamV2{
		DetectFaces:   false,
		IsOperational: false,
		IsRunning:     false,
		StopStream:    make(chan bool),
	}

	// Initialize camera manager with provided config
	c.manager = NewCameraManagerWithConfig(config)
	if err := c.manager.Initialize(); err != nil {
		log.Printf("CAMERA: Failed to initialize camera manager: %v", err)
		return c, err
	}

	// Initialize image matrix
	c.ImgMat = gocv.NewMat()

	// Initialize face classifier
	c.initFaceClassifier()

	c.IsOperational = true
	log.Printf("CAMERA: Camera V2 initialized with backend: %s",
		c.manager.GetConfig().Backend)

	return c, nil
}

// initFaceClassifier initializes the face detection classifier
func (c *CamV2) initFaceClassifier() {
	classifier := gocv.NewCascadeClassifier()
	c.faceClassifier = &classifier

	// Try multiple common paths for the face classifier
	classifierPaths := []string{
		"/usr/share/opencv4/haarcascades/haarcascade_frontalface_default.xml",
		"/usr/local/share/opencv4/haarcascades/haarcascade_frontalface_default.xml",
		"/home/ara/opencv/data/haarcascades/haarcascade_frontalface_default.xml",
		"./data/haarcascade_frontalface_default.xml",
	}

	for _, path := range classifierPaths {
		if c.faceClassifier.Load(path) {
			log.Printf("CAMERA: Face classifier loaded from: %s", path)
			return
		}
	}

	log.Printf("CAMERA: Warning - Could not load face classifier, face detection disabled")
	c.DetectFaces = false
}

// Start begins camera streaming and processing
func (c *CamV2) Start() error {
	if !c.IsOperational {
		return fmt.Errorf("camera is not operational")
	}

	if c.IsRunning {
		return fmt.Errorf("camera is already running")
	}

	log.Printf("CAMERA: Starting camera stream...")
	c.IsRunning = true

	webcam := c.manager.GetVideoCapture()
	if webcam == nil {
		return fmt.Errorf("webcam is not initialized")
	}

	go c.streamLoop(webcam)

	log.Printf("CAMERA: Camera stream started successfully")
	return nil
}

// streamLoop handles the main camera streaming loop
func (c *CamV2) streamLoop(webcam *gocv.VideoCapture) {
	defer func() {
		c.IsRunning = false
		log.Printf("CAMERA: Stream loop ended")
	}()

	frameCount := 0
	startTime := time.Now()

	for {
		select {
		case <-c.StopStream:
			log.Printf("CAMERA: Received stop signal")
			return

		default:
			if !c.readFrame(webcam) {
				log.Printf("CAMERA: Failed to read frame, retrying...")
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if c.ImgMat.Empty() {
				continue
			}

			c.processFrame()

			frameCount++
			if frameCount%30 == 0 { // Log FPS every 30 frames
				elapsed := time.Since(startTime)
				fps := float64(frameCount) / elapsed.Seconds()
				log.Printf("CAMERA: Processing at %.2f FPS", fps)
			}

			// Control frame rate
			time.Sleep(33 * time.Millisecond) // ~30 FPS
		}
	}
}

// readFrame reads a single frame from the camera
func (c *CamV2) readFrame(webcam *gocv.VideoCapture) bool {
	c.mux.Lock()
	defer c.mux.Unlock()

	if !webcam.Read(&c.ImgMat) {
		return false
	}

	return !c.ImgMat.Empty()
}

// processFrame processes the current frame (face detection, encoding, etc.)
func (c *CamV2) processFrame() {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.DetectFaces && c.faceClassifier != nil {
		c.detectFaces()
	}

	// Resize frame to configured dimensions
	config := c.manager.GetConfig()
	if c.ImgMat.Cols() != config.Width || c.ImgMat.Rows() != config.Height {
		gocv.Resize(c.ImgMat, &c.ImgMat,
			image.Point{config.Width, config.Height}, 0, 0, gocv.InterpolationDefault)
	}

	// Encode frame as JPEG
	buf, err := gocv.IMEncodeWithParams(".jpg", c.ImgMat,
		[]int{gocv.IMWriteJpegQuality, 95})
	if err != nil {
		log.Printf("CAMERA: Error encoding frame: %v", err)
		return
	}
	defer buf.Close()

	c.Buf = buf.GetBytes()
}

// detectFaces performs face detection on the current frame
func (c *CamV2) detectFaces() {
	if c.faceClassifier == nil {
		return
	}

	// Color for face rectangles
	blue := color.RGBA{0, 0, 255, 0}

	// Detect faces
	rects := c.faceClassifier.DetectMultiScale(c.ImgMat)

	if len(rects) > 0 {
		log.Printf("CAMERA: Detected %d face(s)", len(rects))
	}

	// Draw rectangles around detected faces
	for _, rect := range rects {
		gocv.Rectangle(&c.ImgMat, rect, blue, 3)
	}
}

// Stop stops the camera stream and releases resources
func (c *CamV2) Stop() error {
	log.Printf("CAMERA: Stopping camera...")

	if c.IsRunning {
		c.StopStream <- true
		c.IsRunning = false
	}

	if c.manager != nil {
		if err := c.manager.Close(); err != nil {
			log.Printf("CAMERA: Error closing camera manager: %v", err)
		}
	}

	if !c.ImgMat.Empty() {
		c.ImgMat.Close()
	}

	if c.faceClassifier != nil {
		c.faceClassifier.Close()
	}

	c.IsOperational = false
	log.Printf("CAMERA: Camera stopped successfully")
	return nil
}

// TakePicture captures a single frame and saves it
func (c *CamV2) TakePicture() error {
	if !c.IsOperational {
		return fmt.Errorf("camera is not operational")
	}

	log.Printf("CAMERA: Taking picture...")

	webcam := c.manager.GetVideoCapture()
	if webcam == nil {
		return fmt.Errorf("webcam is not available")
	}

	// Capture frame
	tempMat := gocv.NewMat()
	defer tempMat.Close()

	if !webcam.Read(&tempMat) || tempMat.Empty() {
		return fmt.Errorf("failed to capture frame")
	}

	// Apply face detection if enabled
	if c.DetectFaces && c.faceClassifier != nil {
		c.mux.Lock()
		c.ImgMat = tempMat.Clone()
		c.detectFaces()
		tempMat = c.ImgMat.Clone()
		c.mux.Unlock()
	}

	// Save image
	filename := fmt.Sprintf("picture_%d.jpg", time.Now().Unix())
	if !gocv.IMWrite(filename, tempMat) {
		return fmt.Errorf("failed to save picture")
	}

	log.Printf("CAMERA: Picture saved as %s", filename)
	return nil
}

// GetCurrentFrame returns the current frame as JPEG bytes
func (c *CamV2) GetCurrentFrame() []byte {
	c.mux.Lock()
	defer c.mux.Unlock()

	if len(c.Buf) == 0 {
		return nil
	}

	// Return a copy to avoid race conditions
	frameCopy := make([]byte, len(c.Buf))
	copy(frameCopy, c.Buf)
	return frameCopy
}

// GetCameraInfo returns detailed camera information
func (c *CamV2) GetCameraInfo() map[string]interface{} {
	info := map[string]interface{}{
		"operational":            c.IsOperational,
		"running":                c.IsRunning,
		"detect_faces":           c.DetectFaces,
		"face_classifier_loaded": c.faceClassifier != nil,
	}

	if c.manager != nil {
		managerInfo := c.manager.GetCameraInfo()
		for k, v := range managerInfo {
			info[k] = v
		}
	}

	return info
}

// SetFaceDetection enables or disables face detection
func (c *CamV2) SetFaceDetection(enabled bool) {
	c.DetectFaces = enabled && c.faceClassifier != nil
	log.Printf("CAMERA: Face detection set to %v", c.DetectFaces)
}

// StreamToServer sends frames to an external server (placeholder for future implementation)
func (c *CamV2) StreamToServer() {
	// This method maintains compatibility with the original interface
	// but should be reimplemented based on specific requirements

	client := &http.Client{}

	for c.IsRunning {
		frame := c.GetCurrentFrame()
		if frame == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("frame", "frame.jpg")
		if err != nil {
			log.Printf("CAMERA: Error creating form file: %v", err)
			continue
		}

		if _, err := part.Write(frame); err != nil {
			log.Printf("CAMERA: Error writing frame: %v", err)
			continue
		}

		if err := writer.Close(); err != nil {
			log.Printf("CAMERA: Error closing writer: %v", err)
			continue
		}

		req, err := http.NewRequest("POST", "http://localhost:9090/api/v1/upload", &body)
		if err != nil {
			log.Printf("CAMERA: Error creating request: %v", err)
			continue
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("CAMERA: Connection terminated: %v", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("CAMERA: Server response: %d", resp.StatusCode)
		}
		resp.Body.Close()

		time.Sleep(33 * time.Millisecond) // ~30 FPS
	}
}
