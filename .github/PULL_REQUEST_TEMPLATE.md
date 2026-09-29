## Description

<!-- What changed, why, and which issue it closes. -->

Closes #

## Evidence

<!-- Evidence over narrative: paste the command(s) you ran and the observed
     result. "Tested locally" without a command and output does not count. -->

- Command:
- Result:

## Change Category

<!-- The changelog category decides the release version:
     fix => patch, feat => minor, breaking change => major. -->

- [ ] `fix` (patch)
- [ ] `feat` (minor)
- [ ] breaking change (major)
- [ ] internal only (`chore`, `test`, `ci`, `docs`, refactor without behavior change)

## Checklist

- [ ] PR title follows [Conventional Commits](https://www.conventionalcommits.org/).
- [ ] Regression test added for the bug (or the reason it cannot be tested is written below).
- [ ] No speculative abstraction, unused code, TODO, or drive-by refactor.
- [ ] No secrets, credentials, tokens, personal or infrastructure data in the diff.
- [ ] New dependency (if any) has a one-line justification above.
- [ ] `go test ./...` and `pnpm test` were run locally.
