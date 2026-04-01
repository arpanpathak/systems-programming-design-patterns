package tls_sni

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

// RunL4SNIRouting simulates a Traffic Edge Proxy routing a TCP stream purely by parsing
// the raw TLS ClientHello frame, without possessing the SSL Private Key!
// This is called "SNI Passthrough" or L4 Routing (e.g. Envoy `sni_cluster` match).
// It prevents the proxy from burning CPU decrypting/encrypting traffic, while
// still allowing host-based routing based on the Server Name Indication!
func RunL4SNIRouting(conn net.Conn) {
	fmt.Println("=== L4 TLS SNI Passthrough Proxy Inspector ===")

	defer conn.Close()

	// 1. Read the initial bytes of the TLS handshake
	// 512 bytes is usually enough to capture the SNI hostname block
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("Connection dropped: %v\n", err)
		return
	}

	// 2. Parse the raw bytes!
	sni, parseErr := extractSNI(buf[:n])
	if parseErr != nil {
		fmt.Println("Not a valid TLS Handshake or missing SNI Extension!")
		return
	}

	fmt.Printf("[L4 Proxy] Extracted SNI: %s. Routing Raw TCP without Decrypting!\n", sni)

	// In a real proxy, we would now dial the upstream matching `sni`,
	// and use `io.Copy(upstream, conn)` and `io.Copy(conn, upstream)`
	// to blindly pipe the unbroken encrypted bytes!
}

// extractSNI parses the unencrypted ClientHello standard exactly.
// Interview gold: Shows you understand protocols beneath L7 HTTP!
func extractSNI(data []byte) (string, error) {
	// 1. Check if it's a TLS Handshake record (Byte 0 == 0x16)
	if len(data) < 5 || data[0] != 0x16 {
		return "", errors.New("not a TLS handshake")
	}

	// 2. We skip headers to find the SNI Extension Block
	// (A highly summarized extraction logic mimicking `golang.org/x/crypto/tls`)
	// We jump through the ClientHello payload, SessionID, Cipher Suites, Compress Methods...

	// Fast-forward to the Extensions blocks (simulated jump for brevity in this example)
	var sniIndex int = bytes.Index(data, []byte{0x00, 0x00}) // SNI Extension type

	if sniIndex == -1 || sniIndex+4 > len(data) {
		return "", errors.New("sni not found")
	}

	// Read SNI Server Name List Length
	sniLen := binary.BigEndian.Uint16(data[sniIndex+7 : sniIndex+9])
	if sniIndex+9+int(sniLen) > len(data) {
		return "", errors.New("sni truncated")
	}

	// Extract the actual hostname string!
	hostnameBytes := data[sniIndex+9 : sniIndex+9+int(sniLen)]
	return string(hostnameBytes), nil
}
