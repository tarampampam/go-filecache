package filecache

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tarampampam/go-filecache/file"
)

type Item struct {
	Pool     CachePool
	hashing  hash.Hash
	fileName string
	key      string
	mutex    *sync.Mutex
}

// DefaultItemFilePerms is default permissions for file, associated with cache item
var DefaultItemFilePerms os.FileMode = 0664

// DefaultItemFileSignature is default signature for cache files
var DefaultItemFileSignature file.FSignature = nil

// newItem creates cache item.
func newItem(pool CachePool, key string) *Item {
	item := &Item{
		Pool:    pool,
		hashing: md5.New(), //nolint:gosec
		key:     key,
		mutex:   &sync.Mutex{},
	}

	// generate file name based on hashed key value
	item.fileName = item.keyToFileName(key)

	return item
}

// keyToFileName returns file name, based on key name.
func (item *Item) keyToFileName(key string) string {
	return hex.EncodeToString(item.hashing.Sum([]byte(key))) + ".cache"
}

// GetKey returns the key for the current cache item.
func (item *Item) GetKey() string { return item.key }

// GetFilePath returns path to the associated file.
func (item *Item) GetFilePath() string { return filepath.Join(item.Pool.GetDirPath(), item.fileName) }

// IsHit confirms if the cache item lookup resulted in a cache hit.
func (item *Item) IsHit() bool {
	item.mutex.Lock() // @todo: blocking is required here?
	defer item.mutex.Unlock()

	return item.isHit()
}

func (item *Item) isHit() bool {
	// check for file exists
	if info, err := os.Stat(item.GetFilePath()); err == nil && info.Mode().IsRegular() {
		return true
	}

	return false
}

// Get retrieves the value of the item from the cache associated with this object's key.
func (item *Item) Get(to io.Writer) error {
	item.mutex.Lock()
	defer item.mutex.Unlock()

	return item.get(to)
}

func (item *Item) get(to io.Writer) error {
	// try to open file for reading
	f, openErr := file.OpenRead(item.GetFilePath(), DefaultItemFileSignature)
	if openErr != nil {
		return newError(ErrFileOpening, fmt.Sprintf("file [%s] cannot be opened", item.GetFilePath()), openErr)
	}
	defer func(f *file.File) { _ = f.Close() }(f)

	if err := f.GetData(to); err != nil {
		return newError(ErrFileReading, fmt.Sprintf("file [%s] read error", item.GetFilePath()), err)
	}

	return nil
}

// Set the value represented by this cache item.
func (item *Item) Set(from io.Reader) error {
	item.mutex.Lock()
	defer item.mutex.Unlock()

	return item.set(from)
}

// openOrCreateFile opens OR create file for item
func (item *Item) openOrCreateFile(filePath string, perm os.FileMode, signature file.FSignature) (*file.File, error) {
	if info, err := os.Stat(filePath); err == nil && info.Mode().IsRegular() {
		opened, openErr := file.Open(filePath, perm, signature)
		if openErr != nil {
			return nil, newError(ErrFileOpening, fmt.Sprintf("file [%s] cannot be opened", filePath), openErr)
		}
		return opened, nil
	}

	created, createErr := file.Create(filePath, perm, signature)
	if createErr != nil {
		return nil, newError(ErrFileWriting, fmt.Sprintf("cannot create file [%s]", filePath), createErr)
	}
	return created, nil
}

func (item *Item) set(from io.Reader) error {
	var filePath = item.GetFilePath()

	f, err := item.openOrCreateFile(filePath, DefaultItemFilePerms, DefaultItemFileSignature)
	if err != nil {
		return err
	}
	defer func(f *file.File) { _ = f.Close() }(f)

	if err := f.SetData(from); err != nil {
		return newError(ErrFileWriting, fmt.Sprintf("cannot write into file [%s]", filePath), err)
	}

	return nil
}

// Indicates if cache item expiration time is exceeded. If expiration data was not set - error will be returned.
func (item *Item) IsExpired() (bool, error) {
	item.mutex.Lock()
	defer item.mutex.Unlock()

	return item.isExpired()
}

func (item *Item) isExpired() (bool, error) {
	exp, expErr := item.expiresAt()

	if exp != nil {
		return exp.UnixNano() < time.Now().UnixNano(), nil
	}

	return false, newError(ErrExpirationDataNotAvailable, "expiration data reading error", expErr)
}

// ExpiresAt returns the expiration time for this cache item. If expiration doesn't set - nil will be returned.
// Important notice: returned time will be WITHOUT nanoseconds (just milliseconds).
func (item *Item) ExpiresAt() *time.Time {
	item.mutex.Lock()
	defer item.mutex.Unlock()

	exp, _ := item.expiresAt()

	return exp
}

func (item *Item) expiresAt() (*time.Time, error) {
	f, openErr := file.Open(item.GetFilePath(), DefaultItemFilePerms, DefaultItemFileSignature)
	if openErr != nil {
		return nil, openErr
	}
	defer func(f *file.File) { _ = f.Close() }(f)

	exp, expErr := f.GetExpiresAt()

	if expErr != nil {
		return nil, expErr
	}

	return &exp, nil
}

// SetExpiresAt sets the expiration time for this cache item.
// Important notice: time will set WITHOUT nanoseconds (just milliseconds).
func (item *Item) SetExpiresAt(when time.Time) error {
	item.mutex.Lock()
	defer item.mutex.Unlock()

	return item.setExpiresAt(when)
}

func (item *Item) setExpiresAt(when time.Time) error {
	f, err := item.openOrCreateFile(item.GetFilePath(), DefaultItemFilePerms, DefaultItemFileSignature)
	if err != nil {
		return err
	}
	defer func(f *file.File) { _ = f.Close() }(f)

	if err := f.SetExpiresAt(when); err != nil {
		return err
	}

	return nil
}
