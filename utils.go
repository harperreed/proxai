package main

import (
    "fmt"
    "time"

    "github.com/charmbracelet/lipgloss"
)

var (
    infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
    successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
    errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
    boldStyle    = lipgloss.NewStyle().Bold(true)
    statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("235")).Background(lipgloss.Color("11")).Padding(0, 1)
    lipglossStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func (s *ProxyServer) updateStatusBar() {
    for {
        requestCount, tokensCount := s.getStats()
        status := fmt.Sprintf("🐨: %s | Status: %s | Tokens: %d | Requests: %d", fmt.Sprintf("http://%s:%d", *address, *port), "✅", tokensCount, requestCount)
        statusBar := statusStyle.Render(status)

        fmt.Print("\033[s")    // Save cursor position
        fmt.Print("\033[999B") // Move cursor to the bottom
        fmt.Print("\r")        // Move cursor to the beginning of the line
        fmt.Print(statusBar)
        fmt.Print("\033[u") // Restore cursor position

        time.Sleep(1 * time.Second)
    }
}

func clearStatusBar() {
    fmt.Print("\033[s")    // Save cursor position
    fmt.Print("\033[999B") // Move cursor to the bottom
    fmt.Print("\r")        // Move cursor to the beginning of the line
    fmt.Print("\033[K")    // Clear the line
    fmt.Print("\033[u")    // Restore cursor position
}
