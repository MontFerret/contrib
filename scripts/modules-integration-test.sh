#!/usr/bin/env bash
set -euo pipefail

modules="$(cd "$(dirname "$0")" && pwd)/modules.sh"
suite_root=$(mktemp -d)
trap 'rm -rf "$suite_root"' EXIT
cd "$suite_root"
mkdir -p bin modules/{first,second,empty} tests/modules/{first,second} tmp
touch modules/{first,second,empty}/go.mod
mkfifo block
export TMPDIR="$suite_root/tmp" PATH="$suite_root/bin:$PATH"

cat > bin/runtime <<'RUNTIME'
#!/usr/bin/env bash
set -eu
[[ "$1" == --listen=127.0.0.1:0 ]]
echo "$$" >> children
case "$scenario" in
  startup) exit 41 ;;
  cleanup|both) trap 'exit 42' TERM ;;
  *) trap 'exit 0' TERM ;;
esac
exec 3<>block
if [[ "$scenario" != timeout ]]; then
  echo 'wire+tcp://127.0.0.1:43210'
fi
read -r -t 30 -u 3 _
RUNTIME

cat > bin/lab <<'LAB'
#!/usr/bin/env bash
set -eu
[[ "$2" == --runtime=wire+tcp://127.0.0.1:43210 ]]
echo "$$" >> children
echo "$1" >> calls
case "$1/$scenario" in
  version/handshake) exit 21 ;;
  run/lab|run/both) exit 22 ;;
  run/interrupt) kill -TERM "$PPID"; exec sleep 30 ;;
esac
LAB
chmod +x bin/{runtime,lab}

check_suite() {
  export scenario="$1"
  local expected="$2" expected_calls="$3" status=0 pid
  shift 3
  : > calls
  : > children
  bash "$modules" test-integration "$@" > log 2>&1 || status=$?
  if [[ "$status" != "$expected" || "$(paste -sd, calls)" != "$expected_calls" ]]; then
    cat log calls >&2
    echo "$scenario: exit $status (expected $expected), expected calls '$expected_calls'" >&2
    exit 1
  fi
  while read -r pid; do
    if kill -0 "$pid" 2>/dev/null; then
      echo "$scenario: process $pid is still running" >&2
      exit 1
    fi
  done < children
  if [[ -n "$(ls -A "$TMPDIR")" ]]; then
    echo "$scenario: temporary endpoint directory was left behind" >&2
    exit 1
  fi
}

check_suite success 0 'version,run,version,run' first empty second
check_suite startup 1 '' first
check_suite timeout 1 '' first
check_suite handshake 21 version first
check_suite lab 22 version,run first
check_suite cleanup 42 version,run first
check_suite both 22 version,run first
grep -q 'Wire runtime exited with status 42' log
check_suite interrupt 143 version,run first
