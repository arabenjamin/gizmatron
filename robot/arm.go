package robot

import (
	"log"
	_ "math"
	"strconv"
	"time"

	"gobot.io/x/gobot/v2/drivers/i2c"
	"gobot.io/x/gobot/v2/platforms/raspi"
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
	err     error
	adaptor *raspi.Adaptor
	name    string
	driver  *i2c.PCA9685Driver
	pins    []int
	joints  []*Servo
	x_max   int
	y_max   int
}

func InitArm(adaptor *raspi.Adaptor) (*Arm, error) {

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

	a := &Arm{
		adaptor: adaptor,
		driver:  i2c.NewPCA9685Driver(adaptor),
		name:    "Gizmatron Arm",
		pins:    pins,
		joints:  servos,
		x_max:   20,
		y_max:   20,
	}

	err := a.driver.Start()
	if err != nil {
		log.Printf("Could not start Arm Device: %v", err)
		a.err = err
		return a, err
	}

	// set the PWM Frequency
	a.driver.SetPWMFreq(50)

	for _, v := range a.joints {
		// TODO: Update this to use the Update method
		err := a.driver.ServoWrite(strconv.Itoa(v.pin), v.init_degree)
		if err != nil {
			a.err = err
			log.Printf("Falied to write to servo:  Error: %v", err)
			return a, err
		}
	}
	//log.Printf("JOINTS: %v", a.joints)
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
			err := a.driver.ServoWrite(strconv.Itoa(servo.pin), servo.current_degree)
			if err != nil {
				log.Printf("Falied to write to servo:  Error: %v", err)
				return err
			}
		}
	} else {

		for servo.target_degree < servo.current_degree {
			time.Sleep(time.Duration(1000*speed) * time.Nanosecond)
			servo.current_degree = servo.current_degree - 1
			err := a.driver.ServoWrite(strconv.Itoa(servo.pin), servo.current_degree)
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
		err := a.Update(i, 2000)
		if err != nil {
			log.Println("failed to stop arm: %v", err)
			return err
		}
	}
	return nil
}

/* TODO: Impliment reset */
func (a *Arm) Reset() error { return nil }
