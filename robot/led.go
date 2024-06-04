package robot


import (

	"log"
	"github.com/warthog618/go-gpiocdev"

)




type Leds struct {

	chip *gpiocdev.Chip

}




func NewLedLine( pin int, label string) (*gpiocdev.Line, error){

	// Set the gpio pin to output low for now
	// I'm just assuming that I'll never need a different chip
	line , err := gpiocdev.RequestLine("gpiochip0", pin, gpiocdev.AsOutput(0), gpiocdev.WithConsumer(label))
	if err != nil {
		log.Printf("Error Turning on Led: %v", err)
		
		return nil, err
	}
	return line, nil
}


