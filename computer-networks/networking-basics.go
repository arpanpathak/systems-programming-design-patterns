package main

import (
	"fmt"
	"net"
)

func startTCPServer() {
	// Basic bitch TCP/IP Server
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Server started on :8080 and waiting for incomign connections")
	// Handle incoming requests
	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println(err)
			continue
		}

		go handleRequest(conn)

	}
}

func handleRequest(conn net.Conn) {
	// Close the connection once done
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)

		if err != nil {
			return
		}

		fmt.Printf("Received %s", string(buf[:n]))
	}
}

func main() {

	// Start the TCP/IP Server
	startTCPServer()
}
