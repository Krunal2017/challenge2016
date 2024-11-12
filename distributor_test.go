package main

import (
	"testing"
)

func TestCreateDistributor(t *testing.T) {
	// Sample distributor data
	request := &Request{
		Name:    "test-distributor",
		Include: []string{"city:province:country"},
		Exclude: []string{"city:province:anotherCountry"},
		Inherits: "",
	}

	// Create distributor from request
	distributor := request.createDistributor()

	// Check that distributor is not nil
	if distributor == nil {
		t.Fatalf("Expected distributor to be created, but got nil")
	}

	// Check if the distributor has the correct name
	if distributor.Name != request.Name {
		t.Errorf("Expected distributor name to be '%s', but got '%s'", request.Name, distributor.Name)
	}

	// Check the Include and Exclude lists
	if len(distributor.Include) != 1 || distributor.Include[0] != "city:province:country" {
		t.Errorf("Expected include list to have 1 item: 'city:province:country', but got %v", distributor.Include)
	}

	if len(distributor.Exclude) != 1 || distributor.Exclude[0] != "city:province:anotherCountry" {
		t.Errorf("Expected exclude list to have 1 item: 'city:province:anotherCountry', but got %v", distributor.Exclude)
	}
}

func TestCloneDistributor(t *testing.T) {
	// Sample parent distributor
	parent := &Distributor{
		Name:    "parent-distributor",
		Include: []string{"parentCountry"},
		Exclude: []string{"excludeCountry"},
		Inherits: "",
	}

	// Sample request to clone
	request := &Request{
		Name:    "child-distributor",
		Include: []string{"city:province:parentCountry"},
		Exclude: []string{"city:province:excludeCountry"},
		Inherits: "parent-distributor",
	}

	// Add parent to the global distributorDB
	distributorDB["parent-distributor"] = parent

	// Create distributor from request (it should clone from parent)
	distributor := request.createDistributor()

	// Check that the distributor has inherited the parent's data
	if len(distributor.Include) != 1 || distributor.Include[0] != "city:province:parentCountry" {
		t.Errorf("Expected include list to have 'city:province:parentCountry', but got %v", distributor.Include)
	}

	if len(distributor.Exclude) != 2 || distributor.Exclude[0] != "excludeCountry" || distributor.Exclude[1] != "city:province:excludeCountry" {
		t.Errorf("Expected exclude list to have both 'city:province:excludeCountry', but got %v", distributor.Exclude)
	}

	// Check inheritance
	if distributor.Inherits != "parent-distributor" {
		t.Errorf("Expected distributor to inherit from 'parent-distributor', but got '%s'", distributor.Inherits)
	}
}
