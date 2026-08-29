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
  - Compiled size: ~9.2 MB (without debug symbols via `-ldflags="-s -w"`).
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

## Next Steps:
- [ ] Build an automated benchmark script (`benchmark/run_benchmark.sh`) to measure cold-start latency and RSS directly in QEMU.
- [ ] Write a clean, comprehensive `README.md` with real benchmark tables.
- [ ] Clean up loose test files for the final release.
