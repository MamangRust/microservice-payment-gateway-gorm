#!/bin/bash
# run_k6_suite.sh - runs a list of k6 scripts, exports JSON summaries, prints per-file results
# Usage: ./run_k6_suite.sh <service_group_name> <file1.js> [file2.js ...]
# Output: /tmp/k6_results/<group>/<file>.json + compact summary to stdout

set -u

GROUP="$1"
shift

RESULTS_DIR="/tmp/k6_results/$GROUP"
mkdir -p "$RESULTS_DIR"

PASS=0
FAIL=0
FAILED_FILES=()

for f in "$@"; do
  base=$(basename "$f" .js)
  echo "=== RUN $f ==="
  if timeout 2400 k6 run --summary-export "$RESULTS_DIR/$base.json" "$f" > /tmp/k6_results/${GROUP}_${base}.out 2>&1; then
    # extract key metrics
    checks=$(python3 -c "
import json,sys
try:
    d=json.load(open('$RESULTS_DIR/$base.json'))
    m=d.get('metrics',{})
    def v(name,key): return m.get(name,{}).get('values',{}).get(key,0)
    checks_passed=v('checks','passes')
    checks_failed=v('checks','fails')
    failed=v('http_req_failed','rate')
    p95=v('http_req_duration','p(95)')
    rps=v('http_reqs','rate')
    print(f'checks_passed={checks_passed} checks_failed={checks_failed} failed_rate={failed:.3f} p95_ms={p95*1000:.1f} rps={rps:.1f}')
except Exception as e:
    print('PARSE_ERR', e)
")
    echo "  ✓ $f :: $checks"
    PASS=$((PASS+1))
  else
    echo "  ✗ $f FAILED (exit != 0)"
    FAIL=$((FAIL+1))
    FAILED_FILES+=("$f")
    tail -5 /tmp/k6_results/${GROUP}_${base}.out
  fi
done

echo
echo "===== $GROUP SUMMARY ====="
echo "PASS: $PASS  FAIL: $FAIL"
if [[ ${#FAILED_FILES[@]} -gt 0 ]]; then
  printf 'FAILED: %s\n' "${FAILED_FILES[@]}"
fi
exit 0
