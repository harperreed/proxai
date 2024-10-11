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
    totalCost    float64
    cache        *Cache
    logger       *Logger
    mutex        sync.RWMutex
}

func NewProxyServer(cacheDir, logDir string) (*ProxyServer, error) {
    cache := NewCache(cacheDir)
    logger, err := NewLogger(logDir)
    if err != nil {
        return nil, err
    }

    return &ProxyServer{
        client: &http.Client{Timeout: 30 * time.Second},
        cache:  cache,
        logger: logger,
    }, nil
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

func (s *ProxyServer) addCost(cost float64) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.totalCost += cost
}

func (s *ProxyServer) getStats() (int64, int64, float64) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    return s.requestCount, s.tokensCount, s.totalCost
}

func (s *ProxyServer) resetCounters() {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.requestCount = 0
    s.tokensCount = 0
    s.totalCost = 0
}
