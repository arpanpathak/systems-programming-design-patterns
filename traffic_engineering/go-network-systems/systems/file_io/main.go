// File I/O, directory operations, and atomic file writes.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func fileIODemo() {
	fmt.Println("\n--- File I/O ---")

	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "go-systems-test.txt")

	// Write file
	content := []byte("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		fmt.Printf("  Write error: %v\n", err)
		return
	}
	fmt.Printf("  Written %d bytes to %s\n", len(content), testFile)

	// Read entire file
	data, err := os.ReadFile(testFile)
	if err != nil {
		fmt.Printf("  Read error: %v\n", err)
		return
	}
	fmt.Printf("  Read %d bytes\n", len(data))

	// Buffered reading (efficient for large files)
	f, err := os.Open(testFile)
	if err != nil {
		fmt.Printf("  Open error: %v\n", err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		fmt.Printf("  Line %d: %s\n", lineNum, scanner.Text())
	}

	// Seek operations
	f.Seek(0, io.SeekStart)
	buf := make([]byte, 6)
	n, _ := f.Read(buf)
	fmt.Printf("  First %d bytes: %s\n", n, string(buf[:n]))

	// File stat
	info, err := os.Stat(testFile)
	if err == nil {
		fmt.Printf("  File: %s, Size: %d, ModTime: %v\n",
			info.Name(), info.Size(), info.ModTime().Format(time.RFC3339))
	}
	os.Remove(testFile)
}

func directoryDemo() {
	fmt.Println("\n--- Directory Operations ---")

	tmpBase := filepath.Join(os.TempDir(), "go-systems-dir-test")
	nestedDir := filepath.Join(tmpBase, "a", "b", "c")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		fmt.Printf("  MkdirAll error: %v\n", err)
		return
	}
	fmt.Printf("  Created: %s\n", nestedDir)

	for _, name := range []string{"file1.txt", "file2.go", "file3.txt"} {
		path := filepath.Join(tmpBase, name)
		os.WriteFile(path, []byte("test"), 0644)
	}

	fmt.Println("  Walking directory tree:")
	filepath.Walk(tmpBase, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(tmpBase, path)
		if info.IsDir() {
			fmt.Printf("    [DIR]  %s\n", relPath)
		} else {
			fmt.Printf("    [FILE] %s (%d bytes)\n", relPath, info.Size())
		}
		return nil
	})

	matches, _ := filepath.Glob(filepath.Join(tmpBase, "*.txt"))
	fmt.Printf("  *.txt files: %d\n", len(matches))
	os.RemoveAll(tmpBase)
}

func atomicWriteDemo() {
	fmt.Println("\n--- Atomic File Write ---")

	targetPath := filepath.Join(os.TempDir(), "go-atomic-write.txt")

	// Atomic write: write to temp file, then rename.
	// Ensures readers never see a partial write.
	atomicWrite := func(path string, data []byte) error {
		dir := filepath.Dir(path)
		tmpFile, err := os.CreateTemp(dir, "tmp-*")
		if err != nil {
			return err
		}
		tmpPath := tmpFile.Name()

		if _, err := tmpFile.Write(data); err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
			return err
		}
		if err := tmpFile.Sync(); err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
			return err
		}
		tmpFile.Close()
		return os.Rename(tmpPath, path)
	}

	if err := atomicWrite(targetPath, []byte("safely written data")); err != nil {
		fmt.Printf("  Atomic write error: %v\n", err)
		return
	}
	data, _ := os.ReadFile(targetPath)
	fmt.Printf("  Read back: %s\n", string(data))
	os.Remove(targetPath)
}

func main() {
	fmt.Println("=== File I/O & Directory Operations ===")
	fileIODemo()
	directoryDemo()
	atomicWriteDemo()
}
