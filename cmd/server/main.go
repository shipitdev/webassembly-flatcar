package main

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/shipitdev/webassembly-flatcar/pkg/engine"
	"github.com/shipitdev/webassembly-flatcar/pkg/socket"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// wasip1 single-threaded scheduler pump: prevents false deadlock detection
	// when net.Accept blocks with no other goroutines in the run queue.
	go func() {
		for range time.Tick(100 * time.Millisecond) {}
	}()

	var totalRequests uint64
	startedAt := time.Now()
	store := engine.NewStateStore()

	mux := http.NewServeMux()

	// Core API routes
	mux.HandleFunc("/api/orders", engine.OrderHandler(store, &totalRequests))
	mux.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&totalRequests, 1)
		w.Write([]byte("PONG"))
	})
	mux.HandleFunc("/metrics", engine.MetricsHandler(store, &totalRequests, startedAt))

	listener, err := socket.GetListener(port)
	if err != nil {
		log.Fatalf("[FATAL] Failed to initialize socket listener on port %s: %v", port, err)
	}
	defer listener.Close()

	log.Printf("[SCALE-ENGINE] Service live on %s (Runtime: %s/%s)\n", listener.Addr().String(), runtime.GOOS, runtime.GOARCH)

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[FATAL] HTTP server exited unexpectedly: %v", err)
	}
}
