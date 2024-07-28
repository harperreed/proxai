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

func (s *Pro
