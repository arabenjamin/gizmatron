package robot

import (
	"log"
	"math"
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


	the distance between the middle servos is 2.8cm
	the distance between the outter servos and the inner servos is 10.4cm

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
	L1			float64 // Length of the first link
	L2			float64 // Length of the second link
	L3			float64 // Length of the third link
}

func InitArm() (*Arm, error) {


	var servos []*Servo
	s0 := NewServo(true, BASE_SERVO, 3.0)
	servos = append(servos, s0)

	s1 := NewServo(true, JOINT_1_SERVO, 10.4)
	servos = append(servos, s1)

	s2 := NewServo(true, JOINT_2_SERVO, 2.8)
	servos = append(servos, s2)

	s3 := NewServo(false, JOINT_3_SERVO, 10.4)
	servos = append(servos, s3)

	s4 := NewServo(false, JOINT_4_SERVO, 2.3)
	servos = append(servos, s4)


	// TODO: The driver should be part of the servo struct
	arm_driver, err := NewPCA9685Driver()
	if err != nil {
		log.Printf("Could not initialize arm driver: %v", err)
		return nil, err
	}
	//defer arm_driver.Close()

	a := &Arm{

		driver: arm_driver,
		Name:   "Gizmatron Arm",
		joints: servos,
		x_max:  20,
		y_max:  20,
		speed:  10 * time.Millisecond, // default speed of 10ms per degree
		L1:     10.4, // Length of the first link
		L2:     2.3,  // Length of the second link (initialize as needed)
		L3:     10.4,  // Length of the third link (initialize as needed)
	
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
	a.joints[JOINT_3_SERVO].target_degree = 180
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
func (a *Arm) UpdateArm() error {
	// Update this servo
	for _, joint := range a.joints {

		log.Printf("Setting servo %d from %d degrees to %d degrees at %d rate", joint.pin, joint.current_degree, joint.target_degree, speed)
		if err := a.driver.ServoWrite(int(joint.pin), int(joint.target_degree), a.speed); err != nil {

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
	a.joints[JOINT_1_SERVO].target_degree = 30
	a.joints[JOINT_2_SERVO].target_degree = 30
	a.joints[JOINT_3_SERVO].target_degree = 120
	a.joints[JOINT_4_SERVO].target_degree = 130

	err := a.UpdateArm()
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
	a.joints[JOINT_3_SERVO].target_degree = 180
	a.joints[JOINT_4_SERVO].target_degree = 180

	err := a.UpdateArm()
	if err != nil {
		log.Printf("failed to stop arm: %v", err)
		return err
	}
	a.IsRunning = false
	return nil
}

/* TODO: Impliment reset */
func (a *Arm) Reset() error { return nil }


func (a *Arm) SetSpeed(speed time.Duration) {
	// Set the speed for the arm movements
	a.speed = speed
	log.Printf("Arm movement speed set to %v milliseconds", a.speed)
}

func (a *Arm) MoveToTarget(x,y,z float64) error {
	
	// Clean method to move the arm to a target position

	log.Printf("Moving arm to target position: x=%.2f, y=%.2f, z=%.2f", x, y, z)
	
	// Solve the Inverse Kinematics for the arm
	err := a.SolveIK(x,y,z)
	if err != nil {
		log.Printf("Error solving IK for ShowHappy: %v", err)
		return err
	}
	
	// Update the arm with the new angles
	if err := a.UpdateArm(); err != nil {
		log.Printf("Error updating arm for ShowHappy: %v", err)
		return err
	}

	log.Printf("ShowHappy: Base Angle: %.2f, Elbow One: %.2f, Elbow Two: %.2f, Wrist Angle: %.2f",
		baseAngle, elbowOne, elbowTwo, wristAngle)
	
	return nil
}



func (a *Arm) SolveIK(x, y, z float64)  error {

	// Solve the Inverse Kinematics for the arm

	// For now we'll just assume the base servo is at 90 degrees
	// I'll solve for that later
	a.joints[BASE_SERVO].target_degree = 90

	// Calculate the base angle in radians the height should be half our z value
	baseAngleRad := math.Arcsin((z/2.0) / a.L1) 
	if math.IsNaN(baseAngleRad) {
		log.Printf("Error calculating base angle: %v", err)
		return 0, 0, 0, 0, err
	}
	baseAngle = baseAngleRad * (180 / math.Pi) // Convert to degrees
	a.joints[JOINT_1_SERVO].target_degree = baseAngle

	elbowTwo = 180 - baseAngle // The second elbow is the opposite of the base angle
    a.joints[JOINT_3_SERVO].target_degree = elbowTwo

	elbowOne = baseAngle  // should remain parallel to the ground as base angle is adjusted
	a.joints[JOINT_2_SERVO].target_degree = elbowOne

	// should also be parallel to the ground
	// The wrist angle is the sum of the elbow angles ?
	//wristAngle =  180 - (elbowOne + elbowTwo) 
	wristAngle = 180 - elbowOne // The wrist angle is the opposite of the first elbow angle	 
	a.joints[JOINT_4_SERVO].target_degree = wristAngle

	return nil
}