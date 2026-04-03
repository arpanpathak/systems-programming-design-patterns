// Process execution, management, and environment variables.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func processExecution() {
	fmt.Println("\n--- Process Execution ---")

	// Run external command and capture output
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "echo Hello from subprocess")
	} else {
		cmd = exec.Command("echo", "Hello from subprocess")
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  Command output: %s", output)

	// Command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", "echo Timed command")
	} else {
		cmd = exec.CommandContext(ctx, "echo", "Timed command")
	}
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("  Timed command error: %v\n", err)
	} else {
		fmt.Printf("  Timed command output: %s", out)
	}

	// Piped command (stdin -> process -> stdout)
	if runtime.GOOS != "windows" {
		cmd = exec.Command("grep", "hello")
		cmd.Stdin = strings.NewReader("hello world\nfoo bar\nhello go\n")
		out, err = cmd.Output()
		if err != nil {
			fmt.Printf("  Piped grep error: %v\n", err)
		} else {
			fmt.Printf("  Piped grep output: %s", out)
		}
	}

	// Process info
	fmt.Printf("  PID: %d\n", os.Getpid())
	fmt.Printf("  PPID: %d\n", os.Getppid())
}

func environmentDemo() {
	fmt.Println("\n--- Environment Variables ---")

	// Set/Get
	os.Setenv("MY_APP_MODE", "testing")
	fmt.Printf("  MY_APP_MODE=%s\n", os.Getenv("MY_APP_MODE"))

	// LookupEnv distinguishes between empty and unset
	if val, exists := os.LookupEnv("MY_APP_MODE"); exists {
		fmt.Printf("  MY_APP_MODE exists: %s\n", val)
	}
	if _, exists := os.LookupEnv("NONEXISTENT_VAR"); !exists {
		fmt.Println("  NONEXISTENT_VAR does not exist")
	}

	// Expand environment variables in strings
	os.Setenv("GREETING", "Hello")
	expanded := os.ExpandEnv("$GREETING, World!")
	fmt.Printf("  Expanded: %s\n", expanded)

	// Cleanup
	os.Unsetenv("MY_APP_MODE")
	os.Unsetenv("GREETING")
}

func main() {
	fmt.Println("=== Process & Environment ===")
	processExecution()
	environmentDemo()
}
