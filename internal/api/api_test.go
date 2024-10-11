package api

import (
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHelpHandler(t *testing.T) {
	tests := []struct {
		name           string
		readFile       FileReader
		parseTemplate  TemplateParser
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Successful rendering",
			readFile: func(filename string) ([]byte, error) {
				return []byte("# Test Markdown"), nil
			},
			parseTemplate: func(name, text string) (*template.Template, error) {
				return template.New(name).Parse(text)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "<h1>Test Markdown</h1>",
		},
		{
			name: "README.md not found",
			readFile: func(filename string) ([]byte, error) {
				return nil, errors.New("file not found")
			},
			parseTemplate: func(name, text string) (*template.Template, error) {
				return template.New(name).Parse(text)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
		},
		{
			name: "Template parsing fails",
			readFile: func(filename string) ([]byte, error) {
				return []byte("# Test Markdown"), nil
			},
			parseTemplate: func(name, text string) (*template.Template, error) {
				return nil, errors.New("template parsing error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/help", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				HelpHandler(w, r, tt.readFile, tt.parseTemplate)
			})

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				if contentType := rr.Header().Get("Content-Type"); contentType != "text/html" {
					t.Errorf("content type header does not match: got %v want %v", contentType, "text/html")
				}
			}

			if !strings.Contains(rr.Body.String(), tt.expectedBody) {
				t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestDefaultHelpHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/help", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(DefaultHelpHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html" {
		t.Errorf("content type header does not match: got %v want %v", contentType, "text/html")
	}

	if !strings.Contains(rr.Body.String(), "<html") {
		t.Errorf("handler returned unexpected body: HTML content not found")
	}
}
