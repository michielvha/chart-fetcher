# ChartFetch Improvement Plan

## Overview

This document captures the planned improvements to ChartFetch covering Go project
conventions, code correctness, and security scanning integration.

---

## 1. Bug Fix

**File:** `main.go:91`

On error logging after a failed chart pull, the code logs `err` (from the outer
scope, which may be `nil` or from a previous iteration) instead of `pullErr` (the
actual pull error). This means failed pulls are silently logged with no error detail.

**Fix:** Change `Err(err)` to `Err(pullErr)`.

---

## 2. Go Standard Directory Layout

The original flat layout (`handlers/`, `helpers/` at the repo root alongside `main.go`)
does not follow the [standard Go project layout](https://github.com/golang-standards/project-layout).

### Before

```
ChartFetch/
├── handlers/
│   ├── auth.go
│   ├── chart_management.go
│   ├── helm_handler.go
│   └── repo_management.go
├── helpers/
│   ├── config.go
│   └── env.go
└── main.go
```

### After

```
ChartFetch/
├── cmd/
│   └── chart-fetcher/
│       └── main.go          ← entry point
├── internal/
│   ├── helm/                ← was handlers/
│   │   ├── auth.go
│   │   ├── chart_management.go
│   │   ├── helm_handler.go
│   │   └── repo_management.go
│   └── config/              ← was helpers/
│       ├── config.go
│       └── env.go
└── ...
```

**Rationale:**
- `cmd/` is the idiomatic location for `main` packages in Go projects with multiple packages.
- `internal/` signals these packages are not importable by external consumers.
- Package names become more descriptive: `helm` and `config` instead of `handlers` and `helpers`.

**Required change in `.goreleaser.yml`:** add `main: ./cmd/chart-fetcher` to the build entry.

---

## 3. Code Quality Fixes

### 3a. `RepoNames` map not initialised in constructor

`NewHelmHandler()` returns a `HelmHandler` with a `nil` `RepoNames` map. The map is
lazily initialised inside `AddAndFetchRepo`, but any other caller that tries to read
from `RepoNames` before that would get a nil-map read (safe in Go, but fragile). The
map should be initialised at construction time.

**Fix:** Add `RepoNames: make(map[string]string)` to the struct literal in `NewHelmHandler`.

### 3b. HTTP client has no timeout

Both `chart_management.go` and `repo_management.go` create `&http.Client{}` with no
timeout. A slow or unresponsive registry will hang the process indefinitely.

**Fix:** Use `&http.Client{Timeout: 30 * time.Second}`.

### 3c. `repositories.yaml` written with world-readable permissions

`EnsureRepoFileExists` and `AddAndFetchRepo` write `repositories.yaml` with `0o644`
(world-readable). This file can contain plaintext credentials.

**Fix:** Change permission to `0o600`.

---

## 4. Security Scanning in CI

Two scanners added to `.github/workflows/build-release.yml` as steps that run before
the build:

### govulncheck

[govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) is Google's
official Go vulnerability checker. It cross-references the module graph against the
Go vulnerability database and only reports vulnerabilities that are actually reachable
in the call graph, reducing noise from transitive-only exposure.

```yaml
- name: Run govulncheck
  uses: golang/govulncheck-action@v1
  with:
    go-version-input: '1.25.2'
    go-package: './...'
```

### Trivy filesystem scan

[Trivy](https://trivy.dev) scans the repository for:
- Go module CVEs (via `go.sum`)
- Kubernetes manifest misconfigurations (in `examples/manifests/`)
- Exposed secrets

```yaml
- name: Run Trivy vulnerability scanner
  uses: aquasecurity/trivy-action@0.30.0
  with:
    scan-type: 'fs'
    scan-ref: '.'
    format: 'table'
    exit-code: '1'
    severity: 'CRITICAL,HIGH'
    scanners: 'vuln,secret,misconfig'
```

---

## 5. Unit Tests

No tests existed in the original codebase. Basic tests added:

| File | What is tested |
|---|---|
| `internal/config/config_test.go` | `LoadConfig` with YAML, JSON, unsupported format, and missing file |
| `internal/helm/repo_management_test.go` | `EnsureRepoFileExists` creates file; idempotent on second call |

---

## Tracking

| # | Item | Status |
|---|---|---|
| 1 | Bug fix (`Err(pullErr)`) | Done |
| 2 | Go standard layout | Done |
| 3a | `RepoNames` init in constructor | Done |
| 3b | HTTP client timeout | Done |
| 3c | `repositories.yaml` permissions | Done |
| 4 | govulncheck + Trivy in CI | Done |
| 5 | Unit tests | Done |
