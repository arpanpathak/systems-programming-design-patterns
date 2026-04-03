package main

import (
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

	file.Write([]byte("Hello World\n"))
	fmt.Println("Wrote simple ASCII text to test.txt")

	// --- 2. Literal "Binary String" Example ---
	binFile, err := os.OpenFile("test_data.bin", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer binFile.Close()

	var gibbirishTest string = "Some random bullshit..."

	// You want an ACTUAL binary string!
	// We will convert every single letter into its base-2 binary format (1s and 0s)!
	var literalBinaryString string
	for _, b := range []byte(gibbirishTest) {
		literalBinaryString += fmt.Sprintf("%08b ", b)
	}

	// Write the raw 1s and 0s to the file so you can physically see the binary form of the string!
	binFile.Write([]byte(literalBinaryString))

	fmt.Printf("Wrote the literal binary string format to test_data.bin:\n%s\n", literalBinaryString)
}
