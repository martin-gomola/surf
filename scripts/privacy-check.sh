#!/usr/bin/env bash
set -euo pipefail

base="${1:-}"
head="${2:-HEAD}"

if [ -n "$base" ] && [ "$base" != "0000000000000000000000000000000000000000" ]; then
  range="$base..$head"
else
  range="$head^..$head"
fi

bad_identity=0
while IFS=$'\t' read -r name email; do
  [ "$name" = "Martin Gomola" ] || continue
  case "$email" in
    *@users.noreply.github.com) ;;
    *) printf 'Martin Gomola commits must use a GitHub noreply address.\n' >&2; bad_identity=1 ;;
  esac
done < <(git log --format='%an%x09%ae%n%cn%x09%ce' "$range")
[ "$bad_identity" -eq 0 ] || exit 1

if git ls-tree -r --name-only "$head" | grep -E '(^|/)\.env$|(^|/)(id_rsa|id_ed25519)$|\.(pem|p12|pfx|key)$'; then
  echo 'Tracked environment or credential file found.' >&2
  exit 1
fi

mac_home='/''Users/'
linux_home='/''home/'
windows_home='[A-Za-z]:\\''Users\\'
home_pattern="(${mac_home}[[:alnum:]_.-]+|${linux_home}[[:alnum:]_.-]+|${windows_home}[[:alnum:]_.-]+)"
if git grep -n -I -E "$home_pattern" "$head" -- .; then
  echo 'Tracked absolute home path found.' >&2
  exit 1
fi

echo "Privacy checks passed for $range"
