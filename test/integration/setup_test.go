// Copyright 2026 go-ordersvc Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	defaultBaseURL = "http://localhost:8080"
	maxRetries     = 30
	retryInterval  = 1 * time.Second
)

var baseURL string

func TestMain(m *testing.M) {
	// Get base URL from environment or use default
	baseURL = os.Getenv("ORDERSVC_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	// Wait for service to be ready
	if err := waitForService(); err != nil {
		fmt.Printf("Service not ready: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()
	os.Exit(code)
}

func waitForService() error {
	client := &http.Client{Timeout: 5 * time.Second}

	for i := 0; i < maxRetries; i++ {
		resp, err := client.Get(baseURL + "/healthz")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			fmt.Printf("Service ready after %d attempts\n", i+1)
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(retryInterval)
	}

	return fmt.Errorf("service not ready after %d retries", maxRetries)
}

// HTTP helper functions

func doRequest(t *testing.T, method, path string, body interface{}) (*http.Response, []byte) {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	return resp, respBody
}

func post(t *testing.T, path string, body interface{}) (*http.Response, []byte) {
	return doRequest(t, http.MethodPost, path, body)
}

func get(t *testing.T, path string) (*http.Response, []byte) {
	return doRequest(t, http.MethodGet, path, nil)
}

func patch(t *testing.T, path string, body interface{}) (*http.Response, []byte) {
	return doRequest(t, http.MethodPatch, path, body)
}

func delete(t *testing.T, path string) (*http.Response, []byte) {
	return doRequest(t, http.MethodDelete, path, nil)
}
