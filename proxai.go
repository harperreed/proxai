package main

import (
    "bytes"
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
    "bufio"
    // "path/filepath"
    "github.com/eiannone/keyboard"
    "github.com/peterbourgon/diskv"
)

var (
    port       = flag.Int("port", 8080, "Port to listen on (default: 8080)")
    address    = flag.String("address", "localhost", "Address to listen on (default: localhost)")
    cacheDir   = flag.String("cache-dir", "cache", "Directory to store cached responses")
    logFile    = flag.String("log-file", "proxy.log", "File to log requests and responses")
    promptFile = flag.String("prompt-file", "prompts.log", "File to log prompts")
    costFile   = flag.String("cost-file", "costs.log", "File to log API costs")
)

var (
    infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
    successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
    errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
    boldStyle    = lipgloss.NewStyle().Bold(true)
    statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("235")).Background(lipgloss.Color("11")).Padding(0, 1)
)

var (
    requestCount   = 0
    tokensCount    = 0
    totalCost      = 0.0
    quit           = make(chan os.Signal, 1)
    cache          *diskv.Diskv
    logger         *log.Logger
    promptLogger   *log.Logger
    costLogger     *log.Logger
    reader         = bufio.NewReader(os.Stdin)
    commandChannel = make(chan string)
)

func main() {
    flag.Parse()

    cache = diskv.New(diskv.Options{
        BasePath:     *cacheDir,
        CacheSizeMax: 100 * 1024 * 1024, // 100MB cache size
    })

    logFile, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Fatal("Failed to open log file:", err)
    }
    defer logFile.Close()
    logger = log.New(logFile, "", log.LstdFlags)

    promptLogFile, err := os.OpenFile(*promptFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Fatal("Failed to open prompt log file:", err)
    }
    defer promptLogFile.Close()
    promptLogger = log.New(promptLogFile, "", log.LstdFlags)

    costLogFile, err := os.OpenFile(*costFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Fatal("Failed to open cost log file:", err)
    }
    defer costLogFile.Close()
    costLogger = log.New(costLogFile, "", log.LstdFlags)

    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-quit
        clearStatusBar()
        fmt.Println("Exiting...")
        os.Exit(0)
    }()

    go handleKeyboardCommands()

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

    var requestBody map[string]interface{}

    model := "utility"
    if r.Method == "POST" {
        defer r.Body.Close()
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

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

    jsonString, err := json.MarshalIndent(requestBody, "", "    ")
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Check cache
    cacheKey := fmt.Sprintf("%s:%s", r.Method, r.URL.String())
    if cachedResponse, err := cache.Read(cacheKey); err == nil {
        log.Println(successStyle.Render("Cache hit"))
        w.Header().Set("Content-Type", "application/json")
        w.Write(cachedResponse)
        return
    }

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
    promptTokens := 0
    completionTokens := 0
    if usage, ok := responseData["usage"].(map[string]interface{}); ok {
        tokensUsed = int(usage["total_tokens"].(float64))
        promptTokens = int(usage["prompt_tokens"].(float64))
        completionTokens = int(usage["completion_tokens"].(float64))
    }

    // Cache response
    cache.Write(cacheKey, responseBody)


    // Log prompt
    if r.Method == "POST" {
        if prompt, ok := requestBody["prompt"].(string); ok {
            promptLogger.Println(prompt)
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
    costLogger.Printf("Tokens Used: %d, Total Cost: $%.5f\n", tokensUsed, totalCost)

    requestCount++

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(resp.StatusCode)
    w.Write(responseBody)

    log.Println(successStyle.Render("Request:") + fmt.Sprintf(" %s %s", r.Method, r.URL))
    log.Println(infoStyle.Render("Model:") + fmt.Sprintf(" %s", model))
    log.Println(infoStyle.Render("Prompt Tokens:") + fmt.Sprintf(" %d", promptTokens))
    log.Println(infoStyle.Render("Completion Tokens:") + fmt.Sprintf(" %d", completionTokens))
    log.Println(infoStyle.Render("Tokens Used:") + fmt.Sprintf(" %d", tokensUsed))
    log.Println(successStyle.Render("Response Status:") + fmt.Sprintf(" %d", resp.StatusCode))
    log.Println(infoStyle.Render("Timestamp:") + fmt.Sprintf(" %s", time.Now().Format(time.RFC3339)))
    log.Println(successStyle.Render("Request Body:"))
    log.Println("\n" + string(jsonString))
    log.Println(successStyle.Render("Response Body:"))
    log.Println("\n" +string(responseBody))
    log.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("\n\n--------------------\n\n"))

    // Log request and response

    logger.Printf("Method: %s %s, Model: %s, Tokens used: %d\n\nRequest:\n%s\nResponse:\n%s\n\n---------------------------\n", r.Method, r.URL,model, tokensUsed, string(jsonString), string(responseBody))


}

func updateStatusBar() {
    for {
        select {
        case <-quit:
            return
        default:
            status := fmt.Sprintf("🐨: %s | Status: %s | Tokens: %d | Requests: %d | Total Cost: $%.5f",
                fmt.Sprintf("http://%s:%d", *address, *port), "✅", tokensCount, requestCount, totalCost)
            statusBar := statusStyle.Render(status)

            fmt.Print("\033[s")
            fmt.Print("\033[999B")
            fmt.Print("\r")
            fmt.Print(statusBar)
            fmt.Print("\033[u")

            time.Sleep(1 * time.Second)
        }
    }
}

func clearStatusBar() {
    fmt.Print("\033[s")
    fmt.Print("\033[999B")
    fmt.Print("\r")
    fmt.Print("\033[K")
    fmt.Print("\033[u")
}

func handleKeyboardCommands() {
    err := keyboard.Open()
    if err != nil {
        log.Fatal(err)
    }
    defer keyboard.Close()

    for {
        char, key, err := keyboard.GetKey()
        if err != nil {
            log.Fatal(err)
        }

        if key == keyboard.KeyCtrlC {
            clearStatusBar()
            fmt.Println("Exiting...")
            os.Exit(0)
        }

        switch char {
        case 'r', 'R':
            requestCount = 0
            tokensCount = 0
            totalCost = 0.0
            fmt.Println("Counters reset.")
        case 'c', 'C':
            cmd := exec.Command("clear")
            cmd.Stdout = os.Stdout
            cmd.Run()
        }
    }
}

func init() {
    cmd := exec.Command("clear")
    cmd.Stdout = os.Stdout
    cmd.Run()
}
