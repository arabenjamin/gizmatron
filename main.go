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

	// Ensure the camera is initialized
	if bot.Camera == nil {
		bot.Camera = &robot.Cam{}
	}

	/*  Seems like we have a bot to work with */
	log.Printf("Robot: %v initialized", bot.Name)

	/* Strart the server */
	serverlog.Println("Starting Gizmatron api server...")
	err := server.Start(bot, serverlog)
	if err != nil {
		/*
			Ideally the server should always be available
		*/
		serverlog.Println("something real bad happened to the server ... going down ...")
		serverlog.Println(err)
	}

}
