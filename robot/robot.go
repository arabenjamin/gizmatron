package robot



import (

	_"fmt"
	_"time"
	"log"
	"strconv"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"

)


type Robot struct {
	Name string
	State bool
	adaptor *raspi.Adaptor
	runningled *gpio.LedDriver
	arm *Arm
}

func InitRobot() (*Robot, error) {

	log.Println("Initializing Bot")
	robot := &Robot{
		Name: "Gizmatron",
		adaptor: raspi.NewAdaptor(),
	}

	//robot.arm := InitArm(robot.adaptor)
	err := robot.initDevices()
	if err != nil {

		log.Printf("%v failed to intialize device: %v", robot.Name, err)
		return nil, err
	}

	log.Println("Bot initialized")
	return robot, nil
}

func (r *Robot) initDevices() error {

	// Setup Running Led
	r.runningled = gpio.NewLedDriver(r.adaptor, strconv.Itoa(RUNNING_LED))
	r.runningled.Start()

	// Setup Arm
	arm, err := InitArm(r.adaptor)
	if err != nil {
		log.Printf("%v failed to initialize Arm: %v", r.Name, err)
		return err
	}
	r.arm = arm

	return nil
}

func (r *Robot) Start() (bool, error) {
	log.Println("starting Bot")
	r.State = true
	err := r.runningled.On()
	if err != nil {
		log.Printf("Error Turning on Led: %v", err)
	}

	r.arm.Start()

	return r.State, nil
}

func (r *Robot) Stop() (bool, error) {
	log.Println("stoping Bot")
	r.State = false
	err := r.runningled.Off()
	if err != nil {
		log.Printf("Error Turning Led Off: %v", err)
	}

	r.arm.Stop()

	return r.State, nil
}



