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

	"github.com/charmbracelet/lipgloss"
	"github.com/harperreed/proxai/internal/api"
	"github.com/harperreed/proxai/internal/proxy"
	"github.com/harperreed/proxai/internal/ui"
)

var (
	port     = flag.Int("port", 8080, "Port to listen on")
	address  = flag.String("address", "localhost", "Address to listen on")
	cacheDir = flag.String("cache-dir", "./cache", "Directory for caching responses")
	logDir   = flag.String("log-dir", "./logs", "Directory for log files")
)

func main() {
	flag.Parse()

	server, err := proxy.NewProxyServer(*cacheDir, *logDir)
	if err != nil {
		log.Fatalf("Failed to create proxy server: %v", err)
	}

	// Create a new serve mux and server
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.OpenAIProxy)
	mux.HandleFunc("/help", api.HelpHandler)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", *address, *port),
		Handler: mux,
	}

	// Start the server in a goroutine
	go func() {
		log.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("OpenAI Proxy Server is running on") +
			lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf(" http://%s:%d", *address, *port)))

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Set up signal handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start the UI
	go func() {
		if err := ui.StartUI(server); err != nil {
			log.Printf("Error running UI: %v", err)
			stop <- os.Interrupt
		}
	}()

	<-stop // Wait for SIGINT or SIGTERM

	log.Println("Shutting down server...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	server.Logger.Close()
	log.Println("Server exiting")
}
