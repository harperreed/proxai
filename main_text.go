package main

import (
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

var client HTTPClient = &http.Client{}  // Default, real client


func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
    return m.Resp, m.Err
}

func TestHelpHandler(t *testing.T) {
    req, err := http.NewRequest("GET", "/help", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(helpHandler)

    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    expectedContentType := "text/html"
    if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, expectedContentType) {
        t.Errorf("content type header does not match: got %v want %v", contentType, expectedContentType)
    }
}
