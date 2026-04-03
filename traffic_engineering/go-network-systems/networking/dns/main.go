// DNS resolution: A, MX, NS, TXT, CNAME, SRV records + custom resolver.
package main

import (
	"context"
	"fmt"
	"net"
	"time"
)

func dnsResolution() {
	fmt.Println("\n--- Standard DNS Lookups ---")

	ips, err := net.LookupIP("google.com")
	if err != nil {
		fmt.Printf("  LookupIP error: %v\n", err)
	} else {
		for _, ip := range ips {
			fmt.Printf("  google.com -> %s (IPv4=%v)\n", ip, ip.To4() != nil)
		}
	}

	names, err := net.LookupAddr("8.8.8.8")
	if err != nil {
		fmt.Printf("  Reverse DNS error: %v\n", err)
	} else {
		for _, name := range names {
			fmt.Printf("  8.8.8.8 -> %s\n", name)
		}
	}

	mxRecords, err := net.LookupMX("google.com")
	if err != nil {
		fmt.Printf("  MX error: %v\n", err)
	} else {
		for _, mx := range mxRecords {
			fmt.Printf("  MX: %s (priority %d)\n", mx.Host, mx.Pref)
		}
	}

	nsRecords, err := net.LookupNS("google.com")
	if err != nil {
		fmt.Printf("  NS error: %v\n", err)
	} else {
		for _, ns := range nsRecords {
			fmt.Printf("  NS: %s\n", ns.Host)
		}
	}

	txtRecords, err := net.LookupTXT("google.com")
	if err != nil {
		fmt.Printf("  TXT error: %v\n", err)
	} else {
		for _, txt := range txtRecords {
			if len(txt) > 60 {
				txt = txt[:60] + "..."
			}
			fmt.Printf("  TXT: %s\n", txt)
		}
	}

	cname, err := net.LookupCNAME("www.google.com")
	if err != nil {
		fmt.Printf("  CNAME error: %v\n", err)
	} else {
		fmt.Printf("  CNAME: www.google.com -> %s\n", cname)
	}
}

func customResolver() {
	fmt.Println("\n--- Custom DNS Resolver ---")

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ips, err := resolver.LookupIPAddr(ctx, "example.com")
	if err != nil {
		fmt.Printf("  Custom resolver error: %v\n", err)
		return
	}
	for _, ip := range ips {
		fmt.Printf("  example.com (via 8.8.8.8) -> %s\n", ip.IP)
	}
}

func main() {
	fmt.Println("=== DNS Resolution ===")
	dnsResolution()
	customResolver()
}
