---
phase: 18
slug: category-display-cli
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-09
---

# Phase 18 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — Go built-in testing |
| **Quick run command** | `go test ./internal/storage ./internal/cli ./internal/handlers -v -short` |
| **Full suite command** | `go test ./... -v` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/storage ./internal/cli ./internal/handlers -v -short`
- **After every plan wave:** Run `go test ./... -v`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 18-01-01 | 01 | 1 | CATG-08 | — | N/A | integration | `go test ./internal/handlers -v -run TestSidebarGrouping` | ❌ W0 | ⬜ pending |
| 18-01-02 | 01 | 1 | CATG-09 | — | N/A | integration | `go test ./internal/handlers -v -run TestUncategorizedSection` | ❌ W0 | ⬜ pending |
| 18-02-01 | 02 | 2 | CATG-10 | — | N/A | unit | `go test ./internal/cli -v -run TestCategoryFlag` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/handlers/sidebar_test.go` — stubs for CATG-08/09 (sidebar grouping tests)
- [ ] `internal/cli/commands/article_test.go` — stubs for CATG-10 (category flag tests)
- [ ] `internal/storage/database_test.go` — extend for GetCategoryByName, ListArticlesByCategory

*Existing infrastructure: Go test framework already in use (Phase 1-15).*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| localStorage persistence (expand/collapse) | CATG-08 | Browser-local state, not server-testable | 1. Open sidebar in browser 2. Click category header to collapse 3. Refresh page 4. Verify collapsed state persists |
| Chevron rotation animation | CATG-08 | CSS animation timing | 1. Click category header 2. Observe chevron rotates -90deg over 250ms |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending