package filecache

import (
	"crypto/sha1" //nolint:gosec
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tarampampam/filecache.v1/file"
)

type Item struct {
	Pool     CachePool
	mutex    *sync.Mutex
	hashing  hash.Hash
	fileName string
	key      string
}

// DefaultItemFilePerms is default permissions for file, associated with cache item
var DefaultItemFilePerms os.FileMode = 0664

// DefaultItemFileSignature is default signature for cache files
var DefaultItemFileSignature file.FSignature = nil

// newItem creates cache item.
func newItem(pool CachePool, key string) *Item {
	item := &Item{
		Pool:    pool,
		mutex:   &sync.Mutex{},
		hashing: sha1.New(), //nolint:gosec
		key:     key,
	}

	// generate file name based on hashed key value
	item.fileName = item.keyToFileName(key)

	return item
}

// keyToFileName returns file name, based on key name.
func (i *Item) keyToFileName(key string) string {
	return hex.EncodeToString(i.hashing.Sum([]byte(key)))
}

// GetKey returns the key for the current cache item.
func (i *Item) GetKey() string { return i.key }

// GetFilePath returns path to the associated file.
func (i *Item) GetFilePath() string { return filepath.Join(i.Pool.GetDirPath(), i.fileName) }

// IsHit confirms if the cache item lookup resulted in a cache hit.
func (i *Item) IsHit() bool {
	i.mutex.Lock() // @todo: blocking is required here?
	defer i.mutex.Unlock()

	return i.isHit()
}

func (i *Item) isHit() bool {
	// check for file exists
	if info, err := os.Stat(i.GetFilePath()); err == nil && info.Mode().IsRegular() {
		return true
	}

	return false
}

// Get retrieves the value of the item from the cache associated with this object's key.
func (i *Item) Get(to io.Writer) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return i.get(to)
}

func (i *Item) get(to io.Writer) error {
	// try to open file for reading
	f, openErr := file.Open(i.GetFilePath(), DefaultItemFilePerms, DefaultItemFileSignature)
	if openErr != nil {
		return newError(ErrFileOpening, fmt.Sprintf("file [%s] cannot be opened", i.GetFilePath()), openErr)
	}
	defer func(f *file.File) { _ = f.Close() }(f)

	if err := f.GetData(to); err != nil {
		return newError(ErrFileReading, fmt.Sprintf("file [%s] read error", i.GetFilePath()), err)
	}

	return nil
}

// Set the value represented by this cache item.
func (i *Item) Set(from io.Reader) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return i.set(from)
}

// openOrCreateFile opens OR create file for item
func (i *Item) openOrCreateFile(filePath string, perm os.FileMode, signature file.FSignature) (*file.File, error) {
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

func (i *Item) set(from io.Reader) error {
	var filePath = i.GetFilePath()

	f, err := i.openOrCreateFile(filePath, DefaultItemFilePerms, DefaultItemFileSignature)
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
func (i *Item) IsExpired() (bool, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return i.isExpired()
}

func (i *Item) isExpired() (bool, error) {
	exp, expErr := i.expiresAt()

	if exp != nil {
		return exp.UnixNano() < time.Now().UnixNano(), nil
	}

	return false, newError(ErrExpirationDataNotAvailable, "expiration data reading error", expErr)
}

// ExpiresAt returns the expiration time for this cache item. If expiration doesn't set - nil will be returned.
// Important notice: returned time will be WITHOUT nanoseconds (just milliseconds).
func (i *Item) ExpiresAt() *time.Time {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	exp, _ := i.expiresAt()

	return exp
}

func (i *Item) expiresAt() (*time.Time, error) {
	f, openErr := file.Open(i.GetFilePath(), DefaultItemFilePerms, DefaultItemFileSignature)
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
func (i *Item) SetExpiresAt(when time.Time) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return i.setExpiresAt(when)
}

func (i *Item) setExpiresAt(when time.Time) error {
	f, err := i.openOrCreateFile(i.GetFilePath(), DefaultItemFilePerms, DefaultItemFileSignature)
	if err != nil {
		return err
	}
	defer func(f *file.File) { _ = f.Close() }(f)

	if err := f.SetExpiresAt(when); err != nil {
		return err
	}

	return nil
}
