package main

import (
	"fmt"
	"os"
	"log"
	"encoding/json"
)

func readJSON(name string) Distributor{
	file, err := os.Open(name)
	defer file.Close()
	if err != nil{
		log.Fatal("Error while reading file", err)
	}
	// Decode JSON data from the file
	var data Distributor
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	fmt.Printf("%+v\n", data)

	return data
}
