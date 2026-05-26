package localcache

import (
	"os"
	"path/filepath"
	"time"

	"github.com/wallissonmarinho/GoSubs/internal/core/domain"
)

type Cache struct {
	Dir string
	TTL time.Duration
}

func (c *Cache) Get(key string) ([]byte, bool, error) {
	path := c.path(key)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if c.TTL > 0 && time.Since(info.ModTime()) > c.TTL {
		_ = os.Remove(path)
		return nil, false, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

func (c *Cache) Put(key string, data []byte) error {
	if err := os.MkdirAll(c.Dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(c.path(key), data, 0o644)
}

func (c *Cache) Cleanup() (domain.CleanupResult, error) {
	var out domain.CleanupResult
	out.TTL = c.TTL.String()
	err := filepath.Walk(c.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		out.Scanned++
		if c.TTL > 0 && time.Since(info.ModTime()) > c.TTL {
			out.Removed++
			out.FreedBytes += info.Size()
			if rmErr := os.Remove(path); rmErr != nil && !os.IsNotExist(rmErr) {
				return rmErr
			}
		}
		return nil
	})
	if os.IsNotExist(err) {
		return out, nil
	}
	return out, err
}

func (c *Cache) path(key string) string {
	return filepath.Join(c.Dir, key+".srt")
}
