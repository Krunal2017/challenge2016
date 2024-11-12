package main

import (
	"fmt"
	"log"
	"net/http"
)

func startServer(){
	http.HandleFunc("/distributor", requestHandler)

	fmt.Println("Starting server on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// Main driver logic
func main() {

	startServer()

}