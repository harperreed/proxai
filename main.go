package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    "sync"

    "github.com/charmbracelet/lipgloss"
)

var (
    port     = flag.Int("port", 8080, "Port to listen on")
    address  = flag.String("address", "localhost", "Address to listen on")
    cacheDir = flag.String("cache-dir", "./cache", "Directory for caching responses")
    logDir   = flag.String("log-dir", "./logs", "Directory for log files")
)

func main() {
    flag.Parse()

    server, err := NewProxyServer(*cacheDir, *logDir)
    if err != nil {
        log.Fatalf("Failed to create proxy server: %v", err)
    }

    // Create a new serve mux and server
    mux := http.NewServeMux()
    mux.HandleFunc("/", server.openAIProxy)
    mux.HandleFunc("/help", server.helpHandler)

    srv := &http.Server{
        Addr:    fmt.Sprintf("%s:%d", *address, *port),
        Handler: mux,
    }

    // Create a context that can be cancelled
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Create a WaitGroup to wait for all goroutines to finish
    var wg sync.WaitGroup

    // Set up signal handling
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    // Start the server in a goroutine
    wg.Add(1)
    go func() {
        defer wg.Done()
        log.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("OpenAI Proxy Server is running on") +
            lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf(" http://%s:%d", *address, *port)))

        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Printf("ListenAndServe(): %v", err)
            cancel() // Cancel the context if the server fails to start
        }
    }()

    // Start the UI
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := startUI(ctx, server); err != nil {
            log.Printf("Error running UI: %v", err)
            cancel() // Cancel the context if the UI fails
        }
    }()

    // Wait for interrupt signal or context cancellation
    select {
    case <-stop:
        log.Println("Received interrupt signal")
    case <-ctx.Done():
        log.Println("Context cancelled")
    }

    // Shutdown the server
    log.Println("Shutting down server...")
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer shutdownCancel()

    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Printf("Server forced to shutdown: %v", err)
    }

    // Wait for all goroutines to finish
    wg.Wait()

    server.logger.Close()
    log.Println("Server exiting")
}
