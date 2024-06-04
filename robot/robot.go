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

/*
type Device Driver
*/

type device interface {
	Start()
	Stop()
}

type Robot struct {
	Name       string
	IsRunning  bool
	State      bool // depreciated
	adaptor    *raspi.Adaptor // I really want to depreciate this
	runningled *gpiocdev.Line
	serverled  *gpiocdev.Line
	armled     *gpiocdev.Line
	arm        *Arm
	camera     *Cam
	Devices    map[string]interface{}
}

func InitRobot() (*Robot, error) {

	log.Println("Initializing startup ")
	robot := &Robot{
		Name:    "Gizmatron",
		adaptor: raspi.NewAdaptor(),
		Devices: make(map[string]interface{}),
	}

	/* Start our devices*/
	r.IsRunning = true
	err := robot.initDevices()
	if err != nil {
		// TODO: This error handler needs to be rethought.
		// When initDevices() runs,
		// we want it to tell us which devices have errors, not that initDevices()
		// had errors
		log.Printf("%v failed to intialize device: %v", robot.Name, err)
		r.IsRunning = false
	}
	
	// Turn on our operating light
	r.runningled.SetValue(1)

	log.Println("Startup Complete")
	return robot, nil
}

func (r *Robot) initDevices() error {

	// TODO: When initDevices() runs,
	// we want it to return which devices have errors,
	// So this method needs to either return a list of device errors
	// an empty list should mean that all the devices are runnning and operational

	/* Setup our running LED*/
	runningled, runLedErr := robot.NewLedLine(RUNNING_LED, "Running LED")
	if runLedErr != nil {

		log.Printf("Warning !! Running LED Failec: %v", runLedErr)
		r.Devices["runningLedError"] = runLedErr
		// TODO: set device error list
	}
	r.Devices["runningLed"] = "Operational"
	r.runningled = runningled

	/* Setup Arm */
	arm, err := InitArm(r.adaptor)
	if err != nil {
		errmsg = fmt.Sprintf("Warning!! Arm Initialization Failed!: %v", err)
		log.Printf(errmsg)
		// TODO Set the arm error in the device status
		r.Devices["armError"] = errmsg
	}
	r.arm = arm

	if arm.IsRunning {

		armled, armLedErr := robot.NewLedLine(ARM_LED, "Arm LED")
		if armLedErr != nil {
			errMsg = fmt.Sprintf("Warning!! Arm LED Failed: %v", armLedErr)
			log.Printf(errMsg)
			r.Devices["ArmLedError"] = armLedErr
		}
		r.Devices["ArmLed"] = "Operational"
		r.armled = armled
	}

	/* Set up pur camera */
	cam, camerr := InitCam()
	if camerr != nil {
		errMsg = fmt.Sprintf("Warning !! Camera Initialization Failed: %v", camerr)
		log.Printf(errMsg)
		r.Devices["CameraError"] = camerr
	}

	if cam.IsRunning {
		r.Devices["Camera"] = "Operational"
		r.camera = cam
	}

	// TODO: This should be an empty list
	return nil
}

func (r *Robot) Start() (bool, error) {
	
	log.Println("Starting Arm and Camera...")
	
	if r.arm.IsRunning {
		r.armled.SetValue(1)
		err = r.arm.Start(); !ok {
			errMsg = fmt.Sprintf("Error Failed to move arm to starting position :%v", err)
			log.Printf(errMsg)
			r.Devices["ArmError"] = errMsg
		}

	}

	if r.camera.IsRunning {
		// TODO: This should probably have an error handler
		go r.camera.Start()
	}

	r.IsRunning = true
	return r.IsRunning, nil
}

func (r *Robot) Stop() (bool, error) {
	log.Println("Stoping Arm and Camera")
	r.IsRunning = false

	if r.arm.IsRunning {
		r.armled.SetValue(0)
		err = r.arm.Stop(); !ok {
			errMsg = fmt.Sprintf("Error Faild to return arm to default positon:%v", err)
			log.Printf(errMsg)
			r.Devices["ArmError"] = errMsg
		}
	}
	

	return r.IsRunning, nil
}

func (r *Robot) Reset() error { return nil }
