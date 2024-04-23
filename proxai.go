package main

import (
    "bytes"
    // "crypto/md5"
    // "encoding/hex"
    "encoding/json"
    "flag"
    "fmt"
    "github.com/gomarkdown/markdown"
    "io/ioutil"
    "github.com/charmbracelet/lipgloss"
    "html/template"
    "log"
    "net/http"
    "strings"
    "time"
    "os"
    "os/exec"
    "os/signal"
    "syscall"
)

var (
    port    = flag.Int("port", 8080, "Port to listen on (default: 8080)")
    address = flag.String("address", "localhost", "Address to listen on (default: localhost)")
)

var (
    infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
    successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
    errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
    boldStyle    = lipgloss.NewStyle().Bold(true)
    statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("235")).Background(lipgloss.Color("11")).Padding(0, 1)
)

var (
    requestCount = 0
    tokensCount = 0
    quit         = make(chan os.Signal, 1)
)


func main() {
    flag.Parse()

    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
	    <-quit
	    clearStatusBar()
	    os.Exit(0)
	}()

    http.HandleFunc("/", openAIProxy)
    http.HandleFunc("/help", helpHandler)
    log.Println(infoStyle.Render("OpenAI Proxy Server is running on") + boldStyle.Render(fmt.Sprintf(" http://%s:%d", *address, *port)))
    log.Println(successStyle.Render("For integration help, visit ") + boldStyle.Render(fmt.Sprintf(" http://%s:%d", *address, *port)) + boldStyle.Render("/help"))
    go updateStatusBar()
    http.ListenAndServe(fmt.Sprintf("%s:%d", *address, *port), nil)
}

func helpHandler(w http.ResponseWriter, r *http.Request) {
    data := struct {
        Port    int
        Address string
    }{
        Port:    *port,
        Address: *address,
    }

    helpMarkdown, err := ioutil.ReadFile("README.md")
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    err = tmpl.Execute(w, data)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
}

func openAIProxy(w http.ResponseWriter, r *http.Request) {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
        http.Error(w, "Bad Request: Missing or malformed authorization header", http.StatusBadRequest)
        return
    }

    model := "utility"
    if r.Method == "POST" {
        defer r.Body.Close()
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }
        var requestBody map[string]interface{}
        err = json.Unmarshal(body, &requestBody)
        if err != nil {
            http.Error(w, "Bad Request: Invalid JSON body", http.StatusBadRequest)
            return
        }
        if modelValue, ok := requestBody["model"]; ok {
            model = modelValue.(string)
        }
        r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
    }

    // auth := strings.Split(authHeader, " ")[1]
    // authHash := md5.Sum([]byte(auth))
    // paramsHash := md5.Sum([]byte(r.URL.RawQuery))
    // cachePath := fmt.Sprintf("cache/%x/%s/%x", authHash, model, paramsHash)

    // Check cache here (omitted for simplicity)

    targetURL := fmt.Sprintf("https://api.openai.com%s", r.URL.Path)
    log.Println(infoStyle.Render(fmt.Sprintf("Proxying request to %s", targetURL)))


    req, err := http.NewRequest(r.Method, targetURL, r.Body)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    req.Header.Set("Authorization", authHeader)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    responseBody, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    var responseData map[string]interface{}
    err = json.Unmarshal(responseBody, &responseData)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    tokensUsed := 0
    if usage, ok := responseData["usage"].(map[string]interface{}); ok {
        tokensUsed = int(usage["total_tokens"].(float64))
    }

    // Save request and response details, model, tokens used, and timestamp to cache (omitted for simplicity)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(resp.StatusCode)
    w.Write(responseBody)
    requestCount++
    tokensCount += tokensUsed

    log.Println(successStyle.Render("Request:") + fmt.Sprintf(" %s %s", r.Method, r.URL))
    log.Println(infoStyle.Render("Model:") + fmt.Sprintf(" %s", model))
    log.Println(infoStyle.Render("Tokens Used:") + fmt.Sprintf(" %d", tokensUsed))
    log.Println(successStyle.Render("Response Status:") + fmt.Sprintf(" %d", resp.StatusCode))
    log.Println(infoStyle.Render("Timestamp:") + fmt.Sprintf(" %s", time.Now().Format(time.RFC3339)))
    log.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("--------------------"))
}

func updateStatusBar() {
    for {
        status := fmt.Sprintf("🐨: %s | Status: %s | Tokens: %d | Requests: %d", fmt.Sprintf("http://%s:%d", *address, *port), "✅",tokensCount, requestCount)
        statusBar := statusStyle.Render(status)

        // Move cursor to the bottom of the console
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

func init() {
    // Clear the console screen
    cmd := exec.Command("clear")
    cmd.Stdout = os.Stdout
    cmd.Run()
}
