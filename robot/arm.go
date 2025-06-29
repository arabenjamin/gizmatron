package robot

import (
	"fmt"
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

const (
	BASE_SERVO    = 0
	JOINT_1_SERVO = 1
	JOINT_2_SERVO = 2
	JOINT_3_SERVO = 3
	JOINT_4_SERVO = 4
)

type cords struct {
	x float64
	y float64
	z float64
}

type Arm struct {
	err               error
	State             bool
	IsRunning         bool
	IsOperational     bool
	Name              string
	driver            *PCA9685Driver
	x_max             int
	y_max             int
	speed             time.Duration // speed in ms
	L1                float64       // Length of the first link
	L2                float64       // Length of the second link
	L3                float64       // Length of the third link
	jointTargetAngles [5]int        // Target degrees for each joint
}

func InitArm() (*Arm, error) {

	// TODO: The driver should be part of the servo struct
	arm_driver, err := NewPCA9685Driver()
	if err != nil {
		log.Printf("Could not initialize arm driver: %v", err)
		return nil, err
	}
	//defer arm_driver.Close()

	a := &Arm{

		Name:   "Gizmatron Arm",
		driver: arm_driver,
		x_max:  20,
		y_max:  20,
		speed:  10,   // default speed of 10ms per degree
		L1:     10.4, // Length of the first link
		L2:     2.8,  // Length of the second link (initialize as needed)
		L3:     10.4, // Length of the third link (initialize as needed)
		jointTargetAngles: [5]int{ // Initial target angles
			90,  // BASE_SERVO 3.0cm
			0,   // JOINT_1_SERVO 10.4cm
			0,   // JOINT_2_SERVO 2.8cm
			180, // JOINT_3_SERVO 10.4cm
			180, // JOINT_4_SERVO 2.3cm
		},
	}

	// set the PWM Frequency
	log.Println("Setting PWM frequency to 50Hz...")
	if err := a.driver.SetPWMFreq(50); err != nil {
		log.Fatalf("Could not set PWM frequency: %v", err)
	}

	for i, degree := range a.jointTargetAngles {
		if err := a.driver.setServoPulse(i, degree); err != nil {
			log.Printf("Error setting initial servo position for servo %d: %v\n", i, err)
		}
		// setServoPulse will not update the current angles,
		// so we need to manually set them here.
		// This is important for the first run to ensure the arm starts at the correct position.
		log.Printf("Setting initial angle for servo %d to %d degrees", i, degree)
		a.driver.currentAngles[i] = degree // Initialize current angles
	}

	log.Println("Arm Position: ", a.driver.currentAngles)
	a.State = true
	a.IsOperational = true
	return a, nil
}

func (a *Arm) SetSpeed(speed time.Duration) {
	// Set the speed for the arm movements
	a.speed = speed
	log.Printf("Arm movement speed set to %v milliseconds", a.speed)
}

/* Update servo*/
func (a *Arm) UpdateArm() error {
	// Update this servo
	for i, degree := range a.jointTargetAngles {

		log.Printf("Setting servo %d from %d degrees to %d degrees at %d rate", i, a.driver.currentAngles[i], degree, a.speed)
		if err := a.driver.ServoWrite(i, int(degree), a.speed); err != nil {

			// TODO: Keep track of servos that fail to move
			// and return an error at the end of the function
			// so that we can retry them later
			log.Printf("Error! moving servo: %v\n", err)
			return err
		}
		log.Printf("Joint %d current degree: %d", i, a.driver.currentAngles[i])
		//time.Sleep(time.Duration(1000*speed) * time.Nanosecond)

	}

	return nil
}

/* Put Arm in Start Position */
func (a *Arm) Start() error {

	log.Println("Starting Arm...")

	a.jointTargetAngles = [5]int{
		90,  // BASE_SERVO
		30,  // JOINT_1_SERVO
		30,  // JOINT_2_SERVO
		120, // JOINT_3_SERVO
		130, // JOINT_4_SERVO
	}

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

	log.Println("Stopping Arm...")

	// Set the arm to a default position
	// This is the position we want the arm to be in when it is not running
	// It should be a safe position that does not interfere with any objects
	// or cause any damage to the arm or the environment
	a.jointTargetAngles = [5]int{
		90,  // BASE_SERVO
		0,   // JOINT_1_SERVO
		0,   // JOINT_2_SERVO
		180, // JOINT_3_SERVO
		180, // JOINT_4_SERVO
	}
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

func (a *Arm) MoveToTarget(x, y, z float64) error {

	// Clean method to move the arm to a target position

	log.Printf("Moving arm to target position: x=%.2f, y=%.2f, z=%.2f", x, y, z)

	// Solve the Inverse Kinematics for the arm
	err := a.SolveIK(x, y, z)
	if err != nil {
		log.Printf("Error solving IK for ShowHappy: %v", err)
		return err
	}

	// Update the arm with the new angles
	if err := a.UpdateArm(); err != nil {
		log.Printf("Error updating arm for ShowHappy: %v", err)
		return err
	}
	log.Println("Arm moved to target position successfully")

	return nil
}

func (a *Arm) SolveIK(x, y, z float64) error {

	// Solve the Inverse Kinematics for the arm

	// For now we'll just assume the base servo is at 90 degrees
	// I'll solve for that later
	a.jointTargetAngles[BASE_SERVO] = 90 // Set the base servo to 90 degrees

	// Calculate the base angle in radians the height should be half our z value
	baseAngleRad := math.Asin((z / 2.0) / a.L1)

	if math.IsNaN(baseAngleRad) {
		log.Printf("Error calculating base angle")
		return fmt.Errorf("invalid base angle calculation")
	}

	baseAngle := baseAngleRad * (180 / math.Pi) // Convert to degrees
	a.jointTargetAngles[JOINT_1_SERVO] = int(baseAngle)

	elbowTwo := 180 - baseAngle // The second elbow is the opposite of the base angle
	a.jointTargetAngles[JOINT_3_SERVO] = int(elbowTwo)

	elbowOne := baseAngle // should remain parallel to the ground as base angle is adjusted
	a.jointTargetAngles[JOINT_2_SERVO] = int(elbowOne)
	// should also be parallel to the ground
	// The wrist angle is the sum of the elbow angles ?
	//wristAngle =  180 - (elbowOne + elbowTwo)
	wristAngle := 180 - elbowOne // The wrist angle is the opposite of the first elbow angle
	a.jointTargetAngles[JOINT_4_SERVO] = int(wristAngle)

	return nil
}
