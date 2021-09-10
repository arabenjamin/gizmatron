package robot


import (

	_"fmt"
	"time"
	"log"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)




var LED_PIN = "5"


func BlinkFromPCA(){

	rasp_adapt := raspi.NewAdaptor()
	log.Println("adapter name: ", rasp_adapt.Name())
	
	device := i2c.NewPCA9685Driver(rasp_adapt)
	log.Printf("PCA9685 Device: %v ", device.Name())

	//log.Printf("Connection: %+v", device.Connection().Name())
	//log.Println("Starting device")
	err := device.Start()
	if err != nil {

		log.Printf("Could not start Device: %v", err)
	}
	log.Printf("Connection: %+v ", device.Connection().Name())
	device.SetPWMFreq(100)
	//device.SetAllPWM(0, 0)
	
	
	for {
		log.Printf("Rotating Servo ... ")
		serr := device.SetPWM(13, 0, uint16(255))
		if serr != nil{
			log.Printf("Failed to Move Servo: %v", serr)
		}
	
		time.Sleep(2*time.Second)
		serr = device.SetPWM(13, uint16(255), 0)
		if serr != nil{
			log.Printf("Failed to Move Servo: %v", serr)
		}
		
		time.Sleep(1*time.Second)

	}
	/*
	
	for {
		log.Printf("Blinking Led on %v", device.Name())
		log.Printf("Turning OFF")
		serr := device.ServoWrite("13", 90)
		//serr := device.SetPWM(13, 0, uint16(255))
		if serr != nil{
			log.Printf("Failed to Move Servo: %v", serr)
		}
		err = device.SetPWM(0, 0, uint16(255))
		//err := device.PwmWrite("0", 255)
		if err != nil {
			log.Printf("Failed to Write to device: %v", err)
		}
		time.Sleep(1*time.Second)

		serr = device.ServoWrite("13", 180)
		//serr = device.SetPWM(13, 0, uint16(255))
		if serr != nil{
			log.Printf("Failed to Move Servo: %v", serr)
		}
		
		log.Printf("Turning ON")
		err := device.SetPWM(0, uint16(255), 0 )
		//err = device.PwmWrite("0", 0)
		if err != nil {
			log.Printf("Failed to Write to device: %v", err)
		}
		time.Sleep(10*time.Millisecond)

	}*/
	

}

func Blink(){
	rasp_adapt := raspi.NewAdaptor()

	log.Println("adapter name: ", rasp_adapt.Name())
	

	led := gpio.NewLedDriver(rasp_adapt, LED_PIN)


	log.Printf("Led: %v %v ", led.Name(), led.Pin())
	//led.Connection().Connect()
	led.Start()
	log.Printf("Connection: %+v ", led.Connection().Name())

	for {
		time.Sleep(1*time.Second)
		log.Printf("Toggling Led State: %v ", led.State())
		led.Toggle()
	}

	
}

func BlinkBot(){
	r := raspi.NewAdaptor()
	led := gpio.NewLedDriver(r, LED_PIN)

	work := func() {
        /*gobot.Every(1*time.Second, func() {
            led.Toggle()
		})*/
		for {
			time.Sleep(1*time.Second)
			log.Printf("Toggling Led State: %v ", led.State())
			led.Toggle()
		}
		
    }
	
	log.Printf("blinking ...")
    robot := gobot.NewRobot("blinkBot",
        []gobot.Connection{r},
        []gobot.Device{led},
        work,
	)
	//log.Printf("%v", robot.)
	err := robot.Start()
	if err != nil {
		log.Printf("BlinkBot Error: %v", err )
	}

}
