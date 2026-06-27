// Day 25 walkthrough — rate limiting + graceful shutdown.
// Run: go run .   Hammer:  for i in $(seq 30); do curl -s -o /dev/null -w "%{http_code} " localhost:8080/; done
// Then Ctrl-C and watch the clean drain.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/time/rate"
)

// ---- per-IP rate limiter ------------------------------------------------
type ipLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	burst    int
}

func newIPLimiter(r rate.Limit, burst int) *ipLimiter {
	return &ipLimiter{limiters: map[string]*rate.Limiter{}, r: r, burst: burst}
}
func (l *ipLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if lim, ok := l.limiters[ip]; ok {
		return lim
	}
	lim := rate.NewLimiter(l.r, l.burst)
	l.limiters[ip] = lim
	return lim
}

func (l *ipLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if !l.get(ip).Allow() {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "rate limited", http.StatusTooManyRequests) // 429
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond) // simulate work (in-flight during drain)
		w.Write([]byte("ok\n"))
	})

	limiter := newIPLimiter(5, 10) // 5 req/s sustained, burst 10

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           limiter.middleware(mux),
		ReadTimeout:       5 * time.Second,  // NEVER use the zero-value (no timeouts)
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	// signal-aware context: cancelled on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("listening", "port", "8080", "limit", "5rps burst 10")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done() // wait for shutdown signal
	slog.Info("shutdown signal received, draining...")

	// drain in-flight requests, bounded (must be < orchestrator grace period)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("forced shutdown", "err", err)
	}
	slog.Info("stopped cleanly")
}
