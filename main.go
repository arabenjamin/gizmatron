package main


import (

	"os"
	"log"
	_"fmt"
	"github.com/arabenjamin/gizmatron/server"
)




func main() {
	/* Strart the server */
	thisLogger := log.New(os.Stdout, "http: ", log.LstdFlags)
	thisLogger.Println("Starting Gizmatron api server...")
	err := server.Start()
	if err != nil{
		thisLogger.Println("something real bad happened ... going down ...")
		thisLogger.Println(err)
	}
}
