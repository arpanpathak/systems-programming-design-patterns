package main

import (
	"encoding/hex"
	"fmt"
	"os"
)

func main() {
	// --- 1. Plain Text Example ---
	file, err := os.OpenFile("test.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, _ = file.Write([]byte("Hello World\n"))
	fmt.Println("Wrote simple ASCII text to test.txt")

	// --- 2. Hex Binary Encoding (The easy way) ---
	// "Binary-encoded" strings are usually hexadecimal or base64.
	// Hex is the standard systems representation of raw binary data.
	data := "Some random bullshit..."
	encoded := hex.EncodeToString([]byte(data))

	// os.WriteFile is the "fucking easy way" to write a file in one call.
	err = os.WriteFile("test_data.hex", []byte(encoded), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Wrote hex-encoded binary string to test_data.hex: %s\n", encoded)
}
