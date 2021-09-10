package robot



import (

	_"time"
	"log"
	"strconv"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

const (

	RUNNING_LED = 37
	BASE_SERVO =  0
	JOINT_1_SERVO = 1
	JOINT_2_SERVO = 2
	JOINT_3_SERVO = 3
	JOINT_4_SERVO = 4

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
		r.State = false
		return r.State, err
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
		return false, err
	}

	r.arm.Stop()
	return r.State, nil
}

func (r *Robot) Reset() error { return nil }

