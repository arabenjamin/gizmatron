package robot

import (
	"fmt"
	"log"

	"github.com/warthog618/go-gpiocdev"
	"gobot.io/x/gobot/v2/platforms/raspi"
)

const (
	RUNNING_LED   = 37 //gpio 26 pin 37
	SEVER_LED     = 13 //gpio 13 pin 33
	ARM_LED       = 5  //gpio 05 pin 29
	BASE_SERVO    = 0
	JOINT_1_SERVO = 1
	JOINT_2_SERVO = 2
	JOINT_3_SERVO = 3
	JOINT_4_SERVO = 4
)

type Device struct {
	Name   string
	Status string
	Data   map[string]interface{}
	Error  string
}

type Robot struct {
	Name       string
	IsRunning  bool
	State      bool           // depreciated
	adaptor    *raspi.Adaptor // I really want to depreciate this
	runningled *gpiocdev.Line
	Serverled  *gpiocdev.Line
	armled     *gpiocdev.Line
	arm        *Arm
	Camera     *Cam
	Devices    map[string]*Device
	log        *log.Logger
}

func InitRobot(botlog *log.Logger) (*Robot, error) {

	botlog.Println("Initializing Gizmatron startup ...")
	robot := &Robot{
		Name:    "Gizmatron",
		adaptor: raspi.NewAdaptor(),
		Devices: make(map[string]*Device),
		log:     botlog,
	}

	/* Start our devices*/
	robot.log.Println("Initalizing Gizmatron Devices ...")
	robot.IsRunning = true
	err := robot.initDevices()
	if err != nil {
		// TODO: This error handler needs to be rethought.
		// When initDevices() runs,
		// we want it to tell us which devices have errors, not that initDevices()
		// had errors
		robot.log.Printf("%v failed to intialize device: %v", robot.Name, err)
		robot.IsRunning = false
	}
	robot.log.Println("Gizmatron devices initialized.")

	// NOTE: IF we dont make this check and try to change the value of a non-existent pin ...
	// really narly shit happens.
	// Rember to mind you P's and Q's
	if robot.Devices["runningLed"].Status == "Operational" {
		// Turn on our operating light
		robot.runningled.SetValue(1)
	}

	robot.log.Println("Gizmatron Startup Complete.")
	return robot, nil
}

func (r *Robot) initDevices() error {

	// TODO: When initDevices() runs,
	// we want it to return which devices have errors,
	// So this method needs to either return a list of device errors
	// an empty list should mean that all the devices are runnning and operational

	/* Setup our running LED*/
	//r.Devices["runningLed"] = "Operational"
	r.Devices["runningLed"] = &Device{
		Name:   "runningLed",
		Status: "Operational",
	}
	runningled, runLedErr := NewLedLine(RUNNING_LED, "Running LED")
	if runLedErr != nil {
		r.Devices["runningLed"].Status = "Not Operational"
		r.Devices["runningLed"].Error = runLedErr.Error()
		// TODO: set device error list
	}
	r.runningled = runningled

	/* Setup Arm */
	r.Devices["Arm"] = &Device{
		Name:   "ArmGadget",
		Status: "Operational",
	}
	arm, err := InitArm()
	if err != nil {
		errmsg := fmt.Sprintf("Warning!! Failed to initialize arm!: %v", err)
		r.log.Print(errmsg)
		r.Devices["Arm"].Status = "Not Operational"
		r.Devices["Arm"].Error = errmsg
	}
	r.arm = arm

	if arm.IsRunning {
		r.Devices["ArmLed"] = &Device{
			Name:   "ArmLed",
			Status: "Operational",
		}

		armled, armLedErr := NewLedLine(ARM_LED, "Arm LED")
		if armLedErr != nil {
			errMsg := fmt.Sprintf("Warning!! Arm LED Failed: %v", armLedErr)
			r.log.Print(errMsg)
			r.Devices["ArmLed"].Status = "Not Operational"
			r.Devices["ArmLed"].Error = armLedErr.Error()
		}
		r.armled = armled
	}

	/* Set up pur camera */
	r.Devices["Camera"] = &Device{
		Name:   "Camera",
		Status: "Operational",
	}
	var camerr error
	r.Camera, camerr = InitCam()
	if camerr != nil {
		r.Devices["Camera"].Status = "Not Operational"
		r.Devices["Camera"].Error = camerr.Error()
		r.log.Printf("Error: Failed to initialize Camera: %v", camerr)
	}
	//defer r.Camera.Stop()
	r.Devices["Camera"].Data = map[string]interface{}{
		"Detecting":   r.Camera.DetectFaces,
		"Running":     r.Camera.IsRunning,
		"Operational": r.Camera.IsOperational,
	}

	if r.Camera.IsOperational {
		//go r.Camera.RunCamera()
		//go r.Camera.Start()

	}

	// TODO: This should be an empty list
	return nil
}

func (r *Robot) Start() (bool, error) {

	log.Println("Starting Arm and Camera...")

	if r.arm.IsRunning {
		r.armled.SetValue(1)
		if ok := r.arm.Start(); ok != nil {
			errMsg := fmt.Sprintf("Error Failed to move arm to starting position :%v", ok)
			log.Print(errMsg)
			r.Devices["ArmLed"].Error = errMsg
		}

	}

	if r.Camera.IsOperational {
		// TODO: This should probably have an error handler
		//r.Camera.DetectFaces = true
		//log.Printf("Detecting Faces")
		//go r.Camera.Start()
		log.Printf("Turning on Camera")
	}

	r.IsRunning = true
	return r.IsRunning, nil
}

func (r *Robot) Stop() (bool, error) {
	log.Println("Stoping Arm and Camera")
	r.IsRunning = false

	if r.arm.IsRunning {
		r.armled.SetValue(0)
		if ok := r.arm.Stop(); ok != nil {
			errMsg := fmt.Sprintf("Error Faild to return arm to default positon:%v", ok)
			log.Print(errMsg)
			r.Devices["ArmLed"].Error = errMsg
		}
	}

	if r.Camera.IsOperational && r.Camera.IsRunning {
		//r.Camera.Stop()
		log.Printf("Turning off Camera")
	}

	return r.IsRunning, nil
}

func (r *Robot) Reset() error { return nil }
