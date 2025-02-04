package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/arabenjamin/gizmatron/robot"
)

func error_handler(resp http.ResponseWriter, req *http.Request) {}

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

	respond(resp, thisResponse)
	return
}

func get_status(bot *robot.Robot, resp http.ResponseWriter, req *http.Request) {
	status := fmt.Sprintf("%v, is running", bot.Name)
	if !bot.IsRunning {
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
		"status": status,
		"camera_state": map[string]interface{}{
			"operational": bot.Camera.IsOperational,
			"running":     bot.Camera.IsRunning,
			"empty":       bot.Camera.ImgMat.Empty(),
			"Detected":    bot.Camera.DetectFaces,
		},
		"device_status": bot.Devices,
		"botname":       bot.Name,
		"this_request":  thisRequest,
	}

	//logReq(req)
	respond(resp, thisResponse)
}

func get_video(bot *robot.Robot, resp http.ResponseWriter, req *http.Request) {

	// TODO: The below is really bad, and needs to be refactored

	resp.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")

	// TODO: Build camera running light on pysical Robot
	/* Turn on video light*/

	status := fmt.Sprintf("%v, is running", bot.Name)
	// TODO: refactor .IsRunning to .IsOperational
	if !bot.IsRunning {
		status = fmt.Sprintf("%v, is not running", bot.Name)

		thisRequest := map[string]interface{}{
			"time":           time.Now().Unix(),
			"client_address": req.RemoteAddr,
			"resource":       req.URL.Path,
			"user_agent":     req.Header["User-Agent"],
			"client":         clientHash(req),
		}

		thisResponse := map[string]interface{}{
			"status":        status,
			"device_status": bot.Devices,
			"botname":       bot.Name,
			"this_request":  thisRequest,
		}

		//logReq(req)
		respond(resp, thisResponse)
		return
	}

	if !bot.Camera.IsRunning {
		log.Printf("The camera is not running")
		status = "The camera is not running"
		thisRequest := map[string]interface{}{
			"time":           time.Now().Unix(),
			"client_address": req.RemoteAddr,
			"resource":       req.URL.Path,
			"user_agent":     req.Header["User-Agent"],
			"client":         clientHash(req),
		}

		thisResponse := map[string]interface{}{
			"status": status,
			"camera_state": map[string]interface{}{
				"operational": bot.Camera.IsOperational,
				"running":     bot.Camera.IsRunning,
				"empty":       bot.Camera.ImgMat.Empty(),
				"Detected":    bot.Camera.DetectFaces,
			},
			"device_status": bot.Devices,
			"botname":       bot.Name,
			"this_request":  thisRequest,
		}

		//logReq(req)
		respond(resp, thisResponse)
		return
	}
}

func set_facedetect(bot *robot.Robot, resp http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(resp, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		Enable bool `json:"enable"`
	}

	if err := json.NewDecoder(req.Body).Decode(&requestData); err != nil {
		http.Error(resp, "Invalid request body", http.StatusBadRequest)
		return
	}

	bot.Camera.DetectFaces = requestData.Enable

	status := "Face detection disabled"
	if requestData.Enable {
		status = "Face detection enabled"
	}

	thisRequest := map[string]interface{}{
		"time":           time.Now().Unix(),
		"client_address": req.RemoteAddr,
		"resource":       req.URL.Path,
		"user_agent":     req.Header["User-Agent"],
		"client":         clientHash(req),
	}

	thisResponse := map[string]interface{}{
		"status":        status,
		"device_status": bot.Devices,
		"camera_state": map[string]interface{}{
			"operational": bot.Camera.IsOperational,
			"running":     bot.Camera.IsRunning,
			"empty":       bot.Camera.ImgMat.Empty(),
			"Detected":    bot.Camera.DetectFaces,
		},
		"botname":      bot.Name,
		"this_request": thisRequest,
	}

	//logReq(req)
	respond(resp, thisResponse)
}

func start_bot(bot *robot.Robot, resp http.ResponseWriter, req *http.Request) {

	status := fmt.Sprintf("%v, is already running", bot.Name)
	if !bot.IsRunning {
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
		"status":        status,
		"device_status": bot.Devices,
		"botname":       bot.Name,
		"this_request":  thisRequest,
	}

	//logReq(req)
	respond(resp, thisResponse)
}

func stop_bot(bot *robot.Robot, resp http.ResponseWriter, req *http.Request) {

	status := fmt.Sprintf("%v, is not running", bot.Name)
	if bot.IsRunning {
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
		"status":        status,
		"device_status": bot.Devices,
		"botname":       bot.Name,
		"this_request":  thisRequest,
	}

	//logReq(req)
	respond(resp, thisResponse)

}
