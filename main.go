package main

import (
	"log"
	"os"

	"github.com/arabenjamin/gizmatron/robot"
	"github.com/arabenjamin/gizmatron/server"
)

func main() {

	serverlog := log.New(os.Stdout, "http: ", log.LstdFlags)
	log.Println("Starting Gizmatron")
	log.Println("This Robot sucks")
	log.Println("Never trust a droid")
	log.Println("Why let a droid do what a man can do !!")

	/*
		Initialize the Robot.

		The bot should NEVER fail to initialize,
		though it may initialize without the use of some components.
		This is here so we can go figure out what any other catastophic event happend.
	*/
	bot, oops := robot.InitRobot()
	if oops != nil {
		log.Println("something real bad happened try to initialize the bot ... going down ...")
		log.Println(oops)
	}

	/*  Seems like we have a bot to work with */
	log.Printf("Robot: %v initialized", bot.Name)

	//Setup Server LED ( Blue LED on pin ...)
	bot.Devices["serverLed"] = "Operational"
	serverled, serverErr := robot.NewLedLine(13, "Sever Led")
	if serverErr != nil {
		log.Printf("Error Turning on Server LED: %v", serverErr)
		bot.Devices["severledError"] = serverErr
		bot.Devices["serverLed"] = "NOT Operational"
	}
	bot.Serverled = serverled
	// Turn the server led on now
	// I may want to rethink the way the server light comes on.
	if bot.Devices["severLed"] == "Operational" {
		bot.Serverled.SetValue(1)
	}

	/* Strart the server */
	serverlog.Println("Starting Gizmatron api server...")
	err := server.Start(bot, serverlog)
	if err != nil {
		/*
			Ideally the server should always be available
		*/
		serverlog.Println("something real bad happened to the server ... going down ...")
		serverlog.Println(err)
		bot.Serverled.SetValue(0)
	}

}
