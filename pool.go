package filecache

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tarampampam/filecache.v1/file"
)

type Pool struct {
	dirPath string
	mutex   *sync.Mutex
}

// NewPool creates new cache items pool.
func NewPool(dirPath string) *Pool {
	return &Pool{
		dirPath: dirPath,
		mutex:   &sync.Mutex{},
	}
}

// GetDirPath returns cache directory path.
func (pool *Pool) GetDirPath() string { return pool.dirPath }

// GetItem returns a Cache Item representing the specified key.
func (pool *Pool) GetItem(key string) CacheItem {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	return pool.getItem(key)
}

// GetItem returns a Cache Item representing the specified key.
func (pool *Pool) getItem(key string) CacheItem {
	item := newItem(pool, key)

	// Make check for exists and "is expired?"
	if item.IsHit() {
		if expired, _ := item.IsExpired(); expired {
			_, _ = pool.deleteItem(key)
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
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

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
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	return pool.deleteItem(key)
}

func (pool *Pool) deleteItem(key string) (bool, error) {
	item := newItem(pool, key)

	if rmErr := os.Remove(item.GetFilePath()); rmErr != nil {
		return false, rmErr
	}

	return true, nil
}

// Put a cache item with expiring time.
func (pool *Pool) Put(key string, from io.Reader, expiresAt time.Time) (CacheItem, error) {
	item := newItem(pool, key)

	if err := item.Set(from); err != nil {
		return item, err
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
