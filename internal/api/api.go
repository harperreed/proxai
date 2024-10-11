package api

import (
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/gomarkdown/markdown"
)

type FileReader func(filename string) ([]byte, error)
type TemplateParser func(name, text string) (*template.Template, error)

func HelpHandler(w http.ResponseWriter, r *http.Request, readFile FileReader, parseTemplate TemplateParser) {
	helpMarkdown, err := readFile("README.md")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	htmlContent := markdown.ToHTML(helpMarkdown, nil, nil)

	tmpl, err := parseTemplate("help", `
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

func DefaultHelpHandler(w http.ResponseWriter, r *http.Request) {
	HelpHandler(w, r, ioutil.ReadFile, template.New)
}
