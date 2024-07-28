package main

import (
    "net/http"
    "sync"
    "time"
    "sync/atomic"
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
    atomic.AddInt64(&s.requestCount, 1)
}

func (s *ProxyServer) incrementTokensCount(tokens int) {
    atomic.AddInt64(&s.tokensCount, int64(tokens))
}

func (s *ProxyServer) addCost(cost float64) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.totalCost += cost
}

func (s *ProxyServer) getStats() (int64, int64, float64) {
    return atomic.LoadInt64(&s.requestCount),
           atomic.LoadInt64(&s.tokensCount),
           s.getTotalCost()
}

func (s *ProxyServer) getTotalCost() float64 {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    return s.totalCost
}

func (s *ProxyServer) resetCounters() {
    atomic.StoreInt64(&s.requestCount, 0)
    atomic.StoreInt64(&s.tokensCount, 0)
    s.mutex.Lock()
    s.totalCost = 0
    s.mutex.Unlock()
}
