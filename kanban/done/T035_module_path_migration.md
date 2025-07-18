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

**Overall Status:** `Done`

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
- [x] Update `go.mod` module declaration from `davidlee/vice` to `github.com/davidlee/vice`
- [x] Update all import statements across the codebase (111 Go files)
- [x] Verify all imports resolve correctly after migration
- [x] Run `go mod tidy` to clean up module dependencies

### Validation & Testing
- [x] All packages compile successfully after migration
- [x] All existing tests pass with new import paths
- [x] Linting passes without import-related errors
- [x] No broken imports or missing dependencies

### Documentation & Consistency
- [x] Update any documentation referencing the old module path
- [x] Verify README or setup instructions reference correct import paths
- [x] Check for any hardcoded module references in comments or strings

## Implementation Plan & Progress

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

### 1. Pre-Migration Analysis
- [x] **1.1 Audit current import usage**: Identify all files with import statements
  - *Scope:* Find all Go files with `davidlee/vice` imports
  - *Command:* `find . -name "*.go" -exec grep -l "davidlee/vice" {} \;`
  - *Analysis:* Found 111 Go files with import statements needing update
- [x] **1.2 Identify potential risks**: Check for any import-dependent tooling or scripts
  - *Scope:* Build scripts, CI/CD, code generation that might reference module path
  - *Files:* `Justfile`, GitHub Actions, any code generation tools
  - *Result:* No build/CI files found with module path references

### 2. Module Path Migration
- [x] **2.1 Update go.mod**: Change module declaration to GitHub path
  - *Command:* `sed -i 's/module davidlee\/vice/module github.com\/davidlee\/vice/' go.mod`
  - *Verification:* go.mod syntax remains valid
- [x] **2.2 Update all import statements**: Batch replace imports across codebase
  - *Command:* `find . -name "*.go" -exec sed -i 's/davidlee\/vice/github.com\/davidlee\/vice/g' {} \;`
  - *Scope:* All 111 Go files in project updated
  - *Verification:* All imports now use `github.com/davidlee/vice` prefix

### 3. Validation & Testing
- [x] **3.1 Clean module dependencies**: Run go mod tidy
  - *Command:* `go mod tidy`
  - *Purpose:* Clean up any dependency issues from module path change
  - *Result:* Module dependencies cleaned successfully
- [x] **3.2 Compilation verification**: Ensure all packages compile
  - *Command:* `go build ./...`
  - *Purpose:* Verify no broken imports or missing dependencies
  - *Result:* All packages compile successfully
- [x] **3.3 Test suite validation**: Run full test suite
  - *Command:* `go test ./...`
  - *Purpose:* Ensure no test failures from import changes
  - *Result:* All tests pass - 19 packages tested successfully
- [x] **3.4 Linting verification**: Run linting tools
  - *Command:* `golangci-lint run`
  - *Purpose:* Verify no import-related linting errors
  - *Result:* 0 issues found

### 4. Documentation & Cleanup
- [x] **4.1 Update documentation**: Fix any docs referencing old module path
  - *Scope:* README, setup guides, developer documentation
  - *Pattern:* Search for `davidlee/vice` in markdown files
  - *Result:* Task documentation contains historical references (intentionally preserved)
- [x] **4.2 Final verification**: Manual review of critical import changes
  - *Scope:* Main packages, test files, key interfaces
  - *Purpose:* Ensure migration completeness and correctness
  - *Result:* All critical imports verified - `main.go` and core packages use correct paths

## Roadblocks

*(No roadblocks identified yet)*

## Future Improvements & Refactoring Opportunities

### **Post-Migration Optimizations**
1. **Import Organization** - Consider organizing imports by local vs external packages
2. **Module Structure** - Evaluate if current package organization follows Go best practices
3. **Dependency Audit** - Review and potentially update Go dependencies post-migration

## Notes / Discussion Log

### **Task Completion (2025-07-18 - AI)**

**Migration Results:**
- **Files Updated**: 111 Go files + go.mod successfully migrated
- **Validation**: All compilation, tests, and linting pass with 0 issues
- **No Risks**: No build scripts or CI files required updates
- **Documentation**: Task files contain historical references (preserved intentionally)

**Technical Summary:**
- Module path changed from `davidlee/vice` to `github.com/davidlee/vice`
- All internal imports updated: `internal/*`, `cmd/*` packages
- Dependencies cleaned with `go mod tidy`
- Full test suite passes (19 packages)
- Linting clean (0 issues)

**Key Verification Points:**
- `main.go` imports correctly updated
- All test files compile and run successfully
- No import-related compilation errors
- Module resolution works correctly

**Commit:** f6977ba - chore(module)[T035]: migrate module path from davidlee/vice to github.com/davidlee/vice

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