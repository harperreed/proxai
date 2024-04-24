package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "strings"
    "time"
    "log"
    "github.com/charmbracelet/lipgloss"
    "github.com/gomarkdown/markdown"
    "html/template"
)

func (s *server) openAIProxy(w http.ResponseWriter, r *http.Request) {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
        handleHTTPError(w, s.logger, fmt.Errorf("missing or invalid authorization header"), http.StatusBadRequest)
        return
    }

    var requestBody map[string]interface{}

    model := "utility"
    if r.Method == "POST" {
        defer r.Body.Close()
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            handleHTTPError(w, s.logger, fmt.Errorf("failed to read request body: %w", err), http.StatusInternalServerError)
            return
        }

        err = json.Unmarshal(body, &requestBody)
        if err != nil {
            handleHTTPError(w, s.logger, fmt.Errorf("invalid JSON body: %w", err), http.StatusBadRequest)
            return
        }
        if modelValue, ok := requestBody["model"]; ok {
            model = modelValue.(string)
        }
        r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
    }

    jsonString, err := json.MarshalIndent(requestBody, "", "    ")
    if err != nil {
        handleHTTPError(w, s.logger, fmt.Errorf("failed to marshal JSON body: %w", err), http.StatusInternalServerError)
        return
    }

    // Check cache
    cacheKey := fmt.Sprintf("%s:%s", r.Method, r.URL.String())
    if cachedResponse, err := s.cache.Read(cacheKey); err == nil {
        s.logger.Println(successStyle.Render("Cache hit"))
        w.Header().Set("Content-Type", "application/json")
        w.Write(cachedResponse)
        return
    }

    targetURL := fmt.Sprintf("https://api.openai.com%s", r.URL.Path)
    s.logger.Println(infoStyle.Render(fmt.Sprintf("Proxying request to %s", targetURL)))

    req, err := http.NewRequest(r.Method, targetURL, r.Body)
    if err != nil {
        handleHTTPError(w, s.logger, fmt.Errorf("failed to create request: %w", err), http.StatusInternalServerError)
        return
    }
    req.Header.Set("Authorization", authHeader)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        handleHTTPError(w, s.logger, fmt.Errorf("failed to make request: %w", err), http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    responseBody, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        handleHTTPError(w, s.logger, fmt.Errorf("failed to read response body: %w", err), http.StatusInternalServerError)
        return
    }

    var responseData map[string]interface{}
    err = json.Unmarshal(responseBody, &responseData)
    if err != nil {
        handleHTTPError(w, s.logger, fmt.Errorf("failed to unmarshal response JSON: %w", err), http.StatusInternalServerError)
        return
    }

    tokensUsed := 0
    promptTokens := 0
    completionTokens := 0
    if usage, ok := responseData["usage"].(map[string]interface{}); ok {
        tokensUsed = int(usage["total_tokens"].(float64))
        promptTokens = int(usage["prompt_tokens"].(float64))
        completionTokens = int(usage["completion_tokens"].(float64))
    }

    // Cache response
    s.cache.Write(cacheKey, responseBody)

    // Log request and response
    s.logger.Printf("Request: %s %s\nResponse: %s\n", r.Method, r.URL, string(responseBody))

    // Log prompt
    if r.Method == "POST" {
        if prompt, ok := requestBody["prompt"].(string); ok {
            s.promptLogger.Println(prompt)
        }
    }

    // Update tokens used and total cost
    if model == "gpt-3.5-turbo" {
        tokensCount += tokensUsed
        totalCost += float64(tokensUsed) * 0.002 / 1000
    } else if model == "gpt-4" {
        tokensCount += tokensUsed
        totalCost += float64(tokensUsed) * 0.06 / 1000
    }
    s.costLogger.Printf("Tokens Used: %d, Total Cost: $%.5f\n", tokensUsed, totalCost)

    requestCount++

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(resp.StatusCode)
    w.Write(responseBody)

    s.model.content += fmt.Sprintf("\n%s%s", successStyle.Render("Request:"), fmt.Sprintf(" %s %s", r.Method, r.URL))
    s.model.content += fmt.Sprintf("\n%s%s", infoStyle.Render("Model:"), fmt.Sprintf(" %s", model))
    s.model.content += fmt.Sprintf("\n%s%s", infoStyle.Render("Prompt Tokens:"), fmt.Sprintf(" %d", promptTokens))
    s.model.content += fmt.Sprintf("\n%s%s", infoStyle.Render("Completion Tokens:"), fmt.Sprintf(" %d", completionTokens))
    s.model.content += fmt.Sprintf("\n%s%s", infoStyle.Render("Tokens Used:"), fmt.Sprintf(" %d", tokensUsed))
    s.model.content += fmt.Sprintf("\n%s%s", successStyle.Render("Response Status:"), fmt.Sprintf(" %d", resp.StatusCode))
    s.model.content += fmt.Sprintf("\n%s%s", infoStyle.Render("Timestamp:"), fmt.Sprintf(" %s", time.Now().Format(time.RFC3339)))
    s.model.content += fmt.Sprintf("\n%s", successStyle.Render("Request Body:"))
    s.model.content += fmt.Sprintf("\n%s", string(jsonString))
    s.model.content += fmt.Sprintf("\n%s", successStyle.Render("Response Body:"))
    s.model.content += fmt.Sprintf("\n%s", string(responseBody))
    s.model.content += fmt.Sprintf("\n%s", lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("--------------------"))
    s.program.Send(s.model.content)
}

func (s *server) helpHandler(w http.ResponseWriter, r *http.Request) {
    data := struct {
        Port    int
        Address string
    }{
        Port:    s.cfg.port,
        Address: s.cfg.address,
    }

    helpMarkdown, err := ioutil.ReadFile("README.md")
    if err != nil {
        handleHTTPError(w, s.logger, fmt.Errorf("failed to read README.md: %w", err), http.StatusInternalServerError)
        return
    }

    // Convert Markdown to HTML before rendering
    htmlContent := markdown.ToHTML(helpMarkdown, nil, nil)

    pageContent := `
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Proxai - OpenAI API Proxy</title>
        <link href="https://cdn.tailwindcss.com" rel="stylesheet">
    </head>
    <body class="bg-gray-100 text-red-900 font-sans">
        <div class="container mx-auto px-4 py-8">` + string(htmlContent) + `</div>
    </body>
    </html>`

    tmpl, err := template.New("help").Parse(string(pageContent))
    if err != nil {
        handleHTTPError(w, s.logger, fmt.Errorf("failed to parse help template: %w", err), http.StatusInternalServerError)
        return
    }

    err = tmpl.Execute(w, data)
    if err != nil {
        handleHTTPError(w, s.logger, fmt.Errorf("failed to execute help template: %w", err), http.StatusInternalServerError)
        return
    }
}

func handleHTTPError(w http.ResponseWriter, logger *log.Logger, err error, statusCode int) {
    logger.Printf("HTTP %d: %s", statusCode, err)
    http.Error(w, err.Error(), statusCode)
}
