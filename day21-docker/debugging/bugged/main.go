// Day 21 debugging — "works on my laptop, dead in the container".
//
// This service reads its listen address from config and starts an HTTP
// server. On a developer's machine the container appears to start fine
// (logs say "listening"), but every health probe and every request from
// outside the container fails with connection-refused, and Kubernetes
// kills the pod as "unhealthy".
//
// STDLIB ONLY. We do not actually open a socket here (so the demo exits
// promptly and runs offline) — we resolve the *effective* bind address
// the same way the real server would, then assert whether traffic from
// OUTSIDE the container could ever reach it. That assertion is the bug.
package main

import (
	"fmt"
	"os"
)

// resolveListenAddr builds the address the HTTP server will bind to.
//
// Intent: operators set HOST/PORT env vars in the container; if unset we
// fall back to a sensible default for local development.
func resolveListenAddr() string {
	host := os.Getenv("HOST")
	if host == "" {
		// BUG: defaulting to localhost. Inside a container, 127.0.0.1 is
		// the container's OWN loopback — nothing outside the container
		// (the kube-proxy probe, the service mesh, another pod) can reach
		// it. On a laptop you curl localhost so it "works"; in production
		// the port is bound but unreachable.
		host = "127.0.0.1"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return host + ":" + port
}

// reachableFromOutside reports whether traffic originating outside the
// container could reach a server bound to addr. A server bound to a
// loopback address is only reachable from within the same network
// namespace.
func reachableFromOutside(host string) bool {
	return host == "0.0.0.0" || host == "::" || host == ""
}

func main() {
	addr := resolveListenAddr()
	host := addr[:len(addr)-len(":8080")] // crude host extraction for demo
	if p := os.Getenv("PORT"); p != "" {
		host = addr[:len(addr)-len(":"+p)]
	}

	fmt.Printf("listening on %s\n", addr)

	if reachableFromOutside(host) {
		fmt.Println("OK: external traffic and health probes can reach this server")
		os.Exit(0)
	}
	fmt.Printf("UNREACHABLE: server is bound to %q — probes/external traffic get connection refused\n", host)
	os.Exit(1)
}
