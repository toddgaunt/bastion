#!/usr/bin/env bash
# Serves the app locally. Usage: ./run.sh [port]   (default: first free port from 8000)
set -euo pipefail

cd "$(dirname "$0")"

port="${1:-}"
if [[ -z "$port" ]]; then
  port=$(python3 - <<'EOF'
import socket
for p in range(8000, 8100):
    with socket.socket() as s:
        try:
            s.bind(("127.0.0.1", p))
        except OSError:
            continue
        print(p)
        break
EOF
)
fi

if [[ -z "$port" ]]; then
  echo "No free port found between 8000 and 8099." >&2
  exit 1
fi

echo "Pairings is running at: http://localhost:${port}/"
echo "Press Ctrl+C to stop."
exec python3 -m http.server "$port" --bind 127.0.0.1
