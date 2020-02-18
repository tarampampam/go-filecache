package filecache

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func BenchmarkSetAndGet(b *testing.B) {
	tmpDir, _ := ioutil.TempDir("", "test-")
	defer func(b *testing.B) {
		if err := os.RemoveAll(tmpDir); err != nil {
			b.Fatal(err)
		}
	}(b)

	pool := NewPool(tmpDir)

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("%s_%d", "test_key", i)

		if _, err := pool.PutForever(key, bytes.NewBuffer([]byte(strings.Repeat("x", i)))); err != nil {
			b.Fatal(err)
		}

		item := pool.GetItem(key)

		if err := item.Get(bytes.NewBuffer([]byte{})); err != nil {
			b.Fatal(err)
		}

		if err := item.Set(bytes.NewBuffer([]byte(strings.Repeat("z", i)))); err != nil {
			b.Fatal(err)
		}
	}
}

func TestNewPool(t *testing.T) {
	t.Parallel()

	giveGirPath := "foo"
	pool := NewPool(giveGirPath)

	if dirPath := pool.GetDirPath(); dirPath != "foo" {
		t.Errorf("Wrong directory path is set. Want: %s, got: %s", giveGirPath, dirPath)
	}
}

func TestPool_Clear(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	// Put "non-cache" file into temporary directory
	extraFilePath := filepath.Join(tmpDir, "extra_data")
	extraFile, _ := os.Create(extraFilePath)
	if err := extraFile.Close(); err != nil {
		panic(err)
	}

	// check for "non-cache" file exists
	if info, err := os.Stat(extraFilePath); err != nil || !info.Mode().IsRegular() {
		t.Fatalf("Cannot create non-cache file by path: %s", extraFilePath)
	}

	pool := NewPool(tmpDir)

	tests := []struct {
		name string
		data []byte
	}{
		{name: "foo", data: []byte("foo")},
		{name: "bar", data: []byte("bar")},
	}

	// Write cache items
	for _, tt := range tests {
		if _, err := pool.PutForever(tt.name, bytes.NewBuffer(tt.data)); err != nil {
			t.Error(err)
		}
	}

	// Check for items is exists
	for _, tt := range tests {
		if pool.HasItem(tt.name) != true {
			t.Errorf("Got `false` for just created cache item")
		}
	}

	// Make clear
	if result, clearErr := pool.Clear(); result != true || clearErr != nil {
		t.Errorf("Clearing failed. Result is: %v, Error: %v", result, clearErr)
	}

	// Check for items is NOT exists
	for _, tt := range tests {
		if pool.HasItem(tt.name) != false {
			t.Errorf("Got non-`false` for deleted cache item named `%s`", tt.name)
		}
	}

	// check for "non-cache" still exists
	if info, err := os.Stat(extraFilePath); err != nil || !info.Mode().IsRegular() {
		t.Fatalf("Non-cache file by path %s does not exists", extraFilePath)
	}
}

func TestPool_DeleteItem(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	pool := NewPool(tmpDir)

	// Set two cache items
	if _, err := pool.PutForever("foo", bytes.NewBuffer([]byte("foo"))); err != nil {
		t.Error(err)
	}
	if _, err := pool.PutForever("bar", bytes.NewBuffer([]byte("bar"))); err != nil {
		t.Error(err)
	}

	// Delete one item
	if result, err := pool.DeleteItem("bar"); result != true || err != nil {
		t.Errorf("Error while item deleting. Result: %v, error: %v", result, err)
	}

	// Make checks
	if pool.HasItem("bar") != false {
		t.Errorf("Just deleted item must returns `false` on exists checking")
	}
	if pool.HasItem("foo") != true {
		t.Errorf("Previous item must steel exists")
	}
}

func TestPool_GetDirPath(t *testing.T) {
	t.Parallel()

	pool := NewPool("foo")

	if path := pool.GetDirPath(); path != "foo" {
		t.Errorf("Unepected dir path. Want: %s, got: %s", "foo", path)
	}
}

func TestPool_GetItem(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	pool := NewPool(tmpDir)

	tests := []struct {
		name string
		data []byte
	}{
		{name: "foo", data: []byte("foo")},
		{name: "bar", data: []byte("bar")},
	}

	// Write cache items
	for _, tt := range tests {
		if _, err := pool.PutForever(tt.name, bytes.NewBuffer(tt.data)); err != nil {
			t.Error(err)
		}
	}

	// Check for items is exists
	for _, tt := range tests {
		if pool.HasItem(tt.name) != true {
			t.Errorf("Got `false` for just created cache item")
		}
	}

	// Check for items is NOT exists
	for _, tt := range tests {
		buf := bytes.NewBuffer([]byte{})

		if err := pool.GetItem(tt.name).Get(buf); err != nil || !bytes.Equal(buf.Bytes(), tt.data) {
			t.Errorf("Got wrong content for %s. Want: %v, got: %v", tt.name, tt.data, buf.Bytes())
		}
	}
}

func TestPool_Put(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	pool := NewPool(tmpDir)

	// Set items with "expires at" data
	if _, err := pool.Put("foo", bytes.NewBuffer([]byte("foo")), time.Now().Add(time.Millisecond*100)); err != nil {
		t.Error(err)
	}
	if _, err := pool.Put("bar", bytes.NewBuffer([]byte("bar")), time.Now().Add(time.Millisecond*200)); err != nil {
		t.Error(err)
	}

	// Wait for some time
	time.Sleep(time.Millisecond * 101)

	// And then check availability
	if pool.HasItem("foo") != false {
		t.Errorf("Expired cache item must be not available")
	}
	if pool.HasItem("bar") != true {
		t.Errorf("Non-expired cache item must be available")
	}

	// Wait for some time again
	time.Sleep(time.Millisecond * 100)

	if pool.HasItem("bar") != false {
		t.Errorf("Expired cache item must be not available")
	}
}

func TestPool_PutForever(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	pool := NewPool(tmpDir)

	if _, err := pool.PutForever("foo", bytes.NewBuffer([]byte("foo"))); err != nil {
		t.Error(err)
	}

	// Wait for some time
	time.Sleep(time.Millisecond * 20)

	if pool.HasItem("foo") != true {
		t.Errorf("Never expired cache item must always be available")
	}
}

func TestPool_ConcurrentUsage(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	pool := NewPool(tmpDir)

	if _, err := pool.PutForever("foo", bytes.NewBuffer([]byte("foo"))); err != nil {
		t.Error(err)
	}

	wg := sync.WaitGroup{}

	for i := 0; i < 1024; i++ {
		wg.Add(2)
		go func(pool *Pool) {
			defer wg.Done()

			item := pool.GetItem("foo")

			if err := item.Get(bytes.NewBuffer([]byte{})); err != nil {
				t.Errorf("Got unexpected error on data GET: %v", err)
			}
		}(pool)
		go func(pool *Pool) {
			defer wg.Done()

			item := pool.GetItem("foo")

			if err := item.Set(bytes.NewBuffer([]byte(strings.Repeat("z", 32)))); err != nil {
				t.Errorf("Got unexpected error on data SET: %v", err)
			}
		}(pool)
	}

	wg.Wait()
}
