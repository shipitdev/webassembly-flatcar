package engine

import (
	"fmt"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

// MetricsHandler exports standard Prometheus format metrics for scraper telemetry.
func MetricsHandler(store *StateStore, totalReqs *uint64, startedAt time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP http_requests_total Total HTTP requests handled\n")
		fmt.Fprintf(w, "# TYPE http_requests_total counter\n")
		fmt.Fprintf(w, "http_requests_total %d\n\n", atomic.LoadUint64(totalReqs))

		fmt.Fprintf(w, "# HELP orders_cached_active In-memory active order count\n")
		fmt.Fprintf(w, "# TYPE orders_cached_active gauge\n")
		fmt.Fprintf(w, "orders_cached_active %d\n\n", store.Count())

		fmt.Fprintf(w, "# HELP revenue_usd_total Total gross revenue processed in USD\n")
		fmt.Fprintf(w, "# TYPE revenue_usd_total counter\n")
		fmt.Fprintf(w, "revenue_usd_total %.2f\n\n", store.RevenueUSD())

		fmt.Fprintf(w, "# HELP memory_heap_bytes Current Go heap allocation\n")
		fmt.Fprintf(w, "# TYPE memory_heap_bytes gauge\n")
		fmt.Fprintf(w, "memory_heap_bytes %d\n\n", m.Alloc)

		fmt.Fprintf(w, "# HELP uptime_seconds Service runtime uptime in seconds\n")
		fmt.Fprintf(w, "# TYPE uptime_seconds counter\n")
		fmt.Fprintf(w, "uptime_seconds %d\n", int64(time.Since(startedAt).Seconds()))
	}
}
