package robot

import (
	"fmt"
	"time"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// Constants for PCA9685 registers
const (
	PCA9685_ADDRESS  = 0x40 // Default I2C address
	PCA9685_MODE1    = 0x00
	PCA9685_PRESCALE = 0xFE
	LED0_ON_L        = 0x06
)

// PCA9685Driver represents our custom driver
type PCA9685Driver struct {
	dev           *i2c.Dev
	bus           i2c.BusCloser
	currentAngles [5]int // Keep track of the last angle for each channel
}

// NewPCA9685Driver initializes the I2C bus and connects to the PCA9685 device.
func NewPCA9685Driver() (*PCA9685Driver, error) {
	// Initialize the host hardware. This is a required step for periph.io.
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize host: %w", err)
	}

	// Open the default I2C bus.
	bus, err := i2creg.Open("")
	if err != nil {
		return nil, fmt.Errorf("failed to open I2C bus: %w", err)
	}

	// Create a new device object for communication.
	dev := &i2c.Dev{Addr: PCA9685_ADDRESS, Bus: bus}

	driver := &PCA9685Driver{
		dev: dev,
		bus: bus,
	}
	// Initialize all angles to a neutral position (e.g., 90 degrees)
	for i := range driver.currentAngles {
		driver.currentAngles[i] = 90
	}

	// Reset the device to a known state.
	if err := driver.writeRegister(PCA9685_MODE1, 0x00); err != nil {
		bus.Close()
		return nil, err
	}

	// Allow time for the oscillator to stabilize.
	time.Sleep(1 * time.Millisecond)

	return driver, nil
}

// Close cleans up and closes the I2C bus connection.
func (d *PCA9685Driver) Close() {
	d.bus.Close()
}

// writeRegister is a helper to write a byte to a specific register.
func (d *PCA9685Driver) writeRegister(reg, value byte) error {
	_, err := d.dev.Write([]byte{reg, value})
	return err
}

// SetPWMFreq sets the PWM frequency for all channels.
func (d *PCA9685Driver) SetPWMFreq(freq float64) error {
	// Calculate the prescale value.
	prescaleVal := 25000000.0 // 25MHz internal oscillator
	prescaleVal /= 4096.0     // 12-bit
	prescaleVal /= freq
	prescaleVal -= 1.0
	prescale := byte(prescaleVal + 0.5) // Round to nearest int

	// To set the prescaler, the chip must be in sleep mode.
	oldMode := make([]byte, 1)
	if err := d.dev.Tx([]byte{PCA9685_MODE1}, oldMode); err != nil {
		return err
	}

	// Set the sleep bit (bit 4) in MODE1.
	newMode := (oldMode[0] & 0x7F) | 0x10 // 0x7F is ~0x80
	if err := d.writeRegister(PCA9685_MODE1, newMode); err != nil {
		return err
	}

	// Write the prescale value.
	if err := d.writeRegister(PCA9685_PRESCALE, prescale); err != nil {
		return err
	}

	// Restore the original mode to wake the chip up.
	if err := d.writeRegister(PCA9685_MODE1, oldMode[0]); err != nil {
		return err
	}

	// It's recommended to wait a short time for the oscillator to stabilize.
	time.Sleep(5 * time.Millisecond)

	// Set MODE1 to turn on auto-increment.
	if err := d.writeRegister(PCA9685_MODE1, oldMode[0]|0xa0); err != nil {
		return err
	}

	return nil
}

// SetPWM sets the on and off time for a single PWM channel.
func (d *PCA9685Driver) SetPWM(channel int, on, off uint16) error {
	if channel < 0 || channel > 15 {
		return fmt.Errorf("channel out of range (0-15)")
	}
	regBase := byte(LED0_ON_L + 4*channel)
	data := []byte{
		regBase,
		byte(on),
		byte(on >> 8),
		byte(off),
		byte(off >> 8),
	}
	_, err := d.dev.Write(data)
	return err
}

// setServoPulse is an internal helper that converts an angle to a PWM pulse and sets it instantly.
func (d *PCA9685Driver) setServoPulse(channel int, angle int) error {
	if angle < 0 || angle > 180 {
		return fmt.Errorf("angle out of range (0-180)")
	}
	// These are typical pulse widths for a standard servo.
	servoMin := 150 // Min pulse length (out of 4096)
	servoMax := 600 // Max pulse length (out of 4096)

	pulseLength := servoMin + int(float64(servoMax-servoMin)*float64(angle)/180.0)

	return d.SetPWM(channel, 0, uint16(pulseLength))
}

// ServoWrite moves a servo to a specific angle at a given speed.
// Speed is the delay in milliseconds between each 1-degree step.
// Smaller speed value means faster movement.
func (d *PCA9685Driver) ServoWrite(channel int, angle int, speed time.Duration) error {
	if channel < 0 || channel > 15 {
		return fmt.Errorf("channel out of range (0-15)")
	}
	if angle < 0 || angle > 180 {
		return fmt.Errorf("angle out of range (0-180)")
	}

	startAngle := d.currentAngles[channel]
	endAngle := angle

	// Determine the direction of movement
	if startAngle < endAngle {
		// Move from start to end
		for i := startAngle; i <= endAngle; i++ {
			if err := d.setServoPulse(channel, i); err != nil {
				return err
			}
			time.Sleep(speed * time.Millisecond)
		}
	} else {
		// Move from start to end (in reverse)
		for i := startAngle; i >= endAngle; i-- {
			if err := d.setServoPulse(channel, i); err != nil {
				return err
			}
			time.Sleep(speed * time.Millisecond)
		}
	}

	// Update the current angle for the channel
	d.currentAngles[channel] = angle
	return nil
}
