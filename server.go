package main

import (
    "fmt"
    "net/http"
    "log"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/peterbourgon/diskv"
)

type config struct {
    port       int
    address    string
    cacheDir   string
    logFile    string
    promptFile string
    costFile   string
}

type server struct {
    cfg          config
    logger       *log.Logger
    promptLogger *log.Logger
    costLogger   *log.Logger
    cache        *diskv.Diskv
    program      *tea.Program
    model        *model
    service      Service  // add the service to the server struct
}

func (s *server) start() {
    http.HandleFunc("/", s.openAIProxy)
    http.HandleFunc("/help", s.helpHandler)

    addr := fmt.Sprintf("%s:%d", s.cfg.address, s.cfg.port)
    s.logger.Printf("OpenAI Proxy Server is running on http://%s", addr)
    s.logger.Printf("For integration help, visit http://%s/help", addr)

    if err := http.ListenAndServe(addr, nil); err != nil {
        s.logger.Fatalf("Failed to start server: %v", err)
    }
}

func (s *server) statusBar() string {
    status := fmt.Sprintf("🐨: %s | Status: %s | Tokens: %d | Requests: %d | Total Cost: $%.5f",
        fmt.Sprintf("http://%s:%d", s.cfg.address, s.cfg.port), "✅", tokensCount, requestCount, totalCost)
    return statusStyle.Render(status)
}
