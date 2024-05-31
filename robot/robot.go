package robot

import (
	"log"
	_ "time"

	"github.com/warthog618/go-gpiocdev"
	_ "gobot.io/x/gobot/v2"
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
	State      bool
	adaptor    *raspi.Adaptor
	runningled *gpiocdev.Line
	//serverled  *gpiocdev.Line
	//armled     *gpiocdev.Line
	arm     *Arm
	Devices map[string]interface{}
}

func InitRobot() (*Robot, error) {

	log.Println("Initializing startup ")
	robot := &Robot{
		Name:    "Gizmatron",
		adaptor: raspi.NewAdaptor(),
		Devices: make(map[string]interface{}),
	}

	//robot.arm := InitArm(robot.adaptor)
	err := robot.initDevices()
	if err != nil {

		log.Printf("%v failed to intialize device: %v", robot.Name, err)
		// return nil, err
	}

	log.Println("Startup Complete")
	return robot, nil
}

func (r *Robot) initDevices() error {

	//r.runningled, runningLedErr = gpiocdev.RequestLine("gpiochip0", RUNNING_LED, gpiocdev.AsOutput(0))
	r.runningled, _ = gpiocdev.RequestLine("gpiochip0", RUNNING_LED, gpiocdev.AsOutput(0))

	// Setup Running Led ( Green LED on pin 37 )

	/*
		r.runningled = gpio.NewLedDriver(r.adaptor, strconv.Itoa(RUNNING_LED))
		r.runningled.Start()
		runningLederr := r.runningled.On()
		if runningLederr != nil {
			log.Printf("Error Turning on Running LED: %v", runningLederr)
			r.Devices["runningLedError"] = runningLederr
		}*/

	/*
		//Setup Server LED ( Blue LED on pin ...)
		r.serverled = gpio.NewLedDriver(r.adaptor, strconv.Itoa(SEVER_LED))
		r.serverled.Start()
		serverErr := r.serverled.On()
		if serverErr != nil {
			log.Printf("Error Turning on Server LED: %v", serverErr)
			r.Devices["severledError"] = serverErr
		}

		//Setup Arm LED ( White LED on pin ...)
		r.armled = gpio.NewLedDriver(r.adaptor, strconv.Itoa(ARM_LED))
		r.armled.Start()
		armErr := r.armled.On()
		if armErr != nil {
			log.Printf("Error Turning on arm LED: %v", armErr)
			r.Devices["armLEDError"] = armErr
		}
	*/

	// Setup Arm
	arm, err := InitArm(r.adaptor)
	if err != nil {
		log.Printf("%v failed to initialize Arm: %v", r.Name, err)
		// TODO Set the arm error in the device status
		r.Devices["armError"] = err
		//return err
	}
	r.arm = arm
	return nil
}

func (r *Robot) Start() (bool, error) {
	log.Println("starting Bot")
	r.State = true

	err := r.runningled.SetValue(1)
	if err != nil {

		//log.Printf("Error Turning on Led: %v", err)
		r.State = false
		log.Printf("Error Turning on Running LED: %v", err)
		r.Devices["runningLedError"] = err
		return r.State, err
	}

	r.arm.Start()
	return r.State, nil
}

func (r *Robot) Stop() (bool, error) {
	log.Println("stoping Bot")
	r.State = false

	err := r.runningled.SetValue(0)
	if err != nil {
		log.Printf("Error Turning Led Off: %v", err)
		return false, err
	}

	r.arm.Stop()
	return r.State, nil
}

func (r *Robot) Reset() error { return nil }
