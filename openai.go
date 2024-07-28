package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type OpenAIHandler struct {
	client *http.Client
	logger *Logger
}

func NewOpenAIHandler(client *http.Client, logger *Logger) *OpenAIHandler {
	return &OpenAIHandler{
		client: client,
		logger: logger,
	}
}

func (h *OpenAIHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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

	tokensUsed := h.extractTokenUsage(responseData)

	h.logRequestDetails(r, model, tokensUsed, resp.StatusCode, requestBody, responseBody, responseData)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(responseBody)
}

func (h *OpenAIHandler) extractTokenUsage(responseData map[string]interface{}) int {
	if usage, ok := responseData["usage"].(map[string]interface{}); ok {
		if totalTokens, ok := usage["total_tokens"].(float64); ok {
			return int(totalTokens)
		}
	}
	return 0
}

func (h *OpenAIHandler) logRequestDetails(r *http.Request, model string, tokensUsed, statusCode int, requestBody map[string]interface{}, responseBody []byte, responseData map[string]interface{}) {
	log.Println(successStyle.Render("Request:") + fmt.Sprintf(" %s %s", r.Method, r.URL))
	log.Println(infoStyle.Render("Model:") + fmt.Sprintf(" %s", model))
	log.Println(infoStyle.Render("Tokens Used:") + fmt.Sprintf(" %d", tokensUsed))
	log.Println(successStyle.Render("Response Status:") + fmt.Sprintf(" %d", statusCode))

	if requestBodyJSON, err := json.MarshalIndent(requestBody, "", "    "); err == nil {
		log.Println(successStyle.Render("Request Body:"))
		log.Println(string(requestBodyJSON))
	}

	log.Println(successStyle.Render("Response Body:"))
	log.Println(string(responseBody))
	log.Println(lipglossStyle.Render("--------------------"))

	h.logger.LogRequest(r.Method, r.URL.Path, requestBody)
	h.logger.LogResponse(statusCode, responseData)
	h.logger.LogCost(model, tokensUsed, calculateCost(model, tokensUsed))
}
