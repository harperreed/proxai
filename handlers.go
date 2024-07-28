package main

import (
	"html/template"
	"net/http"
	"os"
	"strings"
	"github.com/gomarkdown/markdown"
)

func (s *ProxyServer) helpHandler(w http.ResponseWriter, r *http.Request) {
	helpMarkdown, err := os.ReadFile("README.md")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	htmlContent := markdown.ToHTML(helpMarkdown, nil, nil)

	tmpl, err := template.New("help").Parse(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Proxai - OpenAI API Proxy</title>
		<link href="https://cdn.tailwindcss.com" rel="stylesheet">
	</head>
	<body class="bg-gray-100 text-red-900 font-sans">
		<div class="container mx-auto px-4 py-8">{{.Content}}</div>
	</body>
	</html>`)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, struct{ Content template.HTML }{template.HTML(htmlContent)})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *ProxyServer) handleLLMRequest(w http.ResponseWriter, r *http.Request) {
	// Check the path and route to the appropriate handler
	if strings.HasPrefix(r.URL.Path, "/v1/") {
		s.openAIHandler.HandleRequest(w, r)
	} else {
		http.Error(w, "Unsupported LLM provider", http.StatusBadRequest)
	}
}
