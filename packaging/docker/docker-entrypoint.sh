#!/bin/bash
set -euo pipefail

display_number="${DISPLAY#:}"
display_number="${display_number%%.*}"
display_socket="/tmp/.X11-unix/X${display_number}"

Xvfb "$DISPLAY" -screen 0 1920x1080x24 -nolisten tcp &

for _ in $(seq 1 50); do
  if [[ -S "$display_socket" ]]; then
    exec /usr/local/bin/surf serve
  fi
  sleep 0.1
done

echo "Xvfb did not become ready on $DISPLAY" >&2
exit 1
