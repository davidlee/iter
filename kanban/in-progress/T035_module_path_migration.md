---
title: "Module Path Migration to GitHub"
tags: ["maintenance", "module", "github", "imports", "go-mod"]
related_tasks: ["extracted-from:T027"]
context_windows: ["go.mod", "**/*.go"]
---
# Module Path Migration to GitHub

**Context (Background)**:
Migrate the Go module path from local `davidlee/vice` to GitHub-compatible `github.com/davidlee/vice` for proper module distribution and standard Go conventions.

**Type**: `maintenance`

**Overall Status:** `In Progress`

## Reference (Relevant Files / URLs)

### Significant Code (Files / Functions)
**Module Configuration:**
- `go.mod` - Current module declaration needs updating
- All Go files with import statements (106+ files)

**Import Patterns to Update:**
- `davidlee/vice/internal/*` → `github.com/davidlee/vice/internal/*`
- `davidlee/vice/cmd/*` → `github.com/davidlee/vice/cmd/*`
- `davidlee/vice/pkg/*` → `github.com/davidlee/vice/pkg/*`

### Related Tasks / History
- Extracted from T027 Phase 6 (flotsam-specific work complete)
- General codebase maintenance task affecting all Go packages

## Habit / User Story

As a Go developer working with the vice codebase, I need the module path to follow standard GitHub conventions so that:
- The module can be properly published and imported by other projects
- Go tooling works correctly with standard module resolution
- The codebase follows Go community best practices
- Dependencies and imports are resolved correctly

## Acceptance Criteria (ACs)

### Module Path Update
- [ ] Update `go.mod` module declaration from `davidlee/vice` to `github.com/davidlee/vice`
- [ ] Update all import statements across the codebase (106+ Go files)
- [ ] Verify all imports resolve correctly after migration
- [ ] Run `go mod tidy` to clean up module dependencies

### Validation & Testing
- [ ] All packages compile successfully after migration
- [ ] All existing tests pass with new import paths
- [ ] Linting passes without import-related errors
- [ ] No broken imports or missing dependencies

### Documentation & Consistency
- [ ] Update any documentation referencing the old module path
- [ ] Verify README or setup instructions reference correct import paths
- [ ] Check for any hardcoded module references in comments or strings

## Implementation Plan & Progress

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

### 1. Pre-Migration Analysis
- [ ] **1.1 Audit current import usage**: Identify all files with import statements
  - *Scope:* Find all Go files with `davidlee/vice` imports
  - *Command:* `find . -name "*.go" -exec grep -l "davidlee/vice" {} \;`
  - *Analysis:* Document import patterns and frequency
- [ ] **1.2 Identify potential risks**: Check for any import-dependent tooling or scripts
  - *Scope:* Build scripts, CI/CD, code generation that might reference module path
  - *Files:* `Justfile`, GitHub Actions, any code generation tools

### 2. Module Path Migration
- [ ] **2.1 Update go.mod**: Change module declaration to GitHub path
  - *Command:* `sed -i 's/module davidlee\/vice/module github.com\/davidlee\/vice/' go.mod`
  - *Verification:* Ensure go.mod syntax remains valid
- [ ] **2.2 Update all import statements**: Batch replace imports across codebase
  - *Command:* `find . -name "*.go" -exec sed -i 's/davidlee\/vice/github.com\/davidlee\/vice/g' {} \;`
  - *Scope:* All Go files in project
  - *Verification:* Manual spot-check of critical files

### 3. Validation & Testing
- [ ] **3.1 Clean module dependencies**: Run go mod tidy
  - *Command:* `go mod tidy`
  - *Purpose:* Clean up any dependency issues from module path change
- [ ] **3.2 Compilation verification**: Ensure all packages compile
  - *Command:* `go build ./...`
  - *Purpose:* Verify no broken imports or missing dependencies
- [ ] **3.3 Test suite validation**: Run full test suite
  - *Command:* Run existing test commands per project conventions
  - *Purpose:* Ensure no test failures from import changes
- [ ] **3.4 Linting verification**: Run linting tools
  - *Command:* Run existing lint commands per project conventions
  - *Purpose:* Verify no import-related linting errors

### 4. Documentation & Cleanup
- [ ] **4.1 Update documentation**: Fix any docs referencing old module path
  - *Scope:* README, setup guides, developer documentation
  - *Pattern:* Search for `davidlee/vice` in markdown files
- [ ] **4.2 Final verification**: Manual review of critical import changes
  - *Scope:* Main packages, test files, key interfaces
  - *Purpose:* Ensure migration completeness and correctness

## Roadblocks

*(No roadblocks identified yet)*

## Future Improvements & Refactoring Opportunities

### **Post-Migration Optimizations**
1. **Import Organization** - Consider organizing imports by local vs external packages
2. **Module Structure** - Evaluate if current package organization follows Go best practices
3. **Dependency Audit** - Review and potentially update Go dependencies post-migration

## Notes / Discussion Log

### **Task Creation (2025-07-18 - AI)**

**Extraction from T027:**
- Originally part of T027 Phase 6, but recognized as general maintenance task
- Not flotsam-specific - affects entire Go codebase
- Medium effort task touching many files, best done when not actively developing features
- Estimated impact: 106+ Go files need import statement updates

**Migration Strategy:**
- **Automated Approach**: Use find/sed commands for bulk replacement
- **Risk Mitigation**: Comprehensive testing after each major step
- **Validation**: Multiple verification steps (compile, test, lint)
- **Timing**: Suitable for dedicated maintenance session

**Implementation Considerations:**
- **Batch Operations**: Update all files at once to avoid partial state
- **Verification Points**: Compile and test after each major change
- **Rollback Strategy**: Git provides rollback if issues discovered
- **Dependencies**: May affect any tools or scripts referencing module path

**Key Commands Identified:**
```bash
# Module path update
sed -i 's/module davidlee\/vice/module github.com\/davidlee\/vice/' go.mod

# Import statement updates  
find . -name "*.go" -exec sed -i 's/davidlee\/vice/github.com\/davidlee\/vice/g' {} \;

# Cleanup and validation
go mod tidy
go build ./...
```

**Benefits:**
- GitHub compatibility for module distribution
- Standard Go module conventions
- Improved tooling compatibility
- Community best practices alignment