package robot

import (
	"fmt"
	"image/color"
	"log"
	"sync"
	"time"

	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

/* Takes picture saves as .jpeg*/
/*
func TakePicture() {

	fmt.Println("Taking Picture")
	//webcam, err := gocv.VideoCaptureDevice(0)
	webcam, err := gocv.OpenVideoCapture(0)
	if err != nil {
		fmt.Println("Error opeing webcam\n")
		return
	}
	defer webcam.Close()

	// prepare image matrix
	ImgMat := gocv.NewMat()
    defer ImgMat.Close()
	if ok := webcam.Read(&ImgMat); !ok {
		fmt.Println("Cannot read from Device")
		return
	}
	if !ImgMat.Empty(){
		fmt.Println("No image on device")
		gocv.IMWrite("image.jpg", ImgMat)
		return
	}
}
*/

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
	mux sync.Mutex
}

func InitCam() (*Cam, error) {
	c := &Cam{
		DetectFaces:   true,
		IsOperational: true,
	}

	c.Webcam, c.err = gocv.OpenVideoCapture(0)
	if c.err != nil {
		log.Printf("Error opening webcam")
		c.IsOperational = false
		return c, c.err
	}
	defer c.Webcam.Close()

	log.Printf("Camera is Initiated")
	return c, nil
}

func (c *Cam) CloseCam() {
	log.Printf("Camera closed")
	c.Webcam.Close()
}

func (c *Cam) Restart() {
	c.CloseCam()
	c.Start()
}

/* Start reading from the camera to the Buffer */
func (c *Cam) Start() {

	log.Printf("Starting Camera stream ...")

	if c.Webcam == nil {
		var err error
		c.Webcam, err = gocv.OpenVideoCapture(0)
		if err != nil {
			fmt.Println("Error: Could not open webcam")
			return
		}
		c.IsOperational = true
	}
	defer c.Webcam.Close()

	if c.IsOperational {

		c.DetectFaces = true

		// prepare image matrix
		c.ImgMat = gocv.NewMat()
		defer c.ImgMat.Close()

		// create the mjpeg stream
		c.Stream = mjpeg.NewStream()

		for {

			if ok := c.Webcam.Read(&c.ImgMat); !ok {

				log.Printf("Warning !! Cannot read from Camera Device: %v", ok)
				//c.RestartCam()
				// TODO : return an error
				return
			}

			if c.ImgMat.Empty() {
				log.Printf("Image Matrix is empty, moving forward ")
				continue
			}

			if !c.ImgMat.Empty() {
				c.IsRunning = true
				//c.mux.Lock()
				if c.DetectFaces {
					c.FaceDetect()
				}
				buf, _ := gocv.IMEncode(".jpg", c.ImgMat)
				c.Buf = buf.GetBytes()
				c.Stream.UpdateJPEG(buf.GetBytes())
				//	//c.mux.Unlock()

				// Sleep for a short duration to control the frame rate
				time.Sleep(33 * time.Millisecond) // ~30 FPS
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
		var base_path string = os.Getenv("GOPATH") + "/src/"
		var project_path string = base_path + "gocv.io/x/gocv/data/"
		var xmlFile string = project_path + "haarcascade_frontalface_default.xml"

		if !classifier.Load(xmlFile) {
			fmt.Printf("Error reading cascade file: %v\n", xmlFile)
			return
		}*/

	if !classifier.Load("/home/ara/opt/opencv/data/haarcascades/haarcascade_frontalface_default.xml") {
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

/* NOTE: This is for testing and debugging/troubleshooting */
func (c *Cam) RunCamera() {

	if c.Webcam == nil {
		var err error
		c.Webcam, err = gocv.OpenVideoCapture(0)
		if err != nil {
			fmt.Println("Error: Could not open webcam")
			return
		}
	}
	defer c.Webcam.Close()

	// create the mjpeg stream
	c.Stream = mjpeg.NewStream()

	// prepare image matrix
	c.ImgMat = gocv.NewMat()
	defer c.ImgMat.Close()

	// Loop to read the frames from the webcam
	for {

		c.IsRunning = true
		if ok := c.Webcam.Read(&c.ImgMat); !ok {
			fmt.Println("Error: Could not read a frame from the webcam")
			return
		}
		if c.ImgMat.Empty() {
			continue
		}

		buf, _ := gocv.IMEncode(".jpg", c.ImgMat)
		c.Buf = buf.GetBytes()
		c.Stream.UpdateJPEG(buf.GetBytes())

		// Sleep for a short duration to control the frame rate
		time.Sleep(33 * time.Millisecond) // ~30 FPS

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
