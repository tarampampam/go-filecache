package filecache

import (
	"filecache/file"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
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
	return newItem(pool, key)
}

// HasItem confirms if the cache contains specified cache item.
func (pool *Pool) HasItem(key string) bool {
	return pool.GetItem(key).IsHit()
}

func (pool *Pool) walkOverCacheFiles(fn func(os.FileInfo)) error {
	files, err := ioutil.ReadDir(pool.dirPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		cacheFile, err := file.OpenRead(filepath.Join(pool.dirPath, f.Name()), DefaultItemFileSignature)

		// skip "wrong" or errored file
		if err != nil || cacheFile == nil {
			continue
		}

		// verify file signature and close file (closing error will be skipped)
		matched, _ := cacheFile.SignatureMatched()

		if closeErr := cacheFile.Close(); matched && closeErr == nil {
			// if all is ok - fall the func
			fn(f)
		}
	}

	return nil
}

// Clear deletes all items in the pool.
func (pool *Pool) Clear() (bool, error) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	var lastErr error

	err := pool.walkOverCacheFiles(func(info os.FileInfo) {
		if rmErr := os.Remove(filepath.Join(pool.dirPath, info.Name())); rmErr != nil {
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

	if rmErr := os.Remove(pool.GetItem(key).GetFilePath()); rmErr != nil {
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
