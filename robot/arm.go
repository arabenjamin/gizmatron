package robot

import (
	"log"
	_ "math"
	"time"
)

/*
	An Arm is really a collection of joints and lengths
	Here we will use the servo to act as a joint and a length
	And so an Arm in this case is a collection of servos.

*/

/*
	Our Servo and length of the link

	Each zero == the rotational axis of the servo

	          Translated Up the y axis
	          End Efector _  _
	                       0\_\
				             \_\
	Initial state			 0|_|
    _ _ _  _ _               0|_|
	 0_ _||_ _|0|            /_/
	_0_ _||_ _|0|    Base _0/_/


	the distance between the middle servos is 2.8mm
	the distance between the outter servos and the inner servos is 10.3mm

	All of our math will be based on those lengths


*/

type Arm struct {
	err       error
	State     bool
	IsRunning bool
	name      string
	driver    *PCA9685Driver
	pins      []int
	joints    []*Servo
	x_max     int
	y_max     int
}

func InitArm() (*Arm, error) {

	pins := []int{
		BASE_SERVO,
		JOINT_1_SERVO,
		JOINT_2_SERVO,
		JOINT_3_SERVO,
		JOINT_4_SERVO,
	}

	var servos []*Servo

	s1 := NewServo(true, pins[1], 10.3)
	servos = append(servos, s1)

	s2 := NewServo(true, pins[2], 2.8)
	servos = append(servos, s2)

	s3 := NewServo(false, pins[3], 10.3)
	servos = append(servos, s3)

	s4 := NewServo(false, pins[4], 2.0)
	servos = append(servos, s4)

	arm_driver, err := NewPCA9685Driver()
	if err != nil {
		log.Printf("Could not initialize arm driver: %v", err)
		return nil, err
	}
	//defer arm_driver.Close()

	a := &Arm{

		driver: arm_driver,
		name:   "Gizmatron Arm",
		pins:   pins,
		joints: servos,
		x_max:  20,
		y_max:  20,
	}

	// set the PWM Frequency
	//a.driver.SetPWMFreq(50)
	log.Println("Setting PWM frequency to 50Hz...")
	if err := a.driver.SetPWMFreq(50); err != nil {
		log.Fatalf("Could not set PWM frequency: %v", err)
	}

	// Set initial angles for servos
	a.driver.currentAngles[BASE_SERVO] = 90
	a.driver.currentAngles[JOINT_1_SERVO] = 0
	a.driver.currentAngles[JOINT_2_SERVO] = 0
	a.driver.currentAngles[JOINT_3_SERVO] = 170
	a.driver.currentAngles[JOINT_4_SERVO] = 180

	for i, servo := range a.driver.currentAngles {
		log.Printf("Setting initial angle for servo %d to %d degrees", i, servo)
		if err := a.driver.setServoPulse(i, int(servo)); err != nil {
			log.Printf("Error setting initial servo position: %v\n", err)
		}
	}

	a.State = true
	a.IsRunning = true

	//log.Printf("JOINTS: %v", a.joints)
	//a.Stop()
	return a, nil
}

/* Update servo*/
func (a *Arm) Update(pin int, speed time.Duration) error {
	// Update this servo

	for i, angle := range a.driver.currentAngles {
		log.Printf("Setting servo %d to %d degrees", i, angle)
		if err := a.driver.ServoWrite(i, angle, speed); err != nil {
			log.Printf("Error moving servo: %v\n", err)
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

/* Put Arm in Start Position */
func (a *Arm) Start() error {

	log.Println("Starting Arm...")

	a.driver.currentAngles[BASE_SERVO] = 90
	a.driver.currentAngles[JOINT_1_SERVO] = 45
	a.driver.currentAngles[JOINT_2_SERVO] = 45
	a.driver.currentAngles[JOINT_3_SERVO] = 135
	a.driver.currentAngles[JOINT_4_SERVO] = 150

	for i, angle := range a.driver.currentAngles {

		log.Printf("Moving Joint: %v to %v", i, angle)
		err := a.Update(i, 15)
		if err != nil {
			log.Printf("Failed to start arm: %v", err)
			return err
		}
	}
	return nil
}

/* Put Arm in Stop Position */
func (a *Arm) Stop() error {

	a.driver.currentAngles[BASE_SERVO] = 90
	a.driver.currentAngles[JOINT_1_SERVO] = 0
	a.driver.currentAngles[JOINT_2_SERVO] = 0
	a.driver.currentAngles[JOINT_3_SERVO] = 170
	a.driver.currentAngles[JOINT_4_SERVO] = 180

	for i, v := range a.driver.currentAngles {

		log.Printf("Moving Joint: %v", v)
		err := a.Update(i, 15)
		if err != nil {
			log.Printf("failed to stop arm: %v", err)
			return err
		}
	}
	return nil
}

/* TODO: Impliment reset */
func (a *Arm) Reset() error { return nil }
