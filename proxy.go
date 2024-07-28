package main

import (
	"net/http"
	"sync"
	"time"
)

type ProxyServer struct {
	client       *http.Client
	requestCount int64
	tokensCount  int64
	mutex        sync.RWMutex
}

func NewProxyServer() *ProxyServer {
	return &ProxyServer{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *ProxyServer) incrementRequestCount() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.requestCount++
}

func (s *ProxyServer) incrementTokensCount(tokens int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.tokensCount += int64(tokens)
}

func (s *ProxyServer) getStats() (int64, int64) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.requestCount, s.tokensCount
}
