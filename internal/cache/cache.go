package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TokenCache struct {
	Token string `json:"token"`
	Exp   int64  `json:"exp"`
}

type Cache struct {
	filePath string
}

func New(filePath string) *Cache {
	return &Cache{filePath: filePath}
}

func DefaultCacheFile(tokenPath string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}
	sanitizedPath := strings.ReplaceAll(tokenPath, "/", "_")
	return filepath.Join(homeDir, ".kube", "vault_"+sanitizedPath+"_token.json")
}

func (c *Cache) Load() (string, bool) {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return "", false
	}

	var cache TokenCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return "", false
	}

	now := time.Now().Unix()
	if cache.Exp > now {
		return cache.Token, true
	}

	return "", false
}

func (c *Cache) Save(token string, exp int64) error {
	dir := filepath.Dir(c.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	cache := TokenCache{
		Token: token,
		Exp:   exp,
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return os.WriteFile(c.filePath, data, 0600)
}

func (c *Cache) Clear() error {
	return os.Remove(c.filePath)
}

func (c *Cache) FilePath() string {
	return c.filePath
}
