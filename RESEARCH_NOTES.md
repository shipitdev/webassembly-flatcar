# Research & Notes: Running WASM on Flatcar without Docker

*Date: Aug 2026*  
*Author: Harsh (shipitdev)*

## The Core Idea / Goal
Can we run lightweight backend microservices on Flatcar without pulling in Docker or containerd?
Docker takes ~500MB of RAM and 15-30s to boot on edge devices. If we can use the `wasmtime` sysext already in `sysext-bakery`, we could theoretically run a compiled `.wasm` binary directly under systemd with near-zero memory footprint.

---

## 1. How does Go WASI networking work? (wasip1)
- Go 1.21+ introduced `GOOS=wasip1 GOARCH=wasm`.
- Regular Go `net.Listen("tcp", ":8080")` on WASI requires the runtime (Wasmtime) to pass in a pre-opened socket descriptor (`--tcplisten 0.0.0.0:8080`).
- Tested `GOOS=wasip1 GOARCH=wasm go build -o app.wasm main.go`:
  - Compiled size: ~9.5 MB (without debug symbols via `-ldflags="-s -w"`).
  - Memory when idle: ~10-14 MB RSS in Wasmtime.

---

## 2. Flatcar System Integration (Sysext + Systemd)
- Flatcar's `/usr` is read-only.
- If we drop `wasmtime.raw` into `/etc/extensions/`, `systemd-sysext.service` will merge it into `/usr/bin/wasmtime` on boot.
- Unit dependency order:
  - The service MUST specify `After=systemd-sysext.service` and `Requires=systemd-sysext.service`, otherwise systemd might try to execute `/usr/bin/wasmtime` before the sysext overlay is mounted.

---

## 3. Container Baseline & Live Edge Telemetry
- **Idea:** Instead of just a dummy `/health` endpoint, added a Prometheus-compatible `/metrics` and rich JSON `/telemetry` endpoint.
- **Why this is useful:**
  1. We can automate scraping memory and latency directly in GitHub Actions CI to catch regressions.
  2. It turns the microservice into a real **Lightweight Edge Node Monitor** (exporting heap alloc, GC count, and uptime).
- **Container Baseline Created:**
  - Built an identical Go HTTP server compiled for Linux scratch container (`container/Dockerfile`).
  - Added `container.bu` to deploy via Docker.
- **Initial Memory Observation:**
  - `wasmtime` process RSS: ~12.4 MB (0 background daemons).
  - `dockerd` + `containerd` + `containerd-shim` stack: ~118.5 MB RSS (3 background daemons).

---

## 4. Production Business Logic & Workload Scaling (Phase 1)
- **Real Production Workload Validation:**
  - To ensure benchmarks represent true production workloads (not just static hello-world strings), implemented an e-commerce order processing pipeline (`/api/orders`):
    1. **JSON Deserialization:** Ingests nested order arrays and line items.
    2. **Tax & Financial Math:** Calculates itemized subtotals, 8% tax, and grand totals.
    3. **Cryptographic Signatures:** Computes SHA-256 checksums on transactions to simulate token/fraud verification.
    4. **In-Memory State:** Thread-safe state updates using Go `sync.RWMutex`.
- **Benchmark Observation:**
  - Full business logic pipeline achieved **112,895 orders/sec** with **0.32ms median latency** under 50 concurrent workers.
- **WASI Go Runtime Scheduler Bug Encountered & Fixed:**
  - In Go wasip1, single-threaded scheduler entered deadlock when `net.Accept()` blocked with zero background timers.
  - Resolved by adding a 100ms background heartbeat ticker and updating `getListener()` to safely check `LISTEN_FDS`.
