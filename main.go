package main


import (

	"os"
	"log"
	"fmt"
	"github.com/arabenjamin/gizmatron/server"
	"github.com/arabenjamin/gizmatron/robot"
)




func main() {
	/* Strart the server */
	thisLogger := log.New(os.Stdout, "http: ", log.LstdFlags)
	thisLogger.Println("Starting Gizmatron api server...")

	
	//thisLogger.Println(robot.LED_PIN)
	//robot.Blink()
	fmt.Println(robot.RobotName())
	thisLogger.Println(robot.RobotName())
	//go robot.BlinkFromPCA()

	bot, oops := robot.InitRobot()
	if oops != nil {

		thisLogger.Println("something real bad happened to the bot ... going down ...")
		thisLogger.Println(oops)
	}

	err := server.Start(bot)
	if err != nil{
		thisLogger.Println("something real bad happened ... going down ...")
		thisLogger.Println(err)
	}

	

}
