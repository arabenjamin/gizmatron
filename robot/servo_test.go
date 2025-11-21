package robot

import (
	"testing"
)

func TestNewServo(t *testing.T) {
	tests := []struct {
		name      string
		direction bool
		pin       int
		length    float64
		wantMin   byte
		wantMax   byte
		wantInit  byte
	}{
		{
			name:      "Forward direction servo",
			direction: true,
			pin:       0,
			length:    10.3,
			wantMin:   0,
			wantMax:   90,
			wantInit:  0,
		},
		{
			name:      "Reverse direction servo",
			direction: false,
			pin:       1,
			length:    2.8,
			wantMin:   90,
			wantMax:   180,
			wantInit:  180,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			servo := NewServo(tt.direction, tt.pin, tt.length)

			if servo == nil {
				t.Fatal("NewServo returned nil")
			}

			if servo.direction != tt.direction {
				t.Errorf("direction = %v, want %v", servo.direction, tt.direction)
			}

			if servo.pin != tt.pin {
				t.Errorf("pin = %v, want %v", servo.pin, tt.pin)
			}

			if servo.length != tt.length {
				t.Errorf("length = %v, want %v", servo.length, tt.length)
			}

			if servo.min_degree != tt.wantMin {
				t.Errorf("min_degree = %v, want %v", servo.min_degree, tt.wantMin)
			}

			if servo.max_degree != tt.wantMax {
				t.Errorf("max_degree = %v, want %v", servo.max_degree, tt.wantMax)
			}

			if servo.init_degree != tt.wantInit {
				t.Errorf("init_degree = %v, want %v", servo.init_degree, tt.wantInit)
			}

			if servo.current_degree != tt.wantInit {
				t.Errorf("current_degree = %v, want %v (should match init_degree)", servo.current_degree, tt.wantInit)
			}
		})
	}
}

func TestServoDirection(t *testing.T) {
	forwardServo := NewServo(true, 0, 10.3)
	reverseServo := NewServo(false, 1, 10.3)

	// Forward servo should have 0-90 range
	if forwardServo.min_degree != 0 || forwardServo.max_degree != 90 {
		t.Errorf("Forward servo has incorrect range: min=%d, max=%d", forwardServo.min_degree, forwardServo.max_degree)
	}

	// Reverse servo should have 90-180 range
	if reverseServo.min_degree != 90 || reverseServo.max_degree != 180 {
		t.Errorf("Reverse servo has incorrect range: min=%d, max=%d", reverseServo.min_degree, reverseServo.max_degree)
	}
}

func TestServoLinkLengths(t *testing.T) {
	linkLengths := []float64{2.0, 10.3, 2.8, 10.3, 2.0}

	for i, length := range linkLengths {
		servo := NewServo(true, i, length)
		if servo.length != length {
			t.Errorf("Servo %d: length = %v, want %v", i, servo.length, length)
		}
	}
}
