#!/bin/bash
# PostToolUse hook: format edited Go files. Reads the hook JSON from stdin (no jq needed).
f=$(sed -n 's/.*"file_path"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)
# Windows: the JSON path uses (escaped) backslashes; normalize to single forward slashes.
f=$(printf '%s' "$f" | sed -e 's#\\\\#/#g' -e 's#\\#/#g' -e 's#//*#/#g')
case "$f" in
  *.go) [ -f "$f" ] && gofmt -w "$f" 2>/dev/null ;;
esac
exit 0
