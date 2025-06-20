package robot

import (
	_ "log"
)

/*
		Our Servo and length of the link

		Each zero == the rotational axis of the servo

		                  Translated Up the y axis
		                      _  _
		                       0\_\
					             \_\
		Initial state			 0|_|
	    _ _ _  _ _               0|_|
		 0_ _||_ _|0|            /_/
		_0_ _||_ _|0|         _0/_/


		the distance between the middle servos is 2.8mm
		the distance between the outter servos and the inner servos is 10.3mm

		All of our math will be based on those lengths
*/
type Servo struct {
	driver 		   *PCA9685Driver // There is some thought that the servo driver belongs here, but currently not in use.
	pin            int //  pysical pin number on PCA9685 servo connected to
	acoords        cords
	bcoords        cords
	direction      bool // invert values based on the pysical direction of the servo
	min_degree     byte
	max_degree     byte
	init_degree    byte
	current_degree byte
	target_degree  byte
	speed          float64
	length         float64
}

type cords struct {
	x float64
	y float64
	z float64
}

func NewServo(direction bool, pin int, length float64) *Servo {

	// Setup Default servo configurations


	s := &Servo{
		direction: direction,
		pin:       pin,
		length:    length,
	}

	// set our inital values
	s.min_degree, s.max_degree, s.init_degree = 0, 90, 0
	if direction == false {
		// then our servo is oriented opposite in real space
		s.min_degree, s.max_degree, s.init_degree = 90, 180, 180
	}
	s.current_degree = s.init_degree

	return s
}
