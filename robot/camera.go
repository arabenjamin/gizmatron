package camera


import (
    "os"
	"fmt"
	_"bytes"
	_"log"
	_"image"
	_"strconv"
	_"net/http"
	"gocv.io/x/gocv"
	"image/color"
	"sync"

	"github.com/hybridgroup/mjpeg"

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

type Cam  struct {
	err error
	Webcam *gocv.VideoCapture
	ImgMat gocv.Mat
	Stream   *mjpeg.Stream
	//Img *image.Image
	mux sync.Mutex
}


func (c *Cam) InitCam() error {

	c.Webcam, c.err = gocv.OpenVideoCapture(-1)
	if c.err != nil {
		fmt.Println("error opening webcam")
		return c.err
	}
	fmt.Println("Camera is Initiated")

	go c.Start()
	go c.FaceDetect()
	return nil
}

func (c *Cam) CloseCam(){
	fmt.Println("Camera closed")
	c.Webcam.Close()
}

func (c *Cam) Restart(){
	c.CloseCam()
	c.InitCam()
}

func (c *Cam) Start(){

	// prepare image matrix
	c.ImgMat = gocv.NewMat()
	defer c.ImgMat.Close()

	// create the mjpeg stream
	c.Stream = mjpeg.NewStream()


	for {
		if ok := c.Webcam.Read(&c.ImgMat); !ok {
			fmt.Println("Cannot read from Device")
			//c.RestartCam()
			return
		}

		if !c.ImgMat.Empty() {
			//fmt.Println("Image Matrix is empty, moving forward ")
			//c.mux.Lock()
			//c.FaceDetect()
			buf, _ := gocv.IMEncode(".jpg", c.ImgMat)
			c.Stream.UpdateJPEG(buf)
			//c.mux.Unlock()
		}
	}
}


func (c *Cam) FaceDetect(){

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
		if !c.ImgMat.Empty(){
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



