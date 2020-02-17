package filecache

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	t.Parallel()

	giveGirPath := "foo"
	pool := NewPool(giveGirPath)

	if dirPath := pool.GetDirPath(); dirPath != "foo" {
		t.Errorf("Wrong directory path is set. Want: %s, got: %s", giveGirPath, dirPath)
	}
}

func TestPool_ClearItems(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	// Put "non-cache" file into temporary directory
	extraFilePath := filepath.Join(tmpDir, "extra_data")
	extraFile, _ := os.Create(extraFilePath)
	extraFile.Close()

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
	type fields struct {
		dirPath string
		mutex   *sync.Mutex
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &Pool{
				dirPath: tt.fields.dirPath,
				mutex:   tt.fields.mutex,
			}
			got, err := pool.DeleteItem(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeleteItem() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPool_GetDirPath(t *testing.T) {
	type fields struct {
		dirPath string
		mutex   *sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &Pool{
				dirPath: tt.fields.dirPath,
				mutex:   tt.fields.mutex,
			}
			if got := pool.GetDirPath(); got != tt.want {
				t.Errorf("GetDirPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPool_GetItem(t *testing.T) {
	type fields struct {
		dirPath string
		mutex   *sync.Mutex
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   CacheItem
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &Pool{
				dirPath: tt.fields.dirPath,
				mutex:   tt.fields.mutex,
			}
			if got := pool.GetItem(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPool_HasItem(t *testing.T) {
	type fields struct {
		dirPath string
		mutex   *sync.Mutex
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &Pool{
				dirPath: tt.fields.dirPath,
				mutex:   tt.fields.mutex,
			}
			if got := pool.HasItem(tt.args.key); got != tt.want {
				t.Errorf("HasItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPool_Put(t *testing.T) {
	type fields struct {
		dirPath string
		mutex   *sync.Mutex
	}
	type args struct {
		key       string
		from      io.Reader
		expiresAt time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    CacheItem
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &Pool{
				dirPath: tt.fields.dirPath,
				mutex:   tt.fields.mutex,
			}
			got, err := pool.Put(tt.args.key, tt.args.from, tt.args.expiresAt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Put() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Put() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPool_walkCacheFiles(t *testing.T) {
	type fields struct {
		dirPath string
		mutex   *sync.Mutex
	}
	type args struct {
		fn func(os.FileInfo)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &Pool{
				dirPath: tt.fields.dirPath,
				mutex:   tt.fields.mutex,
			}
			if err := pool.walkOverCacheFiles(tt.args.fn); (err != nil) != tt.wantErr {
				t.Errorf("walkOverCacheFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
