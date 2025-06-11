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
	defer arm_driver.Close()

	a := &Arm{

		driver: arm_driver,
		name:   "Gizmatron Arm",
		pins:   pins,
		joints: servos,
		x_max:  20,
		y_max:  20,
	}

	a.State = true
	a.IsRunning = true
	// set the PWM Frequency
	a.driver.SetPWMFreq(50)

	//log.Printf("JOINTS: %v", a.joints)
	//a.Stop()
	return a, nil
}

/* Update servo*/
func (a *Arm) Update(pin int, speed int) error {
	// Update this servo
	servo := a.joints[pin]
	if servo.current_degree < servo.target_degree {

		for servo.current_degree != servo.target_degree {

			time.Sleep(time.Duration(1000*speed) * time.Nanosecond)
			servo.current_degree = servo.current_degree + 1
			log.Printf("Moving Servo: on pin %v to %v ", servo.pin, int(servo.current_degree))
			err := a.driver.ServoWrite(servo.pin, int(servo.current_degree))
			if err != nil {
				log.Printf("Falied to write to servo:  Error: %v", err)
				return err
			}
		}
	} else {

		for servo.target_degree < servo.current_degree {
			time.Sleep(time.Duration(1000*speed) * time.Nanosecond)
			servo.current_degree = servo.current_degree - 1
			log.Printf("Moving Servo: on pin %v to %v ", servo.pin, int(servo.current_degree))
			err := a.driver.ServoWrite(servo.pin, int(servo.current_degree))
			if err != nil {
				log.Printf("Falied to write to servo:  Error: %v", err)
				return err
			}
		}
	}
	return nil
}

/* Put Arm in Start Position */
func (a *Arm) Start() error {

	a.joints[0].target_degree = 45
	a.joints[1].target_degree = 45
	a.joints[2].target_degree = 135
	a.joints[3].target_degree = 150

	for i, v := range a.joints {

		log.Printf("Moving Joint: %v", v)
		err := a.Update(i, 100)
		if err != nil {
			log.Printf("Failed to start arm: %v", err)
			return err
		}
	}
	return nil
}

/* Put Arm in Stop Position */
func (a *Arm) Stop() error {

	a.joints[0].target_degree = 0
	a.joints[1].target_degree = 0
	a.joints[2].target_degree = 170
	a.joints[3].target_degree = 180
	for i, v := range a.joints {

		log.Printf("Moving Joint: %v", v)
		err := a.Update(i, 100)
		if err != nil {
			log.Printf("failed to stop arm: %v", err)
			return err
		}
	}
	return nil
}

/* TODO: Impliment reset */
func (a *Arm) Reset() error { return nil }
