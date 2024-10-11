package main

import (
	"os"
	"testing"
)

func TestNewCache(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cache := NewCache(tempDir)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}
	if cache.store == nil {
		t.Fatal("Cache store is nil")
	}
	if cache.store.BasePath != tempDir {
		t.Errorf("Expected BasePath %s, got %s", tempDir, cache.store.BasePath)
	}
}

func TestCacheSetAndGet(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cache := NewCache(tempDir)

	testCases := []struct {
		key   string
		value []byte
	}{
		{"key1", []byte("value1")},
		{"key2", []byte("value2")},
		{"key3", []byte("value3")},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			err := cache.Set(tc.key, tc.value)
			if err != nil {
				t.Fatalf("Failed to set cache: %v", err)
			}

			retrieved, err := cache.Get(tc.key)
			if err != nil {
				t.Fatalf("Failed to get cache: %v", err)
			}

			if string(retrieved) != string(tc.value) {
				t.Errorf("Expected %s, got %s", tc.value, retrieved)
			}
		})
	}
}

func TestCacheDelete(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cache := NewCache(tempDir)

	key := "test_key"
	value := []byte("test_value")

	err = cache.Set(key, value)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	err = cache.Delete(key)
	if err != nil {
		t.Fatalf("Failed to delete cache: %v", err)
	}

	_, err = cache.Get(key)
	if err == nil {
		t.Error("Expected error when getting deleted key, got nil")
	}
}

func TestGenerateCacheKey(t *testing.T) {
	testCases := []struct {
		method string
		path   string
		body   map[string]interface{}
	}{
		{"GET", "/test", map[string]interface{}{"param": "value"}},
		{"POST", "/api", map[string]interface{}{"data": 123}},
		{"PUT", "/update", map[string]interface{}{"id": 1, "name": "test"}},
	}

	for _, tc := range testCases {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			key1 := generateCacheKey(tc.method, tc.path, tc.body)
			key2 := generateCacheKey(tc.method, tc.path, tc.body)

			if key1 != key2 {
				t.Errorf("Expected identical keys for the same input, got %s and %s", key1, key2)
			}

			differentBody := map[string]interface{}{"different": "body"}
			key3 := generateCacheKey(tc.method, tc.path, differentBody)

			if key1 == key3 {
				t.Errorf("Expected different keys for different inputs, got %s and %s", key1, key3)
			}
		})
	}
}
