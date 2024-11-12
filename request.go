package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
)

var distributorMu sync.RWMutex

// split the given location code. Expected format: city:province:country
func getLocation(location string) (string, string, string) {
	parts := strings.Split(location, ":")
	var city, province, country string

	switch len(parts) {
	case 3:
		city = parts[0]
		province = parts[1]
		country = parts[2]
	case 2:
		province = parts[0]
		country = parts[1]
	case 1:
		country = parts[0]
	}
	return city, province, country
}

// Checks if a distributor has access to the given location code
func checkDistributor(distributor string, location string) (string, int) {
	log.Printf("Checking distributor '%s' for location '%s'", distributor, location)
	gcity, gprovince, gcountry := getLocation(location)

	distributorMu.RLock()
	if _, ok := distributorDB[distributor]; !ok {
		distributorMu.RUnlock()
		log.Printf("Distributor '%s' not found", distributor)
		return ErrDistributorNotExist, http.StatusNotFound
	}
	distributorMu.RUnlock()

	d := distributorDB[distributor]
	includeList := d.Include
	excludeList := d.Exclude

	for _, code := range excludeList {
		city, province, country := getLocation(code)
		if country == gcountry {
			if province == "" || (province == gprovince && (city == "" || city == gcity)) {
				log.Printf("Distributor '%s' not authorized for location '%s'", distributor, location)
				return MsgAccessDenied, http.StatusForbidden
			}
		}
	}

	for _, code := range includeList {
		city, province, country := getLocation(code)
		if country == gcountry {
			if province == "" || (province == gprovince && (city == "" || city == gcity)) {
				log.Printf("Distributor '%s' authorized for location '%s'", distributor, location)
				return MsgAccessGranted, http.StatusOK
			}
		}
	}

	log.Printf("Distributor '%s' not authorized for location '%s'", distributor, location)
	return MsgAccessDenied, http.StatusForbidden
}

// Handles GET requests
func getHandler(distributor string, location string) (string, int) {
	log.Printf("Handling GET request for distributor '%s' at location '%s'", distributor, location)
	if distributor == "" || location == "" {
		log.Println("Missing 'distributor' or 'location' parameter")
		return ErrMissingParams, http.StatusBadRequest
	}

	return checkDistributor(distributor, location)
}

// Processes POST request, to create a distributor
func postProcessing(distributor *Request) (string, int, *Distributor) {
	log.Printf("Processing POST request for distributor '%s'", distributor.Name)
	distributorMu.RLock()
	if _, ok := distributorDB[distributor.Name]; ok {
		distributorMu.RUnlock()
		log.Printf("Distributor '%s' already exists", distributor.Name)
		// TODO return existing distributor
		return ErrDistributorExists, http.StatusConflict, distributorDB[distributor.Name]
	}
	distributorMu.RUnlock()

	newDistributor := distributor.createDistributor()

	distributorMu.Lock()
	distributorDB[distributor.Name] = newDistributor
	distributorMu.Unlock()
	log.Printf("Distributor '%s' created successfully", distributor.Name)

	return MsgDistributorCreated, http.StatusCreated, newDistributor
}

// Handles POST requests
func postHandler(r *http.Request) (string, int) {
	var distributor *Request
	err := json.NewDecoder(r.Body).Decode(&distributor)
	if err != nil {
		log.Printf("Invalid JSON payload: %v", err)
		return ErrInvalidJSONPayload, http.StatusBadRequest
	}

	response, status, dist := postProcessing(distributor)
	if response != MsgDistributorCreated && response != ErrDistributorExists{
		return response, status
	}

	// Encode the distributor object as JSON to return in response
	responseBytes, err := json.Marshal(dist)
	if err != nil {
		return ErrInvalidJSONResponse, http.StatusInternalServerError
	}

	return string(responseBytes), status
}

// Handles incoming requests and returns responses
func requestHandler(w http.ResponseWriter, r *http.Request) {
	var response string
	var status int

	switch r.Method {
	case http.MethodGet:
		distributor := r.URL.Query().Get("distributor")
		location := r.URL.Query().Get("location")
		log.Printf("Received GET request with distributor='%s', location='%s'", distributor, location)
		response, status = getHandler(distributor, location)

	case http.MethodPost:
		response, status = postHandler(r)

	default:
		log.Printf("Method '%s' not allowed", r.Method)
		response = ErrMethodNotAllowed
		status = http.StatusMethodNotAllowed
	}

	log.Printf("Response status: %d, response body: %s", status, response)
	w.WriteHeader(status)
	w.Write([]byte(response))
}
