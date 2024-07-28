package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
)

var (
    port       = flag.Int("port", 8080, "Port to listen on")
    address    = flag.String("address", "localhost", "Address to listen on")
    cacheDir   = flag.String("cache-dir", "./cache", "Directory for caching responses")
    logDir     = flag.String("log-dir", "./logs", "Directory for log files")
    quit       = make(chan os.Signal, 1)
)

func main() {
    flag.Parse()

    server, err := NewProxyServer(*cacheDir, *logDir)
    if err != nil {
        log.Fatalf("Failed to create proxy server: %v", err)
    }

    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-quit
        clearStatusBar()
        server.logger.Close()
        os.Exit(0)
    }()

    http.HandleFunc("/", server.openAIProxy)
    http.HandleFunc("/help", server.helpHandler)

    log.Println(infoStyle.Render("OpenAI Proxy Server is running on") + boldStyle.Render(fmt.Sprintf(" http://%s:%d", *address, *port)))
    log.Println(successStyle.Render("For integration help, visit ") + boldStyle.Render(fmt.Sprintf(" http://%s:%d/help", *address, *port)))

    go server.updateStatusBar()
    go handleKeyboardInput(server)

    if err := http.ListenAndServe(fmt.Sprintf("%s:%d", *address, *port), nil); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
