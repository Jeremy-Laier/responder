package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

// atomic variable
var acceptingRequests int32

type RequestEvent struct {
	request http.Request
	writer http.ResponseWriter
}

func processor(requestsChan chan RequestEvent) {
	for req := range requestsChan {
		log.Default().Println("event processed", req.request.Method)
	}
}

func mimicWrapper(requestChan chan RequestEvent) http.Handler {
	fn := func (w http.ResponseWriter, r *http.Request) {
		if acceptingRequests != 0 {
			w.WriteHeader(500)
			log.Default().Fatal("closing SERVER")
			return
		}

		event := RequestEvent{request: *r, writer: w}
		requestChan <- event

		var curTime = time.Now()
		var respJson map[string]interface{}

		err := json.NewDecoder(event.request.Body).Decode(&respJson)

		if err != nil {
			log.Fatalln("could not parse json", err, respJson)
		}

		respJson["receieved"] = curTime.Local().String()
		jsonResp, err := json.Marshal(respJson)

		if err != nil {
			log.Fatalln("could not reencode json", err)
		}

		event.writer.Header().Set("Content-Type", "application/json")
		event.writer.WriteHeader(200)
		event.writer.Write(jsonResp)
	}
	return http.HandlerFunc(fn)
}

func main() {
	requests := make(chan RequestEvent, 100)
	atomic.AddInt32(&acceptingRequests, 0)

	go processor(requests)
	go handleCtrlC()

	mux := http.NewServeMux()
	handler := mimicWrapper(requests)
	mux.Handle("/", handler)

	err := http.ListenAndServe(":3333", mux)
	log.Default().Println(err)
}

// this function handles the sig int and sig term signals
func handleCtrlC() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		sig := <- sigs
		atomic.AddInt32(&acceptingRequests, 1)
		log.Default().Println("signal found closing program, flushing requests", sig)

		os.Exit(1)
	}()
}

