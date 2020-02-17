package filecache

import (
	"bytes"
	"testing"
	"time"
)

func TestUsageCreateAndSetCacheItem(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	content := []byte("bar")
	pool := NewPool(tmpDir)
	item, err := pool.Put("foo", bytes.NewBuffer(content), time.Now().Add(time.Second*10))

	if err != nil {
		t.Fatal(err)
	}

	if item.IsHit() != true {
		t.Errorf("IsHit() must return `true` for new cache item, but returns `false`")
	}

	getItem := pool.GetItem("foo")
	readBuffer := bytes.NewBuffer([]byte{})

	if getError := getItem.Get(readBuffer); getError != nil {
		t.Errorf("Unexpected error while gettind occured: %v", getError)
	}

	if !bytes.Equal(readBuffer.Bytes(), content) {
		t.Errorf("Got unexpected content from cache item. Want: %v, got: %v", content, readBuffer.Bytes())
	}
}
