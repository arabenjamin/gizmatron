package server

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	_"os"
	"time"
	"github.com/arabenjamin/gizmatron/robot"

)




/*Middleware Go wants a comment */
type Middleware func(http.HandlerFunc) http.HandlerFunc

/* log the response */
func logReq(req *http.Request) {
	fmt.Printf("[%v] [%v] [%v] [%v %v] %v\n", time.Now().Unix(), req.RemoteAddr, req.Method, req.Proto, req.URL.Path, req.Header["User-Agent"])
	/*TODO: return request hashmap */
}

func logger(thisLogger *log.Logger) Middleware {

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				thisLogger.Println(r.URL.Path, time.Now().Unix())
			}()
			next(w, r)
		}
	}
}

func respond(res http.ResponseWriter, payload map[string]interface{}) {

	thisJSON, _ := json.Marshal(payload)
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.WriteHeader(http.StatusOK)
	res.Write(thisJSON)

}

func clientHash(req *http.Request) string {
	hash := md5.New()
	clientString := fmt.Sprintf("%v%v%v%v", req.RemoteAddr, req.URL.Path, req.Header["User-Agent"], time.Now().Unix())
	io.WriteString(hash, clientString)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func ping(resp http.ResponseWriter, req *http.Request) {

		/* TODO: Maybe rethink this*/ 
	
	//robot.BlinkBot()

	thisRequest := map[string]interface{}{
		"time":           time.Now().Unix(),
		"client_address": req.RemoteAddr,
		"resource":       req.URL.Path,
		"user_agent":     req.Header["User-Agent"],
		"client":         clientHash(req),
	}

	thisResponse := map[string]interface{}{
		"status":       "ok",
		"message":      "pong!",
		"this_request": thisRequest,
	}

	logReq(req)
	respond(resp, thisResponse)
	return
}

/*Chain handler*/
func Chain(f http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {

	for _, m := range middlewares {
		f = m(f)
	}
	return f
}

func Start( bot *robot.Robot) error {

	//fmt.Println("Starting WebApp")

	//thisLogger := log.New(os.Stdout, "http: ", log.LstdFlags)
	//thisLogger.Println("Starting server...")

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", ping)

	mux.HandleFunc("/bot-status", func(resp http.ResponseWriter, req *http.Request){

		status := fmt.Sprintf("%v, is running", bot.Name)
		if !bot.State {
			status = fmt.Sprintf("%v, is not running", bot.Name)
		}

		thisRequest := map[string]interface{}{
			"time":           time.Now().Unix(),
			"client_address": req.RemoteAddr,
			"resource":       req.URL.Path,
			"user_agent":     req.Header["User-Agent"],
			"client":         clientHash(req),
		}
	
		thisResponse := map[string]interface{}{
			"status":       status,
			"botname":      bot.Name,
			"this_request": thisRequest,
		}
	
		logReq(req)
		respond(resp, thisResponse)

	})

	mux.HandleFunc("/bot-start", func(resp http.ResponseWriter, req *http.Request){


		status := fmt.Sprintf("%v, is already running", bot.Name)
		if !bot.State {
			status = fmt.Sprintf(" Starting %v", bot.Name)
			go bot.Start()
		}

		thisRequest := map[string]interface{}{
			"time":           time.Now().Unix(),
			"client_address": req.RemoteAddr,
			"resource":       req.URL.Path,
			"user_agent":     req.Header["User-Agent"],
			"client":         clientHash(req),
		}
	
		thisResponse := map[string]interface{}{
			"status":       status,
			"botname":      bot.Name,
			"this_request": thisRequest,
		}
	
		logReq(req)
		respond(resp, thisResponse)

	})

	mux.HandleFunc("/bot-stop", func(resp http.ResponseWriter, req *http.Request){

		status := fmt.Sprintf("%v, is not running", bot.Name)
		if bot.State {
			go bot.Stop()
		} 

		thisRequest := map[string]interface{}{
			"time":           time.Now().Unix(),
			"client_address": req.RemoteAddr,
			"resource":       req.URL.Path,
			"user_agent":     req.Header["User-Agent"],
			"client":         clientHash(req),
		}
	
		thisResponse := map[string]interface{}{
			"status":       status,
			"botname":      bot.Name,
			"this_request": thisRequest,
		}
	
		logReq(req)
		respond(resp, thisResponse)

	})


	//mux.HandleFunc("/ping", Chain(ping, logger(thisLogger)))
	err := http.ListenAndServe(":8080", mux)
	if err != nil{
		return err
	}
	return nil
}
/*
func main() {

	fmt.Println("Starting WebApp")

	thisLogger := log.New(os.Stdout, "http: ", log.LstdFlags)
	thisLogger.Println("Starting server...")

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", Chain(ping, logger(thisLogger)))
	http.ListenAndServe(":8080", mux)
}*/
