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
	err           error
	State         bool
	IsRunning     bool
	IsOperational bool
	Name          string
	driver        *PCA9685Driver
	joints        []*Servo
	x_max         int
	y_max         int
	speed         time.Duration // speed in ms
}

func InitArm() (*Arm, error) {


	var servos []*Servo
	s0 := NewServo(true, BASE_SERVO, 2.0)
	servos = append(servos, s0)

	s1 := NewServo(true, JOINT_1_SERVO, 10.3)
	servos = append(servos, s1)

	s2 := NewServo(true, JOINT_2_SERVO, 2.8)
	servos = append(servos, s2)

	s3 := NewServo(false, JOINT_3_SERVO, 10.3)
	servos = append(servos, s3)

	s4 := NewServo(false, JOINT_4_SERVO, 2.0)
	servos = append(servos, s4)


	// TODO: The driver should be part of the servo struct
	arm_driver, err := NewPCA9685Driver()
	if err != nil {
		log.Printf("Could not initialize arm driver: %v", err)
		return nil, err
	}
	defer arm_driver.Close()

	a := &Arm{

		driver: arm_driver,
		Name:   "Gizmatron Arm",
		joints: servos,
		x_max:  20,
		y_max:  20,
		speed:  100,
	}

	// set the PWM Frequency
	//a.driver.SetPWMFreq(50)
	log.Println("Setting PWM frequency to 50Hz...")
	if err := a.driver.SetPWMFreq(50); err != nil {
		log.Fatalf("Could not set PWM frequency: %v", err)
	}

	// Set initial angles for servos
	a.joints[BASE_SERVO].target_degree = 90
	a.joints[JOINT_1_SERVO].target_degree = 0
	a.joints[JOINT_2_SERVO].target_degree = 0
	a.joints[JOINT_3_SERVO].target_degree = 170
	a.joints[JOINT_4_SERVO].target_degree = 180

	for _, joint := range a.joints {
		log.Printf("Setting initial angle for servo %d to %d degrees", joint.pin, joint.target_degree)
		if err := a.driver.setServoPulse(joint.pin, int(joint.target_degree)); err != nil {
			log.Printf("Error setting initial servo position: %v\n", err)
		}
		//time.Sleep(time.Duration(1000*100) * time.Nanosecond)
	}

	a.driver.currentAngles[BASE_SERVO] = int(a.joints[BASE_SERVO].target_degree)
	a.driver.currentAngles[JOINT_1_SERVO] = int(a.joints[JOINT_1_SERVO].target_degree)
	a.driver.currentAngles[JOINT_2_SERVO] = int(a.joints[JOINT_2_SERVO].target_degree)
	a.driver.currentAngles[JOINT_3_SERVO] = int(a.joints[JOINT_3_SERVO].target_degree)
	a.driver.currentAngles[JOINT_4_SERVO] = int(a.joints[JOINT_4_SERVO].target_degree)

	log.Println("Arm Position: ", a.driver.currentAngles)
	a.State = true
	a.IsOperational = true
	return a, nil
}

/* Update servo*/
func (a *Arm) UpdateArm(speed time.Duration) error {
	// Update this servo
	for _, joint := range a.joints {

		log.Printf("Setting servo %d from %d degrees to %d degrees at %d rate", joint.pin, joint.current_degree, joint.target_degree, speed)
		if err := a.driver.ServoWrite(int(joint.pin), int(joint.target_degree), speed); err != nil {

			// TODO: Keep track of servos that fail to move
			// and return an error at the end of the function
			// so that we can retry them later
			log.Printf("Error! moving servo: %v\n", err)
			return err
		}
		joint.current_degree = joint.target_degree
		log.Printf("Joint %d current degree: %d", joint.pin, joint.current_degree)
		//time.Sleep(time.Duration(1000*speed) * time.Nanosecond)

	}

	return nil
}

/* Put Arm in Start Position */
func (a *Arm) Start() error {

	log.Println("Starting Arm...")

	a.joints[BASE_SERVO].target_degree = 90
	a.joints[JOINT_1_SERVO].target_degree = 45
	a.joints[JOINT_2_SERVO].target_degree = 45
	a.joints[JOINT_3_SERVO].target_degree = 135
	a.joints[JOINT_4_SERVO].target_degree = 150

	err := a.UpdateArm(10)
	if err != nil {
		log.Printf("Failed to start arm: %v", err)
		return err
	}

	a.IsRunning = true
	return nil
}

/* Put Arm in Stop Position */
func (a *Arm) Stop() error {

	a.joints[BASE_SERVO].target_degree = 90
	a.joints[JOINT_1_SERVO].target_degree = 0
	a.joints[JOINT_2_SERVO].target_degree = 0
	a.joints[JOINT_3_SERVO].target_degree = 170
	a.joints[JOINT_4_SERVO].target_degree = 180

	err := a.UpdateArm(10)
	if err != nil {
		log.Printf("failed to stop arm: %v", err)
		return err
	}
	a.IsRunning = false
	return nil
}

/* TODO: Impliment reset */
func (a *Arm) Reset() error { return nil }
