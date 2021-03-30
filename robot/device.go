package robot

import (

	"log"
	"strconv"
	"gobot.io/x/gobot/platforms/raspi"
	_"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/drivers/i2c"

)


const (

	RUNNING_LED = 37
	BASE_SERVO =  0
	JOINT_1_SERVO = 1
	JOINT_2_SERVO = 2
	JOINT_3_SERVO = 3
	JOINT_4_SERVO = 4

)

type Pin int

type servo struct{
	pin 			int
	direction   	bool // invert values based on the pysical direction of the servo
	min_degree  	byte
	max_degree		byte
	init_degree 	byte
	current_degree 	byte
}

func NewServo(direction bool, pin int) servo {

	// Setup Default servo configurations

	s := servo{
		direction: direction,
		pin: pin,

	}
	s.min_degree, s.max_degree, s.init_degree = 0, 0, 0
	if direction == false{

		s.min_degree, s.max_degree, s.init_degree = 180, 180, 180
	}
	s.current_degree = s.init_degree
	

	return s
}


type Arm  struct {
	adaptor *raspi.Adaptor
	name    string
	driver  *i2c.PCA9685Driver
	pins 	[]int
	joints 	[]servo
}
	

func InitArm(adaptor *raspi.Adaptor ) *Arm {

	pins := []int{
		BASE_SERVO,
		JOINT_1_SERVO,
		JOINT_2_SERVO,
		JOINT_3_SERVO,
		JOINT_4_SERVO,
	}

	
	var servos []servo
	for i:=0; i<len(pins); i++ {
		// TODO: For now I'm not setting the bottom servo
		if i > 0 && i < 3{
			s := NewServo(true, pins[i])
			servos = append(servos, s )
		}else if i >= 3 {
			s := NewServo(false, pins[i])
			servos = append(servos, s )
		}
	}

	a := &Arm{
		adaptor: adaptor,
		driver: i2c.NewPCA9685Driver(adaptor),
		name: "Gizmatron Arm",
		pins: pins,
		joints: servos,	
	}

	err :=  a.driver.Start()
	if err != nil {
		log.Printf("Could not start Arm Device: %v", err)
	}

	// set the PWM Frequency
	a.driver.SetPWMFreq(50)

	for _,v := range a.joints {
		err := a.driver.ServoWrite(strconv.Itoa(v.pin), v.init_degree)
		if err != nil {
			log.Printf("Falied to write to servo:  Error: %v", err)
		}

	}

	log.Printf("JOINTS: %v", a.joints)

	return a
}

// Move arm to x,y pos using Inverse Kinematics
func (a *Arm) Move(x int, y int){




}


func (a *Arm) Start() { 

	a.joints[0].current_degree = 45
	a.joints[1].current_degree = 45
	a.joints[2].current_degree = 135
	a.joints[3].current_degree = 135
	//log.Printf("JOINTS: %v", a.joints)
	for _,v := range a.joints{
		err := a.driver.ServoWrite(strconv.Itoa(v.pin), v.current_degree)
		if err != nil {
			log.Printf("Falied to write to servo:  Error: %v", err)
		}
	}
	
	return 
}

func (a *Arm) Stop(){ 

	a.joints[0].current_degree = 0
	a.joints[1].current_degree = 0
	a.joints[2].current_degree = 170
	a.joints[3].current_degree = 180
	//log.Printf("JOINTS: %v", a.joints)
	for _,v := range a.joints{

		err := a.driver.ServoWrite(strconv.Itoa(v.pin), v.current_degree)
		if err != nil {
			log.Printf("Falied to write to servo:  Error: %v", err)
		}
	}

	
	return 
}

func (a *Arm) Reset(){ return }