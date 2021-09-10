package drivers

// Adaptor is the interface that describes an adaptor in gobot
type Adaptor interface {
	// Name returns the label for the Adaptor
	Name() string
	// SetName sets the label for the Adaptor
	SetName(n string)
	// Connect initiates the Adaptor
	Connect() error
	// Finalize terminates the Adaptor
	Finalize() error
}

// Porter is the interface that describes an adaptor's port
type Porter interface {
	Port() string
}



// Adaptor is the Gobot Adaptor for the Raspberry Pi
type Adaptor struct {
	mutex              *sync.Mutex
	name               string
	revision           string
	digitalPins        map[int]*sysfs.DigitalPin
	pwmPins            map[int]*PWMPin
	i2cDefaultBus      int
	i2cBuses           [2]i2c.I2cDevice
	PiBlasterPeriod    uint32
}
