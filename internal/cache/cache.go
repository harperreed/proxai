package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"

	"github.com/peterbourgon/diskv"
)

type Cache struct {
	store *diskv.Diskv
}

func NewCache(dir string) *Cache {
	flatTransform := func(s string) []string { return []string{} }
	d := diskv.New(diskv.Options{
		BasePath:     dir,
		Transform:    flatTransform,
		CacheSizeMax: 1024 * 1024 * 1024, // 1GB
	})
	return &Cache{store: d}
}

func (c *Cache) Set(key string, value []byte) error {
	return c.store.Write(key, value)
}

func (c *Cache) Get(key string) ([]byte, error) {
	return c.store.Read(key)
}

func (c *Cache) Delete(key string) error {
	return c.store.Erase(key)
}

func generateCacheKey(method, path string, body map[string]interface{}) string {
	data, _ := json.Marshal(body)
	hash := md5.Sum(append([]byte(method+path), data...))
	return hex.EncodeToString(hash[:])
}
