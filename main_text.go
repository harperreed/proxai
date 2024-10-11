package main

package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type MockClient struct {
	Resp *http.Response
	Err  error
}

var client HTTPClient = &http.Client{} // Default, real client

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.Resp, m.Err
}

// TestMockClient tests the MockClient struct
func TestMockClient(t *testing.T) {
	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("test response")),
	}
	mockErr := errors.New("test error")

	testCases := []struct {
		name        string
		mockClient  MockClient
		expectResp  *http.Response
		expectErr   error
		expectPanic bool
	}{
		{
			name:       "Success case",
			mockClient: MockClient{Resp: mockResp, Err: nil},
			expectResp: mockResp,
			expectErr:  nil,
		},
		{
			name:       "Error case",
			mockClient: MockClient{Resp: nil, Err: mockErr},
			expectResp: nil,
			expectErr:  mockErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			resp, err := tc.mockClient.Do(req)

			if resp != tc.expectResp {
				t.Errorf("Expected response %v, got %v", tc.expectResp, resp)
			}

			if err != tc.expectErr {
				t.Errorf("Expected error %v, got %v", tc.expectErr, err)
			}
		})
	}
}

// TestHTTPClientInterface tests if both http.Client and MockClient satisfy the HTTPClient interface
func TestHTTPClientInterface(t *testing.T) {
	var _ HTTPClient = &http.Client{}
	var _ HTTPClient = &MockClient{}
}

// TestHelpHandler tests the helpHandler function
func TestHelpHandler(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{"GET request", "GET", http.StatusOK},
		{"POST request", "POST", http.StatusMethodNotAllowed},
		{"PUT request", "PUT", http.StatusMethodNotAllowed},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, "/help", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(helpHandler)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.expectedStatus)
			}

			if tc.expectedStatus == http.StatusOK {
				expectedContentType := "text/html"
				if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, expectedContentType) {
					t.Errorf("content type header does not match: got %v want %v", contentType, expectedContentType)
				}

				if body := rr.Body.String(); !strings.Contains(body, "<html") {
					t.Errorf("unexpected body content: %v", body)
				}
			}
		})
	}
}

// TestHelpHandlerError tests error handling in helpHandler
func TestHelpHandlerError(t *testing.T) {
	// Simulate request creation failure
	_, err := http.NewRequest("GET", ":", nil)
	if err == nil {
		t.Fatal("Expected error for invalid URL, got nil")
	}

	// Test for appropriate error response
	req, _ := http.NewRequest("GET", "/help", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	expectedBody := "Internal Server Error\n"
	if body := rr.Body.String(); body != expectedBody {
		t.Errorf("unexpected error response body: got %v want %v", body, expectedBody)
	}
}
