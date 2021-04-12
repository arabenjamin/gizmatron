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
	bot, oops := robot.InitRobot()
	if oops != nil {

		thisLogger.Println("something real bad happened to the bot ... going down ...")
		thisLogger.Println(oops)
	}
	log.Printf("Robot: %v initialized", bot.Name)

	/* Strart the server */
	thisLogger.Println("Starting Gizmatron api server...")
	err := server.Start(bot)
	if err != nil{
		thisLogger.Println("something real bad happened to the server ... going down ...")
		thisLogger.Println(err)
	}

}
