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
	arm *Arm
	runningled *gpio.LedDriver
	//servos *i2c.PCA9685Driver
}


func InitRobot() (*Robot, error) {

	log.Println("Initializing Bot")
	robot := &Robot{

		Name: "Gizmatron",
		adaptor: raspi.NewAdaptor(),
		
	}

	//robot.arm := InitArm(robot.adaptor)
	robot.initDevices()

	log.Println("Bot initialized")

	return robot, nil
}


func (r *Robot) initDevices()  {

	// Setup Running Led
	r.runningled = gpio.NewLedDriver(r.adaptor, strconv.Itoa(RUNNING_LED))
	r.runningled.Start()

	// Setup Arm
	r.arm = InitArm(r.adaptor)
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



