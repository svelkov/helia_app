# Working safely

1. **Read first, then plan.** Before editing, read the relevant code and state the plan: files to change, what
   stays the same, what could break. Wait for approval on anything that changes behavior.
2. **Small steps.** One package or one concern per step. After each step run
   `go build ./... && go vet ./... && go test -count=1 -skip TestZzPorTmp ./...`, check the formatting of the
   changed files (`tr -d '\r' < file | gofmt -l`, the files have CRLF line endings) and report the output. Do not
   start the next step while the build or tests are red because of your change.
3. **Baseline.** Before the first edit, record what already fails. Never report pre-existing failures as yours
   and never "fix" them silently.
4. **No blind bulk edits.** Do not run mass search-and-replace. Edit file by file and review each diff.
5. **Stop and ask** when: a behavior changes (what the user sees, what is saved, an error that appears or
   disappears), a domain type or table differs from what you assumed, a test fails after two attempts, an
   instruction conflicts with a rule, or you are unsure. Propose the change and wait.
6. **Forbidden without explicit permission:** editing files in `migrations/`, running anything against a real
   database (read-only checks only when the user asks, see `database.md`), `git commit`, `git push`,
   `git reset --hard`, `rm -rf`, deleting files outside the repo, adding dependencies, disabling or weakening a
   test. Files in `.claude/` are not in git: deleting them can't be undone, so confirm first.
7. **Never** make a test pass by loosening its assertion. If an expectation looks wrong, explain why and ask.
8. **Report honestly.** End each step with: files changed, commands run with results, behavior changes found,
   open questions. Say what you could not verify (e.g. a screen you could not try).
