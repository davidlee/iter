---
title: "Add Code Quality Analysis Tools"
type: ["chore"]
tags: ["tooling", "quality", "analysis"]
related_tasks: []
context_windows: ["Justfile", "CLAUDE.md", "doc/architecture.md", ".golangci.yml", "go.mod"]
---

# Add Code Quality Analysis Tools

**Context (Background)**:
The project currently has basic linting via golangci-lint but lacks comprehensive code quality analysis tools. Adding specialized analysis tools will help identify architectural debt, measure complexity, and detect potential issues early in development.

**Context (Significant Code Files)**:
- `Justfile` - Contains development commands including current `lint` target
- `.golangci.yml` - Current linter configuration (if exists)
- `go.mod` - Go module dependencies
- Codebase structure across `internal/`, `cmd/`, etc.

## Git Commit History

**All commits related to this task (newest first):**

*No commits yet - task is in backlog*

## 1. Habit / User Story

As a developer, I want comprehensive code quality analysis tools integrated into the development workflow so that I can identify architectural issues, complexity hotspots, and potential maintainability problems before they become technical debt.

## 2. Acceptance Criteria

- [ ] go-architect tool integrated for architectural analysis
- [ ] spm-go tool integrated for software metrics
- [ ] effrit tool integrated for efficiency analysis
- [ ] All tools executable via Justfile commands
- [ ] Tools configured with appropriate thresholds/rules for this codebase
- [ ] Documentation on how to interpret and act on tool outputs

## 3. Architecture

*To be completed during planning phase*

## 4. Implementation Plan & Progress

**Overall Status:** `Not Started`

**Sub-tasks:**
*(To be completed during planning phase)*

## 5. Roadblocks

*(No roadblocks identified yet)*

## 6. Notes / Discussion Log

- `2025-07-15 - User:` Requested integration of three specific tools:
  1. https://github.com/go-architect/go-architect - architectural analysis
  2. https://github.com/fdaines/spm-go - software metrics
  3. https://github.com/Skarlso/effrit - efficiency analysis
- `2025-07-15 - AI:` Created task card based on user request