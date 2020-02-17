package filecache

import (
	"io"
	"time"
)

// Item defines an interface for interacting with objects inside a cache
type CacheItem interface {
	// Returns path to the associated file.
	GetFilePath() string

	// Returns the key for the current cache item.
	GetKey() string

	// Retrieves the value of the item from the cache associated with this object's key.
	Get(to io.Writer) error

	// Confirms if the cache item lookup resulted in a cache hit.
	IsHit() bool

	// Sets the value represented by this cache item.
	Set(from io.Reader) error

	// Returns the expiration time for this cache item. If expiration doesn't set - nil will be returned.
	ExpiresAt() *time.Time

	// Sets the expiration time for this cache item.
	SetExpiresAt(when time.Time) error
}

// Pool generates CacheItemInterface objects
type CachePool interface {
	// Returns cache directory path.
	GetDirPath() string

	// Returns a Cache Item representing the specified key.
	GetItem(key string) CacheItem

	// Confirms if the cache contains specified cache item.
	HasItem(key string) bool

	// Deletes all items in the pool.
	Clear() (bool, error)

	// Removes the item from the pool.
	DeleteItem(key string) (bool, error)

	// Put a cache item with expiring time.
	Put(key string, from io.Reader, expiresAt time.Time) (CacheItem, error)

	// Put a cache item without expiring time.
	PutForever(key string, from io.Reader) (CacheItem, error)
}
