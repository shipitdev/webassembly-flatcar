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

## TODOs & Open Questions for Next Iteration:
- [ ] Build a baseline container app with identical endpoints so we can actually benchmark latency and RSS side-by-side.
- [ ] Figure out how Wasmtime handles `SIGTERM` / graceful stop when systemd shuts down the service.
- [ ] Write a measurement script to capture cold-start time to first HTTP 200.
