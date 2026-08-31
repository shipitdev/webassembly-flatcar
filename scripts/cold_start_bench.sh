#!/bin/bash
# ==============================================================================
# Cold-Start Instantiation Benchmark
# Measures the raw process start-to-first-request latency of Wasmtime on Flatcar
# ==============================================================================
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WASM_BIN="${ROOT_DIR}/bin/app.wasm"
PORT=8091
SAMPLES=10
TOTAL_MS=0

echo "======================================================================"
echo " ⚡ FLATCAR WASM COLD-START INSTANTIATION BENCHMARK"
echo " Sampling ${SAMPLES} cold process start-to-ready cycles on port ${PORT}"
echo "======================================================================"

for i in $(seq 1 "${SAMPLES}"); do
    start_ns=$(date +%s%N)
    
    # Launch fresh cold WASM instance
    wasmtime run -S inherit-network=y --env "PORT=${PORT}" "${WASM_BIN}" >/dev/null 2>&1 &
    PID=$!
    
    # Poll until ready
    while true; do
        if curl -s -o /dev/null -w "%{http_code}" --connect-timeout 0.05 "http://127.0.0.1:${PORT}/api/ping" 2>/dev/null | grep -q "200"; then
            break
        fi
        sleep 0.001
    done
    end_ns=$(date +%s%N)
    
    kill "${PID}" 2>/dev/null || true
    wait "${PID}" 2>/dev/null || true
    
    sample_ms=$(( (end_ns - start_ns) / 1000000 ))
    TOTAL_MS=$(( TOTAL_MS + sample_ms ))
    echo "  • Sample ${i}/${SAMPLES}: ${sample_ms} ms (Cold Instantiation)"
    sleep 0.2
done

avg_ms=$(( TOTAL_MS / SAMPLES ))
echo ""
echo "======================================================================"
echo " 🏆 AVERAGE COLD START: ${avg_ms} ms"
echo " (Contrast: Docker container cold start is 500ms - 2,500ms)"
echo "======================================================================"
