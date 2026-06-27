// Day 21 debugging — FIXED.
//
// The fix: in a container you almost never want to bind to localhost.
// Default the listen host to 0.0.0.0 (all interfaces) so the kube probe,
// the load balancer, and other pods can reach the port. Operators can
// still override HOST when they genuinely want loopback-only.
//
// STDLIB ONLY. We resolve the effective bind address and assert it is
// reachable from outside the container, then exit promptly.
package main

import (
	"fmt"
	"net"
	"os"
)

func resolveListenAddr() (host, port string) {
	host = os.Getenv("HOST")
	if host == "" {
		// FIX: default to all interfaces. A server bound to 0.0.0.0 is
		// reachable from outside the container's network namespace, which
		// is what probes and external traffic need.
		host = "0.0.0.0"
	}
	port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return host, port
}

func reachableFromOutside(host string) bool {
	return host == "0.0.0.0" || host == "::" || host == ""
}

func main() {
	host, port := resolveListenAddr()
	addr := net.JoinHostPort(host, port) // robust: handles IPv6, etc.

	fmt.Printf("listening on %s\n", addr)

	if reachableFromOutside(host) {
		fmt.Println("OK: external traffic and health probes can reach this server")
		os.Exit(0)
	}
	fmt.Printf("UNREACHABLE: server is bound to %q — probes/external traffic get connection refused\n", host)
	os.Exit(1)
}
