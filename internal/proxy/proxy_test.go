package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/harperreed/proxai/internal/cache"
	"github.com/harperreed/proxai/internal/logger"
)

// MockCache is a mock implementation of the Cache interface
type MockCache struct{}

func (m *MockCache) Set(key string, value []byte) error { return nil }
func (m *MockCache) Get(key string) ([]byte, error)     { return nil, nil }
func (m *MockCache) Delete(key string) error            { return nil }

// MockLogger is a mock implementation of the Logger
type MockLogger struct{}

func (m *MockLogger) LogRequest(method, path string, body map[string]interface{}) {}
func (m *MockLogger) LogResponse(statusCode int, body map[string]interface{})     {}
func (m *MockLogger) LogPrompt(prompt string)                                     {}
func (m *MockLogger) LogCost(model string, tokens int, cost float64)              {}
func (m *MockLogger) Close()                                                      {}

func TestNewProxyServer(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "proxyserver_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	server, err := NewProxyServer(tempDir, tempDir)
	if err != nil {
		t.Fatalf("NewProxyServer failed: %v", err)
	}

	if server.Client == nil {
		t.Error("Client is nil")
	}
	if server.Cache == nil {
		t.Error("Cache is nil")
	}
	if server.Logger == nil {
		t.Error("Logger is nil")
	}

	// Test error handling for invalid log directory
	_, err = NewProxyServer(tempDir, "/invalid/directory")
	if err == nil {
		t.Error("Expected error for invalid log directory, got nil")
	}
}

func TestIncrementRequestCount(t *testing.T) {
	server := &ProxyServer{}
	server.incrementRequestCount()
	if server.RequestCount != 1 {
		t.Errorf("Expected RequestCount to be 1, got %d", server.RequestCount)
	}
	server.incrementRequestCount()
	if server.RequestCount != 2 {
		t.Errorf("Expected RequestCount to be 2, got %d", server.RequestCount)
	}
}

func TestIncrementTokensCount(t *testing.T) {
	server := &ProxyServer{}
	server.incrementTokensCount(10)
	if server.TokensCount != 10 {
		t.Errorf("Expected TokensCount to be 10, got %d", server.TokensCount)
	}
	server.incrementTokensCount(5)
	if server.TokensCount != 15 {
		t.Errorf("Expected TokensCount to be 15, got %d", server.TokensCount)
	}
}

func TestAddCost(t *testing.T) {
	server := &ProxyServer{}
	server.addCost(1.5)
	if server.TotalCost != 1.5 {
		t.Errorf("Expected TotalCost to be 1.5, got %f", server.TotalCost)
	}
	server.addCost(0.5)
	if server.TotalCost != 2.0 {
		t.Errorf("Expected TotalCost to be 2.0, got %f", server.TotalCost)
	}
}

func TestResetCounters(t *testing.T) {
	server := &ProxyServer{RequestCount: 10, TokensCount: 100, TotalCost: 5.0}
	server.ResetCounters()
	if server.RequestCount != 0 || server.TokensCount != 0 || server.TotalCost != 0 {
		t.Errorf("Counters were not reset properly. Got RequestCount: %d, TokensCount: %d, TotalCost: %f",
			server.RequestCount, server.TokensCount, server.TotalCost)
	}
}

func TestGetStats(t *testing.T) {
	server := &ProxyServer{RequestCount: 5, TokensCount: 50, TotalCost: 2.5}
	reqCount, tokenCount, totalCost := server.GetStats()
	if reqCount != 5 || tokenCount != 50 || totalCost != 2.5 {
		t.Errorf("GetStats returned incorrect values. Got reqCount: %d, tokenCount: %d, totalCost: %f",
			reqCount, tokenCount, totalCost)
	}
}

func TestOpenAIProxy(t *testing.T) {
	server := &ProxyServer{
		Client: &http.Client{},
		Cache:  &MockCache{},
		Logger: &MockLogger{},
	}

	// Create a test server to mock OpenAI API
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"usage": {"total_tokens": 10}}`))
	}))
	defer testServer.Close()

	// Test with valid authorization
	reqBody := map[string]interface{}{"model": "gpt-3.5-turbo"}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer test-token")
	rr := httptest.NewRecorder()
	server.OpenAIProxy(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	if server.RequestCount != 1 || server.TokensCount != 10 {
		t.Errorf("Counters were not incremented correctly. RequestCount: %d, TokensCount: %d",
			server.RequestCount, server.TokensCount)
	}

	// Test with invalid authorization
	req = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr = httptest.NewRecorder()
	server.OpenAIProxy(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code for invalid token: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestExtractTokenUsage(t *testing.T) {
	server := &ProxyServer{}
	testCases := []struct {
		name     string
		response map[string]interface{}
		expected int
	}{
		{
			name: "Valid usage data",
			response: map[string]interface{}{
				"usage": map[string]interface{}{
					"total_tokens": float64(15),
				},
			},
			expected: 15,
		},
		{
			name:     "Missing usage data",
			response: map[string]interface{}{},
			expected: 0,
		},
		{
			name: "Invalid usage data type",
			response: map[string]interface{}{
				"usage": map[string]interface{}{
					"total_tokens": "invalid",
				},
			},
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := server.extractTokenUsage(tc.response)
			if tokens != tc.expected {
				t.Errorf("Expected %d tokens, got %d", tc.expected, tokens)
			}
		})
	}
}

func TestLogRequestDetails(t *testing.T) {
	server := &ProxyServer{
		Logger: &MockLogger{},
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	requestBody := map[string]interface{}{"model": "gpt-3.5-turbo"}
	responseBody := []byte(`{"usage": {"total_tokens": 10}}`)

	// This test mainly checks if the method runs without panicking
	server.logRequestDetails(req, "gpt-3.5-turbo", 10, http.StatusOK, requestBody, responseBody)
}

// TestHelpers

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func setupTestFixtures() (map[string]interface{}, []byte) {
	requestBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": "Hello, how are you?"},
		},
	}
	responseBody := []byte(`{
		"id": "chatcmpl-123",
		"object": "chat.completion",
		"created": 1677652288,
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "Hello! As an AI language model, I don't have feelings, but I'm functioning well and ready to assist you. How can I help you today?"
			},
			"finish_reason": "stop"
		}],
		"usage": {
			"prompt_tokens": 9,
			"completion_tokens": 29,
			"total_tokens": 38
		}
	}`)
	return requestBody, responseBody
}
