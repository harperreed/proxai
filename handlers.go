package main

import (
    "encoding/json"
    "fmt"
    "html/template"
    "io"
    "log"
    "net/http"
    "os"
    "strings"
    "time"

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

func (s *ProxyServer) openAIProxy(w http.ResponseWriter, r *http.Request) {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
        http.Error(w, "Bad Request: Missing or malformed authorization header", http.StatusBadRequest)
        return
    }

    model := "utility"
    var requestBody map[string]interface{}

    if r.Method == http.MethodPost {
        if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
            http.Error(w, "Bad Request: Invalid JSON body", http.StatusBadRequest)
            return
        }
        if modelValue, ok := requestBody["model"]; ok {
            model = modelValue.(string)
        }
    }

    targetURL := fmt.Sprintf("https://api.openai.com%s", r.URL.Path)
    log.Println(infoStyle.Render(fmt.Sprintf("Proxying request to %s", targetURL)))

    proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    proxyReq.Header = r.Header
    proxyReq.Header.Set("Authorization", authHeader)
    proxyReq.Header.Set("Content-Type", "application/json")

    resp, err := s.client.Do(proxyReq)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        s.logger.LogResponse(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
        return
    }
    defer resp.Body.Close()

    responseBody, err := io.ReadAll(resp.Body)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    var responseData map[string]interface{}
    if err := json.Unmarshal(responseBody, &responseData); err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    tokensUsed := s.extractTokenUsage(responseData)

    s.incrementRequestCount()
    s.incrementTokensCount(tokensUsed)

    s.logRequestDetails(r, model, tokensUsed, resp.StatusCode, requestBody, responseBody)

    s.logger.LogRequest(r.Method, r.URL.Path, requestBody)
    s.logger.LogResponse(resp.StatusCode, responseData)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(resp.StatusCode)
    w.Write(responseBody)
}

func (s *ProxyServer) extractTokenUsage(responseData map[string]interface{}) int {
    if usage, ok := responseData["usage"].(map[string]interface{}); ok {
        if totalTokens, ok := usage["total_tokens"].(float64); ok {
            return int(totalTokens)
        }
    }
    return 0
}

func (s *ProxyServer) logRequestDetails(r *http.Request, model string, tokensUsed, statusCode int, requestBody map[string]interface{}, responseBody []byte) {
    log.Println(successStyle.Render("Request:") + fmt.Sprintf(" %s %s", r.Method, r.URL))
    log.Println(infoStyle.Render("Model:") + fmt.Sprintf(" %s", model))
    log.Println(infoStyle.Render("Tokens Used:") + fmt.Sprintf(" %d", tokensUsed))
    log.Println(successStyle.Render("Response Status:") + fmt.Sprintf(" %d", statusCode))
    log.Println(infoStyle.Render("Timestamp:") + fmt.Sprintf(" %s", time.Now().Format(time.RFC3339)))

    if requestBodyJSON, err := json.MarshalIndent(requestBody, "", "    "); err == nil {
        log.Println(successStyle.Render("Request Body:"))
        log.Println(string(requestBodyJSON))
    }

    log.Println(successStyle.Render("Response Body:"))
    log.Println(string(responseBody))
    log.Println(lipglossStyle.Render("--------------------"))
}
