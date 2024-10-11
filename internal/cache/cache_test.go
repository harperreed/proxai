package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

func TestNewCache(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test")
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
		t.Errorf("Cache store BasePath is incorrect. Got %s, want %s", cache.store.BasePath, tempDir)
	}
}

func TestSetAndGet(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cache := NewCache(tempDir)

	testCases := []struct {
		name  string
		key   string
		value interface{}
	}{
		{"string", "string_key", "test_value"},
		{"int", "int_key", 42},
		{"struct", "struct_key", struct{ Name string }{Name: "Test"}},
		{"empty", "empty_key", ""},
		{"nil", "nil_key", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valueBytes, _ := json.Marshal(tc.value)
			err := cache.Set(tc.key, valueBytes)
			if err != nil {
				t.Fatalf("Failed to set value: %v", err)
			}

			got, err := cache.Get(tc.key)
			if err != nil {
				t.Fatalf("Failed to get value: %v", err)
			}

			var gotValue interface{}
			err = json.Unmarshal(got, &gotValue)
			if err != nil {
				t.Fatalf("Failed to unmarshal value: %v", err)
			}

			if !reflect.DeepEqual(gotValue, tc.value) {
				t.Errorf("Got %v, want %v", gotValue, tc.value)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cache := NewCache(tempDir)

	key := "test_key"
	value := []byte("test_value")

	err = cache.Set(key, value)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	err = cache.Delete(key)
	if err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	_, err = cache.Get(key)
	if err == nil {
		t.Error("Expected error when getting deleted key, got nil")
	}

	err = cache.Delete("non_existent_key")
	if err != nil {
		t.Errorf("Expected no error when deleting non-existent key, got: %v", err)
	}
}

func TestGenerateCacheKey(t *testing.T) {
	testCases := []struct {
		name   string
		method string
		path   string
		body   map[string]interface{}
	}{
		{"simple", "GET", "/test", map[string]interface{}{"key": "value"}},
		{"empty body", "POST", "/api", map[string]interface{}{}},
		{"nil body", "PUT", "/update", nil},
		{"complex body", "POST", "/complex", map[string]interface{}{
			"nested": map[string]interface{}{
				"array": []int{1, 2, 3},
			},
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key1 := generateCacheKey(tc.method, tc.path, tc.body)
			key2 := generateCacheKey(tc.method, tc.path, tc.body)

			if key1 != key2 {
				t.Errorf("Generated keys are not identical for the same input")
			}

			if tc.name != "simple" {
				differentKey := generateCacheKey(tc.method+"X", tc.path, tc.body)
				if key1 == differentKey {
					t.Errorf("Generated keys are identical for different inputs")
				}
			}
		})
	}
}

func TestLargeValue(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cache := NewCache(tempDir)

	key := "large_key"
	value := make([]byte, 10*1024*1024) // 10 MB
	for i := range value {
		value[i] = byte(i % 256)
	}

	err = cache.Set(key, value)
	if err != nil {
		t.Fatalf("Failed to set large value: %v", err)
	}

	got, err := cache.Get(key)
	if err != nil {
		t.Fatalf("Failed to get large value: %v", err)
	}

	if !reflect.DeepEqual(got, value) {
		t.Error("Retrieved large value does not match the original")
	}
}

func TestConcurrentAccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cache := NewCache(tempDir)

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := filepath.Join("concurrent", string(rune('A'+id)))
			value := []byte{byte(id)}

			err := cache.Set(key, value)
			if err != nil {
				t.Errorf("Failed to set value in goroutine %d: %v", id, err)
				return
			}

			got, err := cache.Get(key)
			if err != nil {
				t.Errorf("Failed to get value in goroutine %d: %v", id, err)
				return
			}

			if !reflect.DeepEqual(got, value) {
				t.Errorf("Retrieved value does not match in goroutine %d", id)
			}
		}(i)
	}

	wg.Wait()
}
