#!/bin/bash
# PostToolUse hook: after a Go file is edited, report raw SQL outside the allowed layers and
# leftovers of the old query builder API. Exit code 2 sends the message to Claude so it can fix
# the file. This is a best-effort check; the sql-safety-reviewer agent and tests are the real gate.
#
# Only NEW violations block: a pattern blocks when the file has more hits than the same file at
# git HEAD (an untracked file counts 0 at HEAD). So a file that is half-way through the
# migration is not blocked, but adding old API or raw SQL anywhere is.
f=$(sed -n 's/.*"file_path"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)
# Windows: the JSON path uses (escaped) backslashes; normalize to single forward slashes.
f=$(printf '%s' "$f" | sed -e 's#\\\\#/#g' -e 's#\\#/#g' -e 's#//*#/#g')
case "$f" in *.go) ;; *) exit 0 ;; esac
case "$f" in *_test.go) exit 0 ;; esac
[ -f "$f" ] || exit 0

dir=$(dirname "$f")
base=$(basename "$f")
head_file=$(git -C "$dir" show "HEAD:./$base" 2>/dev/null)

# check LABEL PATTERN [-i]: append the matching lines to msgs when the file has more hits than at HEAD.
msgs=""
check() {
  local label=$1 pattern=$2 flags=${3:-}
  local now before
  now=$(grep -cE $flags "$pattern" "$f")
  before=$(printf '%s' "$head_file" | grep -cE $flags "$pattern")
  if [ "$now" -gt "$before" ]; then
    msgs="$msgs
$label ($before at HEAD, $now now):
$(grep -nE $flags "$pattern" "$f")"
  fi
}

# old API: forbidden everywhere except while it is being removed from internal/common
case "$f" in
  */internal/common/*) ;;
  *) check "Old query builder API added (migrate to infrastructure/db/querybuilder)" \
       'common\.(NewQueryBuilder|NewRepositoryQueryBuilder|QueryBuilder|RepositoryConfig|NewUnionQueryBuilder)|AddOrCondition|AddCustomCondition|AddCustomSearchCondition|GetWhereClause|CreateBasicWhere' ;;
esac
# raw SQL outside the allowed layers
case "$f" in
  */internal/repository/*|*/internal/infrastructure/db/*|*/migrations/*) ;;
  *)
    check "Direct database call added outside internal/repository" \
      '\b(DB|db|tx|Tx|database)\.(Get|Select|Query|QueryRow|Exec)(Context|x)?\('
    check "SQL string built outside internal/repository (use querybuilder via the repository)" \
      'Sprintf\(.*(SELECT |INSERT INTO|UPDATE .* SET|DELETE FROM)' -i
    ;;
esac

if [ -n "$msgs" ]; then
  printf 'SQL rules violated in %s:%s\nSee .claude/rules/database.md. Fix this before continuing.\n' "$f" "$msgs" >&2
  exit 2
fi
exit 0
