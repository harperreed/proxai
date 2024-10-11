package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/harperreed/proxai/internal/cache"
	"github.com/harperreed/proxai/internal/logger"
	"github.com/harperreed/proxai/internal/utils"
)

type ProxyServer struct {
	Client       *http.Client
	RequestCount int64
	TokensCount  int64
	TotalCost    float64
	Cache        *cache.Cache
	Logger       *logger.Logger
	Mutex        sync.RWMutex
}

func NewProxyServer(cacheDir, logDir string) (*ProxyServer, error) {
	cache := cache.NewCache(cacheDir)
	logger, err := logger.NewLogger(logDir)
	if err != nil {
		return nil, err
	}

	return &ProxyServer{
		Client: &http.Client{Timeout: 30 * time.Second},
		Cache:  cache,
		Logger: logger,
	}, nil
}

func (s *ProxyServer) incrementRequestCount() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.RequestCount++
}

func (s *ProxyServer) incrementTokensCount(tokens int) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.TokensCount += int64(tokens)
}

func (s *ProxyServer) addCost(cost float64) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.TotalCost += cost
}

func (s *ProxyServer) ResetCounters() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.RequestCount = 0
	s.TokensCount = 0
	s.TotalCost = 0
}

func (s *ProxyServer) GetStats() (int64, int64, float64) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.RequestCount, s.TokensCount, s.TotalCost
}

func (s *ProxyServer) OpenAIProxy(w http.ResponseWriter, r *http.Request) {

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
	log.Println(utils.InfoStyle.Render(fmt.Sprintf("Proxying request to %s", targetURL)))

	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	proxyReq.Header = r.Header
	proxyReq.Header.Set("Authorization", authHeader)
	proxyReq.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(proxyReq)
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

	tokensUsed := s.extractTokenUsage(responseData)

	s.incrementRequestCount()
	s.incrementTokensCount(tokensUsed)

	s.logRequestDetails(r, model, tokensUsed, resp.StatusCode, requestBody, responseBody)

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
	log.Println(utils.SuccessStyle.Render("Request:") + fmt.Sprintf(" %s %s", r.Method, r.URL))
	log.Println(utils.InfoStyle.Render("Model:") + fmt.Sprintf(" %s", model))
	log.Println(utils.InfoStyle.Render("Tokens Used:") + fmt.Sprintf(" %d", tokensUsed))
	log.Println(utils.SuccessStyle.Render("Response Status:") + fmt.Sprintf(" %d", statusCode))
	log.Println(utils.InfoStyle.Render("Timestamp:") + fmt.Sprintf(" %s", time.Now().Format(time.RFC3339)))

	if requestBodyJSON, err := json.MarshalIndent(requestBody, "", "    "); err == nil {
		log.Println(utils.SuccessStyle.Render("Request Body:"))
		log.Println(string(requestBodyJSON))
	}

	log.Println(utils.SuccessStyle.Render("Response Body:"))
	log.Println(string(responseBody))
	log.Println(utils.LipglossStyle.Render("--------------------"))
}
