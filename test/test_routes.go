package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TestResult struct {
	Endpoint string
	Method   string
	Status   int
	Success  bool
	Error    string
}

func testEndpoint(method, url string, payload interface{}) TestResult {
	var body []byte
	if payload != nil {
		body, _ = json.Marshal(payload)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return TestResult{url, method, 0, false, err.Error()}
	}

	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return TestResult{url, method, 0, false, err.Error()}
	}
	defer resp.Body.Close()

	success := resp.StatusCode < 500
	return TestResult{url, method, resp.StatusCode, success, ""}
}

func main() {
	baseURL := "http://localhost:8080/api/v1"
	
	tests := []struct {
		name     string
		method   string
		endpoint string
		payload  interface{}
	}{
		{
			"Register User",
			"POST",
			"/auth/register",
			map[string]string{
				"email":       "test@example.com",
				"password":    "password123",
				"firstName":   "John",
				"lastName":    "Doe",
				"phoneNumber": "+1234567890",
			},
		},
		{
			"Login User",
			"POST",
			"/auth/login",
			map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
		},
		{
			"Invalid Login",
			"POST",
			"/auth/login",
			map[string]string{
				"email":    "invalid@example.com",
				"password": "wrongpassword",
			},
		},
		{
			"Register Duplicate User",
			"POST",
			"/auth/register",
			map[string]string{
				"email":       "test@example.com",
				"password":    "password123",
				"firstName":   "Jane",
				"lastName":    "Doe",
				"phoneNumber": "+1234567891",
			},
		},
	}

	fmt.Println("Testing API Routes...")
	fmt.Println("====================")

	for _, test := range tests {
		result := testEndpoint(test.method, baseURL+test.endpoint, test.payload)
		
		status := "✓ PASS"
		if !result.Success {
			status = "✗ FAIL"
		}
		
		fmt.Printf("%s | %s %s | Status: %d | %s\n", 
			status, test.method, test.endpoint, result.Status, test.name)
		
		if result.Error != "" {
			fmt.Printf("    Error: %s\n", result.Error)
		}
	}
}
