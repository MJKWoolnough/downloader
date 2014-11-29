package cache

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"testing"
)

func TestPreAllocate(t *testing.T) {
	tests := []int64{
		0, 1, 2, 3, 4, 32, 64, 128, 512, 1024, 32 * 1024, 512 * 1024,
		1024 * 1024, 32 * 1024 * 1024, 32*1024*1024 + 3,
	}

	testPath, err := ioutil.TempDir("", "preallocation-test")
	defer os.RemoveAll(testPath)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	for n, test := range tests {
		f, err := os.Create(path.Join(testPath, "test-file-"+strconv.Itoa(n)))
		if err != nil {
			t.Errorf("unexpected error: %s", err)
			continue
		}
		err = preallocate(f, test)
		if err != nil {
			t.Errorf("test %d: unexpected error while allocating: %s", n+1, err)
		} else if fs, err := f.Stat(); err != nil {
			t.Fatalf("unexpected error: %s", err)
		} else if size := fs.Size(); size != test {
			t.Errorf("test %d: expecting size %d, got %d", n+1, test, size)
		}
		f.Close()
	}
}
