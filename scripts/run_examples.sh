#!/usr/bin/env bash
set -euo pipefail

# Placeholder helper script: assumes a `mordor` binary is on PATH.
for f in examples/*.mor; do
  echo "==> $f"
  mordor run "$f" || true
  echo
done
