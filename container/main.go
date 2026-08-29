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

type HealthStatus struct {
	Status    string `json:"status"`
	Runtime   string `json:"runtime"`
	Arch      string `json:"arch"`
	Timestamp int64  `json:"timestamp"`
}

type SystemMetrics struct {
	Service       string `json:"service"`
	Runtime       string `json:"runtime"`
	UptimeSeconds int64  `json:"uptime_seconds"`
	AllocBytes    uint64 `json:"alloc_bytes"`
	TotalAlloc    uint64 `json:"total_alloc_bytes"`
	SysBytes      uint64 `json:"sys_bytes"`
	HeapAlloc     uint64 `json:"heap_alloc_bytes"`
	NumGC         uint32 `json:"num_gc"`
	NumGoroutine  int    `json:"num_goroutines"`
}

var startedAt = time.Now()

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthStatus{
			Status:    "ok",
			Runtime:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			Arch:      runtime.GOARCH,
			Timestamp: time.Now().Unix(),
		})
	})

	http.HandleFunc("/telemetry", func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SystemMetrics{
			Service:       "flatcar-container-edge-monitor",
			Runtime:       fmt.Sprintf("container/%s", runtime.GOARCH),
			UptimeSeconds: int64(time.Since(startedAt).Seconds()),
			AllocBytes:    m.Alloc,
			TotalAlloc:    m.TotalAlloc,
			SysBytes:      m.Sys,
			HeapAlloc:     m.HeapAlloc,
			NumGC:         m.NumGC,
			NumGoroutine:  runtime.NumGoroutine(),
		})
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP edge_service_uptime_seconds Total service uptime in seconds\n")
		fmt.Fprintf(w, "# TYPE edge_service_uptime_seconds counter\n")
		fmt.Fprintf(w, "edge_service_uptime_seconds %d\n\n", int64(time.Since(startedAt).Seconds()))

		fmt.Fprintf(w, "# HELP edge_mem_alloc_bytes Current allocated memory in bytes\n")
		fmt.Fprintf(w, "# TYPE edge_mem_alloc_bytes gauge\n")
		fmt.Fprintf(w, "edge_mem_alloc_bytes %d\n\n", m.Alloc)

		fmt.Fprintf(w, "# HELP edge_mem_sys_bytes Total memory obtained from system in bytes\n")
		fmt.Fprintf(w, "# TYPE edge_mem_sys_bytes gauge\n")
		fmt.Fprintf(w, "edge_mem_sys_bytes %d\n\n", m.Sys)

		fmt.Fprintf(w, "# HELP edge_goroutines_count Number of active goroutines\n")
		fmt.Fprintf(w, "# TYPE edge_goroutines_count gauge\n")
		fmt.Fprintf(w, "edge_goroutines_count %d\n", runtime.NumGoroutine())
	})

	serverAddr := fmt.Sprintf(":%s", port)
	log.Printf("[CONTAINER] edge monitor started on %s (runtime: %s/%s)\n", serverAddr, runtime.GOOS, runtime.GOARCH)

	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}
