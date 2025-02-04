package server

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/arabenjamin/gizmatron/robot"
)

/*Middleware Go wants a comment */
type Middleware func(http.HandlerFunc) http.HandlerFunc

/* log the response */
func logger(thisLogger *log.Logger) Middleware {

	return func(next http.HandlerFunc) http.HandlerFunc {

		return func(resp http.ResponseWriter, req *http.Request) {

			defer func() {
				thisLogger.Printf("[%v] [%v] [%v %v] %v\n", req.RemoteAddr, req.Method, req.Proto, req.URL.Path, req.Header["User-Agent"])
			}()
			next(resp, req)
		}
	}
}

func respond(res http.ResponseWriter, payload map[string]interface{}) {

	json_resp, _ := json.Marshal(payload)
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.WriteHeader(http.StatusOK)
	res.Write(json_resp)

}

func clientHash(req *http.Request) string {
	hash := md5.New()
	clientString := fmt.Sprintf("%v%v%v%v", req.RemoteAddr, req.URL.Path, req.Header["User-Agent"], time.Now().Unix())
	io.WriteString(hash, clientString)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

/*Chain handler*/
func Chain(f http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {

	for _, middleware := range middlewares {
		f = middleware(f)
	}
	return f
}

func Start(bot *robot.Robot, thisLogger *log.Logger) error {

	//Setup Server LED ( Blue LED on pin ...)
	serverled, serverErr := robot.NewLedLine(13, "Sever Led")
	if serverErr != nil {
		log.Printf("Error Turning on Server LED: %v", serverErr)
		bot.Devices["severledError"] = serverErr
	} else {
		bot.Devices["serverLed"] = "Operational"
	}
	bot.Serverled = serverled
	// Turn the server led on now
	// I may want to rethink the way the server light comes on.
	if bot.Devices["severLed"] == "Operational" {
		bot.Serverled.SetValue(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", Chain(ping, logger(thisLogger)))
	/*
		mux.HandleFunc("/bot-status", Chain(get_status, bot, logger(thisLogger)))
		mux.HandleFunc("/bot-start", Chain(start_bot, bot, logger(thisLogger)))
		mux.HandleFunc("/bot-stop", Chain(stop_bot, bot, logger(thisLogger)))
		mux.HandleFunc("/api/v1/detectfaces", Chain(set_facedetect, bot, logger(thisLogger)))
		mux.HandleFunc("/video", Chain(get_video, bot, logger(thisLogger)))
	*/
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		return err
	}
	return nil
}
