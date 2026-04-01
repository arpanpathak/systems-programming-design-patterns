package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// SetupReverseProxy is a foundational technique showing how Go writes L7 proxies naturally.
// Standard lib `httputil.ReverseProxy` is robust enough to power many production routers.
func SetupReverseProxy(targetRawURL string, listenAddress string) {
	// 1. Parse Upstream/Backend Destination URL
	targetUrl, err := url.Parse(targetRawURL)
	if err != nil {
		log.Fatalf("Invalid Upstream URL: %v", err)
	}

	// 2. Wrap via HttpUtil ReverseProxy
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)

	// Optional Customization: Director allows us to modify the Request IN FLIGHT!
	// (Essential for appending API Keys, trace IDs, headers like X-Forwarded-For).
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		// Retain default Director logic (Host, Scheme logic handling internally)
		originalDirector(req)

		// Let's add custom Trace headers
		req.Header.Set("X-Proxy-Traffic-Trace", "apple-proxy-id: 112345")

		// In a real load balancer, we would append the Client's REAL IP to the
		// existing X-Forwarded-For slice or initialize if empty.
		clientIP := req.RemoteAddr
		if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
			req.Header.Set("X-Forwarded-For", xff+", "+clientIP)
		} else {
			req.Header.Set("X-Forwarded-For", clientIP)
		}

		log.Printf("[Proxy Routing] Rewrote request heading to -> %s%s\n", targetUrl.Host, req.URL.Path)
	}

	// Start HTTP Server
	log.Printf("Starting Custom Reverse Proxy on %s routing to -> %s\n", listenAddress, targetRawURL)
	err = http.ListenAndServe(listenAddress, proxy)
	if err != nil {
		log.Fatalf("Proxy server failed: %v", err)
	}
}
