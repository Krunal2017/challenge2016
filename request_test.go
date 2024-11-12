package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"bytes"
	"encoding/json"
)

func TestGetLocation(t *testing.T) {
	tests := []struct {
		input    string
		expected struct {
			city, province, country string
		}
	}{
		{"city:province:country", struct{ city, province, country string }{"city", "province", "country"}},
		{"province:country", struct{ city, province, country string }{"", "province", "country"}},
		{"country", struct{ city, province, country string }{"", "", "country"}},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			city, province, country := getLocation(test.input)

			if city != test.expected.city || province != test.expected.province || country != test.expected.country {
				t.Errorf("For input '%s', expected (city=%s, province=%s, country=%s), but got (city=%s, province=%s, country=%s)",
					test.input, test.expected.city, test.expected.province, test.expected.country, city, province, country)
			}
		})
	}
}

func TestCheckDistributor(t *testing.T) {
	// Sample distributor
	distributorDB["distributor1"] = &Distributor{
		Name:    "distributor1",
		Include: []string{"city:province:country"},
		Exclude: []string{"city:province:anotherCountry"},
	}

	// Test cases
	tests := []struct {
		distributor string
		location    string
		expected    string
	}{
		{"distributor1", "city:province:country", MsgAccessGranted},
		{"distributor1", "city:province:anotherCountry", MsgAccessDenied},
		{"distributor1", "unknownCity:unknownProvince:unknownCountry", MsgAccessDenied},
		{"nonExistent", "city:province:country", ErrDistributorNotExist},
	}

	for _, test := range tests {
		t.Run(test.distributor+" "+test.location, func(t *testing.T) {
			result,_ := checkDistributor(test.distributor, test.location)
			if result != test.expected {
				t.Errorf("Expected '%s', but got '%s' for distributor '%s' and location '%s'",
					test.expected, result, test.distributor, test.location)
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	// Clear distributorDB before each test to avoid interference between tests
	distributorDB = make(map[string]*Distributor)

	// Mock data for distributors
	distributorDB["test-distributor"] = &Distributor{
		Name:    "test-distributor",
		Include: []string{"city:province:country"},
		Exclude: []string{"city:province:excludeCountry"},
	}

	tests := []struct {
		name           string
		distributor    string
		location       string
		expectedBody   string
		expectedStatus int
	}{
		{
			name:           "Valid request",
			distributor:    "test-distributor",
			location:       "city:province:country",
			expectedBody:   MsgAccessGranted,  // "YES"
			expectedStatus: http.StatusOK,     // 200 OK
		},
		{
			name:           "Location not included",
			distributor:    "test-distributor",
			location:       "city:province:excludeCountry",
			expectedBody:   MsgAccessDenied,   // "NO"
			expectedStatus: http.StatusForbidden,     // 200 OK (not 403)
		},
		{
			name:           "Missing distributor",
			distributor:    "",
			location:       "city:province:country",
			expectedBody:   ErrMissingParams,  // "Missing 'distributor' or 'location' parameter"
			expectedStatus: http.StatusBadRequest, // 400 Bad Request
		},
		{
			name:           "Missing location",
			distributor:    "test-distributor",
			location:       "",
			expectedBody:   ErrMissingParams,  // "Missing 'distributor' or 'location' parameter"
			expectedStatus: http.StatusBadRequest, // 400 Bad Request
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the getHandler with the mock data
			body, status := getHandler(tt.distributor, tt.location)

			// Check if the status code matches the expected value
			if status != tt.expectedStatus {
				t.Errorf("Expected status %v but got %v", tt.expectedStatus, status)
			}

			// Check if the response body matches the expected value
			if body != tt.expectedBody {
				t.Errorf("Expected body '%v' but got '%v'", tt.expectedBody, body)
			}
		})
	}
}

func TestPostHandler(t *testing.T) {
	// Clear distributorDB before each test to avoid interference between tests
	distributorDB = make(map[string]*Distributor)

	tests := []struct {
		name           string
		payload        *Request
		expectedStatus int
		expectedBody   interface{} // to allow flexibility for string or struct comparison
	}{
		{
			name: "Valid request to create distributor",
			payload: &Request{
				Name:    "new-distributor",
				Include: []string{"parentCountry"},
				Exclude: []string{"city:province:excludeCountry"},
				Inherits: "",
			},
			expectedBody: &Distributor{
				Name:     "new-distributor",
				Include:  []string{"parentCountry"},
				Exclude:  []string{"city:province:excludeCountry"},
				Inherits: "",
			},
			expectedStatus: http.StatusCreated, // 201 Created
		},
		{
			name: "Valid request to create distributor with parent",
			payload: &Request{
				Name:    "child-distributor",
				Include: []string{"city:province:parentCountry","city:province:excludeCountry"},
				Exclude: []string{},
				Inherits: "new-distributor",
			},
			expectedBody: &Distributor{
				Name:     "child-distributor",
				Include:  []string{"city:province:parentCountry"},
				Exclude:  []string{"city:province:excludeCountry"},
				Inherits: "new-distributor",
			},
			expectedStatus: http.StatusCreated, // 201 Created
		},
		{
			name: "Distributor already exists",
			payload: &Request{
				Name:    "new-distributor", // Name already exists
				Include: []string{"city:province:parentCountry"},
				Exclude: []string{"city:province:excludeCountry"},
				Inherits: "",
			},
			expectedBody: &Distributor{
				Name:     "new-distributor",
				Include:  []string{"parentCountry"},
				Exclude:  []string{"city:province:excludeCountry"},
				Inherits: "",
			},
			expectedStatus: http.StatusConflict,   // 409 Conflict
		},
		{
			name:           "Invalid JSON payload",
			payload:        nil,                    // Invalid payload (nil)
			expectedBody:   ErrInvalidJSONPayload,  // "Invalid JSON payload"
			expectedStatus: http.StatusBadRequest,  // 400 Bad Request
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare the JSON payload
			var jsonPayload []byte
			if tt.payload != nil {
				var err error
				jsonPayload, err = json.Marshal(tt.payload)
				if err != nil {
					t.Fatalf("Failed to marshal payload: %v", err)
				}
			}

			// Create a request using httptest.NewRequest
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(jsonPayload))

			// Call the postHandler and capture the return values
			body, status := postHandler(req)

			// Check if the status code matches the expected value
			if status != tt.expectedStatus {
				t.Errorf("Expected status %v but got %v", tt.expectedStatus, status)
			}

			// Check if the response body matches the expected value
			switch expectedBody := tt.expectedBody.(type) {
			case string:
				// Compare as string for error messages
				if body != expectedBody {
					t.Errorf("Expected body '%v' but got '%v'", expectedBody, body)
				}
			case *Distributor:
				// Decode JSON response and compare as struct for successful creation
				if status == http.StatusCreated {
					var actualDistributor Distributor
					err := json.Unmarshal([]byte(body), &actualDistributor)
					if err != nil {
						t.Fatalf("Failed to unmarshal JSON response: %v", err)
					}

					// Compare distributor attributes
					if actualDistributor.Name != expectedBody.Name ||
						!equalSlices(actualDistributor.Include, expectedBody.Include) ||
						!equalSlices(actualDistributor.Exclude, expectedBody.Exclude) ||
						actualDistributor.Inherits != expectedBody.Inherits {
						t.Errorf("Expected distributor %v but got %v", expectedBody, actualDistributor)
					}
				}
			}
		})
	}
}

// Helper function to compare slices
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

