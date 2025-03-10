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
}

func InitCam() (*Cam, error) {

	log.Printf("CAMERA: Initializing Camera ...")
	c := &Cam{
		DetectFaces:   false,
		IsOperational: false,
		IsRunning:     false,
	}

	c.open_wecam()

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

func (c *Cam) Stop() {

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

func (c *Cam) Restart() {

	log.Printf("Restarting Camera ...")
	c.Stop()
	go c.Start()
	log.Printf("Restarted camera successfully")

}

/* Start reading from the camera to the Buffer */
func (c *Cam) Start() {

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

					// resize the image to 320x240
					log.Println("Resizing image")
					gocv.Resize(c.ImgMat, &c.ImgMat, image.Point{600, 600}, 0, 0, gocv.InterpolationDefault)
					//buf, img_err := gocv.IMEncode(".jpg", c.ImgMat)
					c.ImgMat.ConvertTo(&c.ImgMat, gocv.MatTypeCV8UC3)
					buf, img_err := gocv.IMEncodeWithParams(".png", c.ImgMat, []int{gocv.IMWriteJpegQuality, 95})
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
	//webcam, err := gocv.VideoCaptureDevice(0)
	if c.Webcam == nil || c.IsOperational == false {
		var err error
		c.Webcam, err = gocv.OpenVideoCapture(0)
		if err != nil {
			fmt.Println("Error opeing webcam\n")
			return

		}
		c.IsOperational = true
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

func (c *Cam) StreamToServer() {

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

/* NOTE: This is for testing and debugging/troubleshooting */
func (c *Cam) RunCamera() {

	if c.Webcam == nil {
		var err error
		c.Webcam, err = gocv.VideoCaptureDevice(0)
		if err != nil {
			fmt.Println("Error: Could not open webcam")
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

func (c *Cam) open_wecam() {
	if c.Webcam == nil || c.IsOperational == false {
		log.Println("Camera is not operational, trying to open camera ...")
		var err error
		c.Webcam, err = gocv.OpenVideoCapture(-1)
		if err != nil {
			log.Printf("Error opeing webcam\nERROR: %v", err)
			c.IsOperational = false
			return
		}
		c.IsOperational = true
		log.Println("Camera is operational")
	}
}
