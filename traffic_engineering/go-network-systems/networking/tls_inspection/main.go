// TLS connection and certificate inspection.
package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("=== TLS Connection & Certificate Inspection ===")

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		},
	}

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp", "google.com:443", tlsConfig,
	)
	if err != nil {
		fmt.Printf("  TLS dial error: %v\n", err)
		return
	}
	defer conn.Close()

	state := conn.ConnectionState()
	fmt.Printf("  TLS Version: %x\n", state.Version)
	fmt.Printf("  Cipher Suite: %s\n", tls.CipherSuiteName(state.CipherSuite))
	fmt.Printf("  Server Name: %s\n", state.ServerName)
	fmt.Printf("  Negotiated Protocol: %s\n", state.NegotiatedProtocol)
	fmt.Printf("  Handshake Complete: %v\n", state.HandshakeComplete)

	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		fmt.Printf("  Certificate Subject: %s\n", cert.Subject)
		fmt.Printf("  Certificate Issuer: %s\n", cert.Issuer)
		fmt.Printf("  Valid: %s to %s\n",
			cert.NotBefore.Format("2006-01-02"),
			cert.NotAfter.Format("2006-01-02"))
		fmt.Printf("  DNS Names: %v\n", cert.DNSNames)
	}
}
