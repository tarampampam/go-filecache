package filecache

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/tarampampam/go-filecache/file"
)

type Pool struct {
	dirPath string
}

// NewPool creates new cache items pool.
func NewPool(dirPath string) *Pool {
	return &Pool{
		dirPath: dirPath,
	}
}

// GetDirPath returns cache directory path.
func (pool *Pool) GetDirPath() string { return pool.dirPath }

// GetItem returns a Cache Item representing the specified key.
func (pool *Pool) GetItem(key string) CacheItem {
	item := newItem(pool, key)

	// Make check for exists and "is expired?"
	if item.IsHit() {
		if expired, _ := item.IsExpired(); expired {
			_, _ = pool.DeleteItem(key)
		}
	}

	return item
}

// HasItem confirms if the cache contains specified cache item.
func (pool *Pool) HasItem(key string) bool {
	return pool.GetItem(key).IsHit()
}

func (pool *Pool) walkOverCacheFiles(fn func(string, os.FileInfo)) error {
	files, err := ioutil.ReadDir(pool.dirPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		path := filepath.Join(pool.dirPath, f.Name())
		cacheFile, err := file.OpenRead(path, DefaultItemFileSignature)

		// skip "wrong" or errored file
		if err != nil || cacheFile == nil {
			continue
		}

		// verify file signature and close file (closing error will be skipped)
		matched, _ := cacheFile.SignatureMatched()

		if closeErr := cacheFile.Close(); matched && closeErr == nil {
			// if all is ok - fall the func
			fn(path, f)
		}
	}

	return nil
}

// Clear deletes all items in the pool.
func (pool *Pool) Clear() (bool, error) {
	var lastErr error

	err := pool.walkOverCacheFiles(func(path string, _ os.FileInfo) {
		if rmErr := os.Remove(path); rmErr != nil {
			lastErr = rmErr
		}
	})

	if err != nil {
		return false, err
	}

	if lastErr != nil {
		return false, lastErr
	}

	return true, nil
}

// DeleteItem removes the item from the pool.
func (pool *Pool) DeleteItem(key string) (bool, error) {
	item := newItem(pool, key)

	if rmErr := os.Remove(item.GetFilePath()); rmErr != nil {
		return false, rmErr
	}

	return true, nil
}

// Put a cache item with expiring time.
func (pool *Pool) Put(key string, from io.Reader, expiresAt time.Time) (CacheItem, error) {
	item, putError := pool.PutForever(key, from)

	if putError != nil {
		return item, putError
	}

	if err := item.SetExpiresAt(expiresAt); err != nil {
		return item, err
	}

	return item, nil
}

// Put a cache item without expiring time.
func (pool *Pool) PutForever(key string, from io.Reader) (CacheItem, error) {
	item := newItem(pool, key)

	if err := item.Set(from); err != nil {
		return item, err
	}

	return item, nil
}
