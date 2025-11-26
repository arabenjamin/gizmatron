package robot

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

// CameraBackend represents the type of camera backend being used
type CameraBackend string

const (
	BackendGStreamer CameraBackend = "gstreamer"
	BackendV4L2      CameraBackend = "v4l2"
	BackendAuto      CameraBackend = "auto"
)

// CameraConfig holds camera configuration
type CameraConfig struct {
	Backend CameraBackend
	Device  int // V4L2 device number (0, 1, etc.)
	Width   int // Frame width
	Height  int // Frame height
	FPS     int // Frames per second
}

type Cam struct {
	IsOperational bool
	IsRunning     bool
	DetectFaces   bool
	err           error
	Webcam        *gocv.VideoCapture
	ImgMat        gocv.Mat
	Stream        *mjpeg.Stream
	Buf           []byte
	//Img *image.Image
	mux        sync.Mutex
	StopStream chan bool
	Config     CameraConfig
	Backend    CameraBackend // Actual backend in use
}

// loadCameraConfig loads camera configuration from environment variables
func loadCameraConfig() CameraConfig {
	config := CameraConfig{
		Backend: BackendAuto,
		Device:  0,
		Width:   640,
		Height:  480,
		FPS:     30,
	}

	// Load backend preference
	if backend := os.Getenv("GIZMATRON_CAMERA_BACKEND"); backend != "" {
		config.Backend = CameraBackend(backend)
		log.Printf("CAMERA: Backend set via environment: %s", backend)
	}

	// Load device number
	if device := os.Getenv("GIZMATRON_CAMERA_DEVICE"); device != "" {
		if d, err := strconv.Atoi(device); err == nil {
			config.Device = d
			log.Printf("CAMERA: Device set via environment: %d", d)
		}
	}

	// Load dimensions
	if width := os.Getenv("GIZMATRON_CAMERA_WIDTH"); width != "" {
		if w, err := strconv.Atoi(width); err == nil {
			config.Width = w
		}
	}
	if height := os.Getenv("GIZMATRON_CAMERA_HEIGHT"); height != "" {
		if h, err := strconv.Atoi(height); err == nil {
			config.Height = h
		}
	}

	// Load FPS
	if fps := os.Getenv("GIZMATRON_CAMERA_FPS"); fps != "" {
		if f, err := strconv.Atoi(fps); err == nil {
			config.FPS = f
		}
	}

	return config
}

// detectGStreamerSupport checks if GStreamer and libcamera are available
func detectGStreamerSupport() bool {
	// Check if libcamera-hello command exists (indicates Pi camera support)
	if _, err := exec.LookPath("libcamera-hello"); err == nil {
		log.Printf("CAMERA: libcamera tools detected")
		return true
	}

	// Check if gst-launch-1.0 exists (GStreamer installed)
	if _, err := exec.LookPath("gst-launch-1.0"); err == nil {
		log.Printf("CAMERA: GStreamer detected")
		return true
	}

	return false
}

// detectV4L2Device checks if a V4L2 device exists
func detectV4L2Device(deviceNum int) bool {
	devicePath := fmt.Sprintf("/dev/video%d", deviceNum)
	if _, err := os.Stat(devicePath); err == nil {
		log.Printf("CAMERA: V4L2 device detected: %s", devicePath)
		return true
	}
	return false
}

func InitCam() (*Cam, error) {

	log.Printf("CAMERA: Initializing Camera ...")

	// Load configuration from environment
	config := loadCameraConfig()

	c := &Cam{
		DetectFaces:   false,
		IsOperational: false,
		IsRunning:     false,
		Config:        config,
	}

	//c.open_wecam()
	//defer c.Webcam.Close()
	/*
		if c.IsOperational {

			resp, err := http.Get("http://localhost:9090/api/v1/ping")
			if err != nil {
				log.Printf("Unable to reach Control server, Request returned Error: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				go c.StreamToServer()
			} else {
				log.Printf("Unable to reach control server, Response returned StatusCode: %v", resp.StatusCode)
			}
		}
	*/

	/* Image matrix */
	c.ImgMat = gocv.NewMat()
	defer c.ImgMat.Close()

	c.StopStream = make(chan bool)
	//defer close(c.StopStream)
	log.Printf("Camera Ready ...")
	return c, nil
}

// tryOpenGStreamer attempts to open camera using GStreamer pipeline
func (c *Cam) tryOpenGStreamer() error {
	log.Printf("CAMERA: Attempting to open with GStreamer + libcamera...")

	// GStreamer pipeline for libcamera (Pi Camera Module)
	pipeline := "libcamerasrc ! video/x-raw,width=%d,height=%d,framerate=%d/1 ! videoconvert ! appsink"
	pipelineStr := fmt.Sprintf(pipeline, c.Config.Width, c.Config.Height, c.Config.FPS)

	log.Printf("CAMERA: Using GStreamer pipeline: %s", pipelineStr)

	var err error
	// GoCV detects GStreamer pipelines automatically from the string format
	c.Webcam, err = gocv.OpenVideoCapture(pipelineStr)
	if err != nil {
		return fmt.Errorf("GStreamer pipeline failed: %w", err)
	}

	// Verify we can read a frame
	testMat := gocv.NewMat()
	defer testMat.Close()
	if ok := c.Webcam.Read(&testMat); !ok || testMat.Empty() {
		c.Webcam.Close()
		c.Webcam = nil
		return fmt.Errorf("GStreamer opened but cannot read frames")
	}

	c.Backend = BackendGStreamer
	log.Printf("CAMERA: Successfully opened with GStreamer + libcamera")
	return nil
}

// tryOpenV4L2 attempts to open camera using V4L2 (USB/integrated webcams)
func (c *Cam) tryOpenV4L2(deviceNum int) error {
	log.Printf("CAMERA: Attempting to open V4L2 device %d...", deviceNum)

	var err error
	c.Webcam, err = gocv.OpenVideoCapture(deviceNum)
	if err != nil {
		return fmt.Errorf("V4L2 device %d failed: %w", deviceNum, err)
	}

	// Set camera properties
	c.Webcam.Set(gocv.VideoCaptureFrameWidth, float64(c.Config.Width))
	c.Webcam.Set(gocv.VideoCaptureFrameHeight, float64(c.Config.Height))
	c.Webcam.Set(gocv.VideoCaptureFPS, float64(c.Config.FPS))

	// Verify we can read a frame
	testMat := gocv.NewMat()
	defer testMat.Close()
	if ok := c.Webcam.Read(&testMat); !ok || testMat.Empty() {
		c.Webcam.Close()
		c.Webcam = nil
		return fmt.Errorf("V4L2 device %d opened but cannot read frames", deviceNum)
	}

	c.Backend = BackendV4L2
	log.Printf("CAMERA: Successfully opened V4L2 device %d", deviceNum)
	return nil
}

func (c *Cam) open_wecam() {
	if c.Webcam != nil && c.IsOperational {
		log.Println("CAMERA: Already operational")
		return
	}

	log.Println("CAMERA: Opening camera with auto-detection...")

	var lastErr error

	// Try based on configuration
	switch c.Config.Backend {
	case BackendGStreamer:
		// User explicitly wants GStreamer
		if err := c.tryOpenGStreamer(); err != nil {
			log.Printf("CAMERA: GStreamer failed (explicit): %v", err)
			c.IsOperational = false
			return
		}
		c.IsOperational = true
		return

	case BackendV4L2:
		// User explicitly wants V4L2
		if err := c.tryOpenV4L2(c.Config.Device); err != nil {
			log.Printf("CAMERA: V4L2 failed (explicit): %v", err)
			c.IsOperational = false
			return
		}
		c.IsOperational = true
		return

	case BackendAuto:
		// Auto-detect: Try GStreamer first (for Pi Camera Module), then V4L2
		log.Printf("CAMERA: Auto-detecting camera backend...")

		// Try GStreamer if supported
		if detectGStreamerSupport() {
			if err := c.tryOpenGStreamer(); err == nil {
				c.IsOperational = true
				return
			} else {
				lastErr = err
				log.Printf("CAMERA: GStreamer attempt failed: %v", err)
			}
		}

		// Try V4L2 device 0 (most common)
		if detectV4L2Device(0) {
			if err := c.tryOpenV4L2(0); err == nil {
				c.IsOperational = true
				return
			} else {
				lastErr = err
				log.Printf("CAMERA: V4L2 device 0 attempt failed: %v", err)
			}
		}

		// Try V4L2 device 1 (alternative USB camera)
		if detectV4L2Device(1) {
			if err := c.tryOpenV4L2(1); err == nil {
				c.IsOperational = true
				return
			} else {
				lastErr = err
				log.Printf("CAMERA: V4L2 device 1 attempt failed: %v", err)
			}
		}

		// All attempts failed
		log.Printf("CAMERA: All camera backends failed. Last error: %v", lastErr)
		c.IsOperational = false
		return
	}

	log.Println("CAMERA: Camera is now operational")
	c.IsOperational = true
}

func (c *Cam) Stop() {
	/* Stop the camera and the stream */
	log.Printf("Closing Camera ....")

	c.StopStream <- true
	c.IsRunning = false

	if c.Webcam != nil {
		c.Webcam.Close()
		c.Webcam = nil
	}
	c.IsOperational = false
	log.Printf("Camera Closed")

}

func (c *Cam) Start() {
	/* Start reading from the camera to the Buffer */
	log.Printf("Starting Camera stream ...")
	c.open_wecam()
	defer c.Webcam.Close()

	// prepare image matrix
	c.ImgMat = gocv.NewMat()
	defer c.ImgMat.Close()

	// create the mjpeg stream
	//c.Stream = mjpeg.NewStream()
	if c.IsOperational && c.Webcam != nil {

		log.Printf("Camera is operational, starting stream ...")
		for {
			select {
			case <-c.StopStream:
				log.Printf("Recieved Stop signal, Stopping Camera Stream..")
				return

			default:

				log.Println("Reading from Camera")
				c.Webcam.Set(gocv.VideoCaptureFrameWidth, 600)
				c.Webcam.Set(gocv.VideoCaptureFrameHeight, 600)

				if ok := c.Webcam.Read(&c.ImgMat); !ok {

					log.Printf("Error !! Cannot read from Camera Device: %v", ok)
				}

				if c.ImgMat.Empty() {
					log.Printf("Image Matrix is empty, moving forward ")
				}

				if !c.ImgMat.Empty() {
					c.IsRunning = true
					//c.mux.Lock()

					if c.DetectFaces {
						c.FaceDetect()
					}

					// resize the image to to make sure it fits into the stream
					log.Println("Resizing image")
					gocv.Resize(c.ImgMat, &c.ImgMat, image.Point{600, 600}, 0, 0, gocv.InterpolationDefault)
					//buf, img_err := gocv.IMEncode(".jpg", c.ImgMat)
					//c.ImgMat.ConvertTo(&c.ImgMat, gocv.MatTypeCV8UC3)
					buf, img_err := gocv.IMEncodeWithParams(".jpg", c.ImgMat, []int{gocv.IMWriteJpegQuality, 95})
					if img_err != nil {
						log.Printf("Error encoding image: %v", img_err)
					}
					defer buf.Close()
					c.Buf = buf.GetBytes()
					//c.Stream.UpdateJPEG(c.Buf)
					//	//c.mux.Unlock()

					// Sleep for a short duration to control the frame rate
					time.Sleep(33 * time.Millisecond) // ~30 FPS
				}
			}
		}
	}
}

func (c *Cam) Restart() {

	log.Printf("Restarting Camera ...")
	c.Stop()
	go c.Start()
	log.Printf("Restarted camera successfully")

}

func (c *Cam) StreamToServer() {

	/* Sends Camera steam to WEbsocket server */

	// Currently this is all wrong, and doesn't use websockets at all.
	// It should be using websockets to send the stream to the server.

	// TODO: Write tests to verify I've done this correctly

	client := &http.Client{}

	for {

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("frame", "frame.jpg")
		if err != nil {
			log.Fatal(err)
		}
		part.Write(c.Buf)
		writer.Close()

		//bufferReader := &CustomBufferReader{buf: bytes.NewBuffer(c.Buf)}
		//log.Println("Makeing request")
		req, err := http.NewRequest("POST", "http://localhost:9090/api/v1/upload", &body) //
		if err != nil {
			log.Fatal(err)
		}
		//req.Header.Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		req.Header.Set("Content-Type", "video/mp4")
		req.Header.Set("Transfer-Encoding", "chunked")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Connection terminated: %v", err)
		}
		defer resp.Body.Close()

		//log.Printf("Request results %d", resp.StatusCode)
		if resp.StatusCode != http.StatusOK {
			log.Println("Response from Server: ", resp.StatusCode)
			break
		}

		// Sleep for a short duration to control the frame rate
		time.Sleep(33 * time.Millisecond) // ~30 FPS

	}
}

func (c *Cam) FaceDetect() {

	//log.Printf("Detecting Faces")

	// color for the rect when faces detected
	blue := color.RGBA{0, 0, 255, 0}

	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	/*
		TODO: I need a better way of pointing to this classifier model file.

		var base_path string = os.Getenv("GOPATH") + "/src/"
		var project_path string = base_path + "gocv.io/x/gocv/data/"
		var xmlFile string = project_path + "haarcascade_frontalface_default.xml"

		if !classifier.Load(xmlFile) {
			fmt.Printf("Error reading cascade file: %v\n", xmlFile)
			return
		}
	*/

	// TODO: This is cheap, and I need a better way. This is not the way.
	if !classifier.Load("/home/ara/opencv/data/haarcascades/haarcascade_frontalface_default.xml") {
		fmt.Println("Error: Could not load Haar Cascade classifier")
		return
	}

	if !c.ImgMat.Empty() {

		//c.mux.Lock()

		// detect faces
		rects := classifier.DetectMultiScale(c.ImgMat)
		//fmt.Printf("found %d faces\n", len(rects))

		// draw a rectangle around each face on the original image,
		for _, r := range rects {
			gocv.Rectangle(&c.ImgMat, r, blue, 3)
			// TODO: add text identifing as "Human face"
		}

	}
}

/* Takes picture saves as .jpeg*/
func (c *Cam) TakePicture() {

	// TODO: serve jpeg to frontend

	fmt.Println("Taking Picture")

	// Use the new backend-aware camera opening
	if c.Webcam == nil || c.IsOperational == false {
		c.open_wecam()
		if !c.IsOperational {
			fmt.Println("Error: Could not open camera")
			return
		}
	}
	defer c.Webcam.Close()

	if ok := c.Webcam.Read(&c.ImgMat); !ok {
		fmt.Println("Cannot read from Device")
		return
	}

	if !c.ImgMat.Empty() {
		if c.DetectFaces {
			c.FaceDetect()
		}
		fmt.Println("No image on device")
		gocv.IMWrite("image.jpg", c.ImgMat)
		return
	}
}

/* NOTE: This is for testing and debugging/troubleshooting */
func (c *Cam) RunCamera() {

	// Use the new backend-aware camera opening
	if c.Webcam == nil {
		c.open_wecam()
		if !c.IsOperational {
			fmt.Println("Error: Could not open camera")
			return
		}
	}
	defer c.Webcam.Close()

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// Loop to read the frames from the webcam
	for {

		c.IsRunning = true
		if !img.Closed() {

			if ok := c.Webcam.Read(&img); !ok {

				log.Printf("Error !! Cannot read from Camera Device: %v", ok)
				c.IsOperational = false

				continue
			}

			if img.Empty() {
				continue
			}

			buf, _ := gocv.IMEncode(".jpg", img)
			c.Buf = buf.GetBytes()

			// Sleep for a short duration to control the frame rate
			time.Sleep(33 * time.Millisecond) // ~30 FPS
		}

	}
}

/* NOTE: this should only be run for testing and debugging/troubleshooting purposes */
func (c *Cam) RunCamInWindow() {

	/* Create a window to display the video */

	window := gocv.NewWindow("Webcam Video")
	defer window.Close()
	if c.IsOperational && c.IsRunning {
		for {

			if c.ImgMat.Empty() {
				continue
			}

			if !c.ImgMat.Empty() {
				window.IMShow(c.ImgMat)
			}

		}
	}
}
