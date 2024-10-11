package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/harperreed/proxai/internal/api"
	"github.com/harperreed/proxai/internal/proxy"
	"github.com/harperreed/proxai/internal/ui"
)

func TestMain(m *testing.M) {
	// Setup test environment
	*cacheDir = "./test_cache"
	*logDir = "./test_logs"

	// Run tests
	code := m.Run()

	// Cleanup
	os.RemoveAll(*cacheDir)
	os.RemoveAll(*logDir)

	os.Exit(code)
}

func TestServerInitialization(t *testing.T) {
	server, err := proxy.NewProxyServer(*cacheDir, *logDir)
	if err != nil {
		t.Fatalf("Failed to create proxy server: %v", err)
	}

	if server == nil {
		t.Fatal("Server is nil")
	}

	if server.Cache == nil {
		t.Error("Server cache is nil")
	}

	if server.Logger == nil {
		t.Error("Server logger is nil")
	}
}

func TestRouteHandling(t *testing.T) {
	server, _ := proxy.NewProxyServer(*cacheDir, *logDir)
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.OpenAIProxy)
	mux.HandleFunc("/help", api.HelpHandler)

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{"OpenAI Proxy", "/", http.StatusBadRequest}, // Expects 400 due to missing auth header
		{"Help Handler", "/help", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(testServer.URL + tt.path)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestGracefulShutdown(t *testing.T) {
	server, _ := proxy.NewProxyServer(*cacheDir, *logDir)
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.OpenAIProxy)

	srv := &http.Server{
		Addr:    "localhost:0",
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			t.Errorf("ListenAndServe(): %v", err)
		}
	}()

	// Allow some time for the server to start
	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Server Shutdown: %v", err)
	}
}

func TestUIInitialization(t *testing.T) {
	server, _ := proxy.NewProxyServer(*cacheDir, *logDir)

	// Create a channel to signal when the UI is "started"
	uiStarted := make(chan bool)

	// Mock the UI start function
	originalStartUI := ui.StartUI
	ui.StartUI = func(s *proxy.ProxyServer) error {
		uiStarted <- true
		return nil
	}
	defer func() { ui.StartUI = originalStartUI }()

	go func() {
		if err := ui.StartUI(server); err != nil {
			t.Errorf("Error running UI: %v", err)
		}
	}()

	select {
	case <-uiStarted:
		// UI started successfully
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for UI to start")
	}
}
