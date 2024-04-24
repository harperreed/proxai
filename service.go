package main

import (
    "fmt"
    // "strings"

    // "github.com/charmbracelet/bubbles/viewport"
    // tea "github.com/charmbracelet/bubbletea"
    // "github.com/charmbracelet/lipgloss"
)

// Service defines the operations available to manage the application state
type Service interface {
    UpdateCounts(tokensAdded, requestsAdded int)
    ResetCounts()
    GetStatus() string
}

// appService implements the Service interface
type appService struct {
    tokensCount  int
    requestCount int
    totalCost    float64
    cfg          config
}

// NewService creates a new instance of a service with the initial configuration
func NewService(cfg config) Service {
    return &appService{
        cfg: cfg,
    }
}

// UpdateCounts updates the counters for tokens and requests
func (s *appService) UpdateCounts(tokensAdded, requestsAdded int) {
    s.tokensCount += tokensAdded
    s.requestCount += requestsAdded
    s.totalCost += float64(tokensAdded) * 0.002 / 1000  // assuming a fixed cost for token usage
}

// ResetCounts resets all counters to zero
func (s *appService) ResetCounts() {
    s.tokensCount = 0
    s.requestCount = 0
    s.totalCost = 0.0
}

// GetStatus returns the current status as a formatted string
func (s *appService) GetStatus() string {
    return fmt.Sprintf("🐨: http://%s:%d | Status: %s | Tokens: %d | Requests: %d | Total Cost: $%.5f",
        s.cfg.address, s.cfg.port, "✅", s.tokensCount, s.requestCount, s.totalCost)
}
