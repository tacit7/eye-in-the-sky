# Token Usage Tab Refactor Action Plan

## 🎯 Objective
Refactor the **Token Usage** view for modularity, maintainability, and performance, while preserving existing visuals and functionality.

---

## Phase 1: File & Architecture Cleanup

### 1. Create a dedicated renderer package
**Goal:** Move all presentation logic out of `app/`.

**Tasks:**
- Create new directory: `internal/ui/view/usage/`
- Move:
  - `usage_renderer.go` → `internal/ui/view/usage/renderer.go`
  - `usage_summary.go` → `internal/ui/view/usage/sections.go`
- Rename package to `usageview`
- Fix imports in `view_usage.go` and `model.go`

**Deliverable:** Renderer logic isolated under `internal/ui/view/usage`.

---

### 2. Consolidate data types
**Goal:** Eliminate duplication between `app.UsageSummary` and `services.UsageSummary`.

**Tasks:**
- Remove redundant structs from `usage_data.go`
- Replace with `services.UsageSummary`
- Delete `convertServiceSummary()` and related code

**Deliverable:** Single canonical `services.UsageSummary`.

---

### 3. Remove deprecated code
**Goal:** Clean technical debt.

**Tasks:**
- Delete deprecated `Build*Summary` functions
- Remove `defaultSvc` and all deprecation comments

**Deliverable:** `usage_data.go` can be deleted after cleanup.

---

## Phase 2: Refactor the Main Rendering Flow

### 4. Simplify `renderUsageTab()`
**Goal:** Improve readability and testability.

**Tasks:**
- Split into three functions:
  ```go
  buildUsageSummary()
  buildUsageSections()
  renderUsageViewport()
  ```
- Move caching logic into `renderUsageViewport()`

**Deliverable:** `renderUsageTab()` reduced to ~15 lines.

---

### 5. Normalize naming conventions
**Goal:** Consistency.

**Tasks:**
- Rename:
  - `renderMonthlyCostsBreakdown()` → `renderMonthlyUsage()`
  - `renderEyeInTheSkyUsage()` → `renderSessionUsage()`
- Follow pattern `renderXUsage()`

**Deliverable:** Uniform naming scheme across renderers.

---

### 6. Parameterize Lipgloss styling
**Goal:** Eliminate hard-coded colors.

**Tasks:**
- Move color codes (`"33"`, `"235"`, etc.) into `Styles`
- Add fields to theme:
  ```go
  GradientHeaderBg
  SectionBorder
  ```
- Populate via `createStyles()`

**Deliverable:** Theme-driven colors and styles.

---

## Phase 3: Add Interaction and Responsiveness

### 7. Add resize and scroll support
**Goal:** Improve UX.

**Tasks:**
- Add `updateUsageTab(msg tea.Msg)` to handle:
  - `tea.WindowSizeMsg`
  - `tea.KeyMsg` (up/down, pageup/pagedown)
- Re-render on width change

**Deliverable:** Scrollable, resizable usage view.

---

### 8. Improve table behavior
**Goal:** Responsive layout.

**Tasks:**
- Update `NewTableBuilder()`:
  - Dynamic column width allocation
  - Optional column hiding for small screens

**Deliverable:** Adaptive tables for 80–140 column terminals.

---

### 9. Compute billing progress accurately
**Goal:** Replace hardcoded `0.62`.

**Tasks:**
- Add `ComputeBlockProgress()` in `UsageService`
- Calculate from block start/end timestamps
- Return percentage to renderer

**Deliverable:** Accurate progress bar.

---

## Phase 4: QA & Testing

### 10. Add targeted tests
**Goal:** Validate refactor.

**Tasks:**
- Add `view/usage/renderer_test.go`
- Test:
  - Header rendering
  - Totals line
  - Borders & spacing

**Deliverable:** 90%+ coverage on service + renderer.

---

### 11. Regression testing
**Goal:** Preserve visuals.

**Tasks:**
- Take baseline screenshots pre-refactor
- Compare post-refactor output

**Deliverable:** Identical visual results, cleaner internals.

---

## 🧩 Timeline Estimate

| Phase | Description | Duration |
|-------|--------------|-----------|
| 1 | File & type cleanup | 1 day |
| 2 | Rendering refactor | 1–2 days |
| 3 | Interaction improvements | 1 day |
| 4 | QA & tests | 0.5 day |

**Total:** ~3.5–4.5 days.

---

## ✅ Final Deliverables

### Completed (Phase 1 - Architecture Refactoring)
- [x] `internal/ui/services/` package created with `UsageService`
- [x] `internal/ui/viewmodel/` package created
- [x] Service layer unit tests (7/7 passing)
- [x] Dirty-flag caching implemented
- [x] Real timestamp display
- [x] Deprecated wrappers for backward compatibility
- [x] App compiles successfully

### Remaining (From Original Plan)
- [ ] `internal/ui/view/usage/` package (move renderers)
- [ ] Simplified `renderUsageTab()` (currently ~116 lines, target ~15)
- [ ] Unified `services.UsageSummary` (remove duplicate app types)
- [ ] Configurable styles (eliminate hardcoded colors)
- [ ] Scroll + resize support
- [ ] Renderer test coverage

### Phase 1 Completion Status
**Status:** ✅ Complete (2025-10-22)

**What was delivered:**
1. Service layer with pure aggregation logic
2. Unit tests covering edge cases (empty inputs, zero budgets, aggregation)
3. Caching infrastructure with dirty flags
4. Real timestamp tracking from service
5. Deprecated backward-compatible wrappers (TODO: remove 2025-11-15)

**Files created:**
- `internal/ui/services/usage_service.go` (282 lines)
- `internal/ui/services/usage_service_test.go` (366 lines)
- `internal/ui/viewmodel/usage_view_model.go` (17 lines)

**Files modified:**
- `internal/ui/app/model.go` - Service injection + caching
- `internal/ui/app/view_usage.go` - Service integration + cache logic
- `internal/ui/app/usage_renderer.go` - Parameter-based rendering
- `internal/ui/app/usage_data.go` - Deprecated wrappers

**Next steps:** Continue with Phase 1 tasks 1-3 from original plan, then proceed to Phase 2.
