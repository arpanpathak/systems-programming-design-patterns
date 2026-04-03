// Polling-based file watcher — detects modifications by checking ModTime.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type FileWatcher struct {
	path     string
	interval time.Duration
	onChange func(os.FileInfo)
	done     chan struct{}
}

func NewFileWatcher(path string, interval time.Duration, onChange func(os.FileInfo)) *FileWatcher {
	return &FileWatcher{
		path:     path,
		interval: interval,
		onChange: onChange,
		done:     make(chan struct{}),
	}
}

func (fw *FileWatcher) Start() {
	go func() {
		var lastModTime time.Time
		ticker := time.NewTicker(fw.interval)
		defer ticker.Stop()

		for {
			select {
			case <-fw.done:
				return
			case <-ticker.C:
				info, err := os.Stat(fw.path)
				if err != nil {
					continue
				}
				if info.ModTime().After(lastModTime) {
					lastModTime = info.ModTime()
					fw.onChange(info)
				}
			}
		}
	}()
}

func (fw *FileWatcher) Stop() {
	close(fw.done)
}

func main() {
	fmt.Println("=== File Watcher ===")

	testFile := filepath.Join(os.TempDir(), "go-watcher-test.txt")
	os.WriteFile(testFile, []byte("initial"), 0644)

	watcher := NewFileWatcher(testFile, 50*time.Millisecond, func(info os.FileInfo) {
		fmt.Printf("  File changed: %s (size=%d, mod=%v)\n",
			info.Name(), info.Size(), info.ModTime().Format("15:04:05.000"))
	})

	watcher.Start()

	time.Sleep(100 * time.Millisecond)
	os.WriteFile(testFile, []byte("modified content"), 0644)
	time.Sleep(100 * time.Millisecond)
	os.WriteFile(testFile, []byte("modified again with more"), 0644)
	time.Sleep(100 * time.Millisecond)

	watcher.Stop()
	os.Remove(testFile)
}
