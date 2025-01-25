package robot

import (
	"fmt"
	"image/color"
	"log"
	"os"
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
	IsRunning bool
	err       error
	Webcam    *gocv.VideoCapture
	ImgMat    gocv.Mat
	Stream    *mjpeg.Stream
	//Img *image.Image
	mux sync.Mutex
}

func InitCam() (*Cam, error) {
	c := &Cam{}
<<<<<<< HEAD
  c.Webcam, c.err = gocv.OpenVideoCapture(0)  
  //c.Webcam, c.err = gocv.OpenVideoCapture(-1)
	//c.Webcam, c.err = gocv.VideoCaptureDevice(0)
	//c.Webcam, c.err = gocv.OpenVideoCaptureWithAPI(0, 1900) //200 V4L 1800 Gstreamer 1900 FFmpeg
=======
	c.Webcam, c.err = gocv.OpenVideoCapture(0)
>>>>>>> attempt to get the camera to stream
	if c.err != nil {
		log.Printf("Error opening webcam")
		c.IsRunning = false
		return c, c.err
	}

	defer c.Webcam.Close()

	// prepare image matrix
	c.ImgMat = gocv.NewMat()
	defer c.ImgMat.Close()

	log.Printf("Camera is Initiated")
	c.IsRunning = true

	//go c.Start()
	//c.Start()
	// go c.FaceDetect()
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

func (c *Cam) Start() {

	// prepare image matrix
	//c.ImgMat = gocv.NewMat()
	//defer c.ImgMat.Close()

	// create the mjpeg stream
	c.Stream = mjpeg.NewStream()

	for {

		if ok := c.Webcam.Read(&c.ImgMat); !ok {
<<<<<<< HEAD
			log.Printf("Warning !! Cannot read from Device: %v", ok)
			//c.RestartCam()  
=======

			log.Printf("Warning !! Cannot read from Camera Device: %v", ok)
			//c.RestartCam()
>>>>>>> attempt to get the camera to stream
			// TODO : return an error
			return
		}

		if c.ImgMat.Empty() {
			log.Printf("Image Matrix is empty, moving forward ")
			continue
		}

		if !c.ImgMat.Empty() {

			//c.mux.Lock()
			//c.FaceDetect()
			buf, _ := gocv.IMEncode(".jpg", c.ImgMat)
			//	buf, _ := gocv.IMEncode(".jpg", img)
			c.Stream.UpdateJPEG(buf.GetBytes())
			//	//c.mux.Unlock()
		}
	}
}

func (c *Cam) FaceDetect() {

	// color for the rect when faces detected
	blue := color.RGBA{0, 0, 255, 0}

	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	/* */
	var base_path string = os.Getenv("GOPATH") + "/src/"
	var project_path string = base_path + "gocv.io/x/gocv/data/"
	var xmlFile string = project_path + "haarcascade_frontalface_default.xml"

	if !classifier.Load(xmlFile) {
		fmt.Printf("Error reading cascade file: %v\n", xmlFile)
		return
	}

	for {
		if !c.ImgMat.Empty() {
			// detect faces
			//c.mux.Lock()
			rects := classifier.DetectMultiScale(c.ImgMat)
			fmt.Printf("found %d faces\n", len(rects))
			// draw a rectangle around each face on the original image,
			// along with text identifing as "Human"
			for _, r := range rects {
				gocv.Rectangle(&c.ImgMat, r, blue, 3)

				//size := gocv.GetTextSize("Human", gocv.FontHersheyPlain, 1.2, 2)
				//pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
				//gocv.PutText(&c.ImgMat, "Human", pt, gocv.FontHersheyPlain, 1.2, blue, 2)
			}
			//c.mux.Unlock()
		}
	}

}

// RunCamera starts the camera
func (c *Cam) RunCamera() {
	//c.IsRunning = true

	/*
		webcam, err := gocv.OpenVideoCapture(0)
		//webcam, err := gocv.VideoCaptureDevice(0, gocv.VideoCaptureV4L2)
		if err != nil {
			fmt.Printf("Error: Could not open the webcam: %v\n", err)
			return
		}
		defer webcam.Close()


	*/

	// Create a Mat to hold the frame
	img := gocv.NewMat()
	defer img.Close()

	// Create a window to display the video
	window := gocv.NewWindow("Webcam Video")
	defer window.Close()

	// create the mjpeg stream
	//c.Stream = mjpeg.NewStream()

	// Loop to read the frames from the webcam
	for {

		if ok := c.Webcam.Read(&c.ImgMat); !ok {
			fmt.Println("Error: Could not read a frame from the webcam")
			return
		}
		if c.ImgMat.Empty() {
			continue
		}

		// Display the frame in the window
		if !c.ImgMat.Empty() {

			window.IMShow(c.ImgMat)
			//buf, _ := gocv.IMEncode(".jpg", c.ImgMat)
			//	buf, _ := gocv.IMEncode(".jpg", img)
			//c.Stream.UpdateJPEG(buf.GetBytes())
		}

		// Wait for 1 millisecond and check if 'q' is pressed
		if window.WaitKey(1) == 'q' {
			break
		}
		// Sleep for a short duration to control the frame rate
		time.Sleep(33 * time.Millisecond) // ~30 FPS

	}

}
