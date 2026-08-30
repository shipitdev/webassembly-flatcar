package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	targetURL := flag.String("url", "http://127.0.0.1:8080/api/orders", "Target endpoint URL")
	totalReqs := flag.Int("n", 10000, "Total number of HTTP requests")
	concurrency := flag.Int("c", 100, "Concurrent worker pool size")
	method := flag.String("m", "POST", "HTTP Method (GET or POST)")
	payload := flag.String("d", `{"order_id":"ord-bench","customer_id":"cust-1","items":[{"sku":"item-a","price":49.99,"qty":2}]}`, "Request body payload")
	flag.Parse()

	fmt.Println("======================================================================")
	fmt.Println(" 🚀 FLATCAR HIGH-CONCURRENCY LOAD BLASTER")
	fmt.Printf(" Target: %s [%s] | Concurrency: %d | Total Reqs: %d\n", *targetURL, *method, *concurrency, *totalReqs)
	fmt.Println("======================================================================")

	// Tune HTTP transport for high-throughput connection pooling & zero keep-alive stalls
	transport := &http.Transport{
		MaxIdleConns:        *concurrency * 2,
		MaxIdleConnsPerHost: *concurrency * 2,
		MaxConnsPerHost:     *concurrency * 2,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  true,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	queue := make(chan int, *totalReqs)
	for i := 0; i < *totalReqs; i++ {
		queue <- i
	}
	close(queue)

	var (
		wg            sync.WaitGroup
		successCount  uint64
		errorCount    uint64
		latenciesLock sync.Mutex
		latencies     = make([]time.Duration, 0, *totalReqs)
	)

	start := time.Now()

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localBuf := make([]time.Duration, 0, *totalReqs / *concurrency + 10)

			for range queue {
				reqStart := time.Now()
				var req *http.Request
				var err error

				if *method == "POST" {
					req, err = http.NewRequest("POST", *targetURL, bytes.NewBufferString(*payload))
					req.Header.Set("Content-Type", "application/json")
				} else {
					req, err = http.NewRequest("GET", *targetURL, nil)
				}

				if err != nil {
					atomic.AddUint64(&errorCount, 1)
					continue
				}

				resp, err := client.Do(req)
				if err != nil {
					atomic.AddUint64(&errorCount, 1)
					continue
				}

				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					atomic.AddUint64(&successCount, 1)
					localBuf = append(localBuf, time.Since(reqStart))
				} else {
					atomic.AddUint64(&errorCount, 1)
				}
			}

			latenciesLock.Lock()
			latencies = append(latencies, localBuf...)
			latenciesLock.Unlock()
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	// Compute sorted percentiles
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	rps := float64(successCount) / duration.Seconds()

	fmt.Println("\n📊 BENCHMARK METRICS:")
	fmt.Printf("  • Total Duration:     %.3f seconds\n", duration.Seconds())
	fmt.Printf("  • Successful Reqs:    %d (%.2f%%)\n", successCount, float64(successCount)/float64(*totalReqs)*100)
	fmt.Printf("  • Errors / Drops:     %d (%.2f%%)\n", errorCount, float64(errorCount)/float64(*totalReqs)*100)
	fmt.Printf("  • Throughput (RPS):   \033[1;32m%.1f reqs/sec\033[0m\n", rps)

	if len(latencies) > 0 {
		fmt.Println("\n⏱️  LATENCY PERCENTILES:")
		fmt.Printf("  • Min:    %v\n", latencies[0])
		fmt.Printf("  • p50:    \033[1;36m%v\033[0m (Median)\n", latencies[len(latencies)*50/100])
		fmt.Printf("  • p90:    %v\n", latencies[len(latencies)*90/100])
		fmt.Printf("  • p95:    %v\n", latencies[len(latencies)*95/100])
		fmt.Printf("  • p99:    \033[1;33m%v\033[0m (Tail Latency)\n", latencies[len(latencies)*99/100])
		fmt.Printf("  • Max:    %v\n", latencies[len(latencies)-1])
	}
	fmt.Println("======================================================================")
}
