# Build-only braces audit exception

Date: 2026-10-06

The project owner approved a narrow audit exception for
[GHSA-vfj7-8cjw-p6xm](https://github.com/advisories/GHSA-vfj7-8cjw-p6xm)
(CVE-2026-93687), which affects `braces <=3.0.3`. No patched upstream version
exists at the time of this decision.

## Exposure and rationale

The vulnerability requires attacker-controlled deeply nested brace patterns.
`braces` is a development-only transitive dependency of `micromatch` and
`fast-glob`, consumed by the i18n build plugin and ESLint configuration. This
repository supplies fixed, locally defined glob patterns. User-uploaded file
names and file content are not used as build/lint glob expressions.

The production Docker image contains the Go binary and embedded frontend
assets, not Node.js, frontend node_modules, ESLint or the build plugin. The
vulnerable library is not a server runtime dependency.

## Controls and expiry

- Ignore only this GHSA in `pnpm.auditConfig.ignoreGhsas`; keep the high-severity
  audit gate and all other security checks enabled.
- Do not pass untrusted glob patterns to build/lint tooling.
- Remove the exception and upgrade as soon as an upstream patched version is
  available, or reassess if the dependency becomes runtime-reachable.
- Other high-severity findings are addressed by upgrading Vue and overriding
  affected undici, brace-expansion and source-map-js versions.
