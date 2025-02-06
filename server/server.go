package server

import (
	"context"
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
func logger(serverlog *log.Logger) Middleware {

	return func(next http.HandlerFunc) http.HandlerFunc {

		return func(resp http.ResponseWriter, req *http.Request) {

			defer func() {
				serverlog.Printf("[%v] [%v] [%v %v] %v\n", req.RemoteAddr, req.Method, req.Proto, req.URL.Path, req.Header["User-Agent"])
			}()
			next(resp, req)
		}
	}
}

func robotware(bot *robot.Robot) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {

		return func(resp http.ResponseWriter, req *http.Request) {

			defer func() {
				//log.Println("Got status")
			}()
			log.Println("Talking to Gizmatron")
			bot := context.WithValue(req.Context(), "bot", bot)
			next(resp, req.WithContext(bot))

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

func Start(bot *robot.Robot, serverlog *log.Logger) error {

	/* Register our routes middlewares and handlers */
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", Chain(ping, logger(serverlog)))
	mux.HandleFunc("/api/v1/bot-status", Chain(get_status, logger(serverlog), robotware(bot)))
	mux.HandleFunc("/api/v1/bot-start", Chain(start_bot, logger(serverlog), robotware(bot)))
	mux.HandleFunc("/api/v1/bot-stop", Chain(stop_bot, logger(serverlog), robotware(bot)))
	mux.HandleFunc("/api/v1/detectfaces", Chain(set_facedetect, logger(serverlog), robotware(bot)))
	mux.HandleFunc("/api/v1/video", Chain(get_video, logger(serverlog), robotware(bot)))
	mux.HandleFunc("/api/v1/takepicture", Chain(take_picture, logger(serverlog), robotware(bot)))
	//mux.Handle("/stream", bot.Camera.Stream)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		return err
	}

	return nil
}
