package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

// Response structs for health check and memory telemetry
type HealthStatus struct {
	Status    string `json:"status"`
	Target    string `json:"target"`
	Timestamp int64  `json:"timestamp"`
}

type TelemetryPayload struct {
	Message    string `json:"message"`
	UptimeSec  int64  `json:"uptime_sec"`
	AllocBytes uint64 `json:"alloc_bytes"`
	NumGC      uint32 `json:"num_gc"`
}

var startedAt = time.Now()

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Basic healthcheck
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthStatus{
			Status:    "ok",
			Target:    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			Timestamp: time.Now().Unix(),
		})
	})

	// Benchmark / telemetry endpoint
	http.HandleFunc("/telemetry", func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TelemetryPayload{
			Message:    "Flatcar WASM Edge Service",
			UptimeSec:  int64(time.Since(startedAt).Seconds()),
			AllocBytes: m.Alloc,
			NumGC:      m.NumGC,
		})
	})

	// TODO: add graceful shutdown handler if needed once we test systemctl stop behavior in QEMU
	serverAddr := fmt.Sprintf(":%s", port)
	log.Printf("[WASM] listening on %s (runtime: %s/%s)\n", serverAddr, runtime.GOOS, runtime.GOARCH)

	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}
