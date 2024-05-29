package main


import (

	"os"
	"log"
	"github.com/arabenjamin/gizmatron/server"
	"github.com/arabenjamin/gizmatron/robot"
)




func main() {

	thisLogger := log.New(os.Stdout, "http: ", log.LstdFlags)
	thisLogger.Println("Starting Gizmatron")

	/* 
		Initialize the Robot.

		The bot should NEVER fail to initialize,
		though it may initialize without the use of some components.
		This is here so we can go figure out what any other catastophic event happend. 
	*/
	bot, oops := robot.InitRobot()
	if oops != nil {
		thisLogger.Println("something real bad happened try to initialize the bot ... going down ...")
		thisLogger.Println(oops)
	}

	/*  Seems like we have a bot to work with */
	log.Printf("Robot: %v initialized", bot.Name)

	/* Strart the server */
	thisLogger.Println("Starting Gizmatron api server...")
	err := server.Start(bot)
	if err != nil{
		/* 
			Ideally the server should always be available
		*/
		thisLogger.Println("something real bad happened to the server ... going down ...")
		thisLogger.Println(err)
	}

}
