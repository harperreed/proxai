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
	port    = flag.Int("port", 8080, "Port to listen on")
	address = flag.String("address", "localhost", "Address to listen on")
)

func main() {
	flag.Parse()

	server := NewProxyServer()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		clearStatusBar()
		os.Exit(0)
	}()

	http.HandleFunc("/", server.openAIProxy)
	http.HandleFunc("/help", server.helpHandler)

	log.Println(infoStyle.Render("OpenAI Proxy Server is running on") + boldStyle.Render(fmt.Sprintf(" http://%s:%d", *address, *port)))
	log.Println(successStyle.Render("For integration help, visit ") + boldStyle.Render(fmt.Sprintf(" http://%s:%d/help", *address, *port)))

	go server.updateStatusBar()

	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", *address, *port), nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
