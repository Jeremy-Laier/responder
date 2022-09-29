package main

import (
	"fmt"
	"net/http"
	"log"
	"encoding/json"
	"time"
)


func getRoot(w http.ResponseWriter, r *http.Request) {
	var curTime = time.Now()
	var respJson map[string]interface{}

	err := json.NewDecoder(r.Body).Decode(&respJson)

	if err != nil {
		log.Fatalln("could not parse json", err, respJson)
	}


	respJson["receieved"] = curTime.Local().String()
	jsonResp, err := json.Marshal(respJson)

	if err != nil {
		log.Fatalln("could not reencode json", err)
	}

	log.Default().Println(string(jsonResp))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsonResp)
}

func main() {
	http.HandleFunc("/", getRoot)

	err := http.ListenAndServe(":3333", nil)
	fmt.Println(err)
}
