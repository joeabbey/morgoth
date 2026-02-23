#!/usr/bin/env bash
set -euo pipefail

# Placeholder helper script: assumes a `morgoth` binary is on PATH.
for f in examples/*.mor; do
  echo "==> $f"
  morgoth run "$f" || true
  echo
done
