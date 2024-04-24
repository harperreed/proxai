package main

import (
    "flag"
    // "fmt"
    "log"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/peterbourgon/diskv"
)

var (
    requestCount = 0
    tokensCount  = 0
    totalCost    = 0.0
)

func main() {
    cfg := parseFlags()
    service := NewService(cfg)  // Create a new service instance
    logger := initLogger(cfg.logFile)
    promptLogger := initLogger(cfg.promptFile)
    costLogger := initLogger(cfg.costFile)

    cache := diskv.New(diskv.Options{
        BasePath:     cfg.cacheDir,
        CacheSizeMax: 100 * 1024 * 1024, // 100MB cache size
    })

    m := initialModel()
    m.service = service  // Pass the service to the model

    p := tea.NewProgram(m)



    server := &server{
        cfg:          cfg,
        logger:       logger,
        promptLogger: promptLogger,
        costLogger:   costLogger,
        cache:        cache,
        program:      p,
        model:        &m,
        service:      service,  // Pass the service to the server
    }

    go func() {
        server.start()
    }()

    if err := p.Start(); err != nil {
        logger.Fatal(err)
    }
}

func parseFlags() config {
    cfg := config{}
    flag.IntVar(&cfg.port, "port", 8080, "Port to listen on (default: 8080)")
    flag.StringVar(&cfg.address, "address", "localhost", "Address to listen on (default: localhost)")
    flag.StringVar(&cfg.cacheDir, "cache-dir", "cache", "Directory to store cached responses")
    flag.StringVar(&cfg.logFile, "log-file", "proxy.log", "File to log requests and responses")
    flag.StringVar(&cfg.promptFile, "prompt-file", "prompts.log", "File to log prompts")
    flag.StringVar(&cfg.costFile, "cost-file", "costs.log", "File to log API costs")
    flag.Parse()
    return cfg
}

func initLogger(filename string) *log.Logger {
    file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("Failed to open log file: %v", err)
    }
    return log.New(file, "", log.LstdFlags)
}
