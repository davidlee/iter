---
title: "External Editor and Fuzzy Finder for Flotsam Notes"
tags: ["flotsam", "editor", "ui", "fuzzy-finder", "external-tools"]
related_tasks: ["part-of:T026"]
context_windows: ["internal/flotsam/*", "internal/repository/*", "zk/internal/adapter/editor/*", "zk/internal/adapter/fzf/*"]
---
# External Editor and Fuzzy Finder for Flotsam Notes

**Context (Background)**:
As part of T026 flotsam knowledge management, implement external editor integration and fuzzy finder for editing flotsam notes. Users should be able to launch their preferred editor (vim, neovim, emacs, etc.) to edit flotsam notes, with fuzzy finder for note selection.

**Type**: `feature`

**Overall Status:** `Backlog`

## Reference (Relevant Files / URLs)

### Significant Code (Files / Functions)
**ZK Reference Implementation:**
- `zk/internal/adapter/editor/editor.go` - External editor integration patterns
- `zk/internal/adapter/fzf/fzf.go` - Fuzzy finder implementation
- `zk/internal/cli/cmd/edit.go` - Edit command with fuzzy finder integration

**Vice Integration Points:**
- `internal/repository/interface.go` - Repository interface for flotsam queries
- `internal/repository/file_repository.go` - File-based flotsam operations
- `internal/flotsam/` - Flotsam core functionality

### Related Tasks / History
- Part of T026 flotsam knowledge management system
- Builds on T027 flotsam data layer implementation
- Requires external dependencies: fzf for fuzzy finding

## Habit / User Story

As a knowledge worker using flotsam notes, I want to:
- Launch my preferred external editor (vim/neovim/emacs) to edit flotsam notes
- Use fuzzy finder to quickly select notes for editing
- Edit multiple notes simultaneously when needed
- Have the editor respect my environment settings (EDITOR, VISUAL)
- Preview note content during selection

## Acceptance Criteria (ACs)

### External Editor Integration
- [ ] Detect editor from environment variables: `ZK_EDITOR` → `VISUAL` → `EDITOR`
- [ ] Launch external editor with proper stdio handling
- [ ] Support editing single or multiple flotsam notes
- [ ] Handle editor exit codes and launch failures gracefully
- [ ] Restore `/dev/tty` as stdin for vim compatibility

### Fuzzy Finder Integration  
- [ ] Integrate fzf for interactive note selection
- [ ] Display note titles, paths, and metadata in fuzzy finder
- [ ] Support preview of note content during selection
- [ ] Handle fuzzy finder cancellation gracefully
- [ ] Custom key bindings for enhanced workflow

### Safety & UX Features
- [ ] Confirmation prompt when editing many notes (>5)
- [ ] Force flag to bypass confirmation prompts
- [ ] Graceful error handling for missing dependencies
- [ ] Clear error messages for setup issues

### CLI Integration
- [ ] Add `flotsam edit` command with filtering options
- [ ] Support standard filtering: tags, date ranges, content matching
- [ ] Interactive mode with fuzzy finder
- [ ] Non-interactive mode for scripting

## Implementation Plan & Progress

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

### 1. External Editor Foundation
- [ ] **1.1 Editor detection**: Implement environment variable fallback chain
  - *Pattern:* Follow zk's `NewEditor()` approach with priority fallback
  - *Env vars:* `ZK_EDITOR` → `VISUAL` → `EDITOR` 
  - *Error handling:* Clear message if no editor configured
- [ ] **1.2 Editor launcher**: Create subprocess management for editor execution
  - *Pattern:* Follow zk's `Open()` method with proper stdio handling
  - *Dependencies:* `github.com/kballard/go-shellquote` for safe path handling
  - *Compatibility:* `/dev/tty` stdin restoration for vim

### 2. Fuzzy Finder Integration
- [ ] **2.1 fzf wrapper**: Create fzf integration following zk patterns
  - *Dependencies:* Require fzf installation, clear error if missing
  - *Features:* Custom delimiters, key bindings, preview commands
  - *Error handling:* Handle cancellation and exit codes
- [ ] **2.2 Note formatting**: Format flotsam notes for fuzzy finder display
  - *Display:* Note title, path, tags, modification date
  - *Preview:* Note content preview with syntax highlighting
  - *Delimiter:* Use `\x01` separator following zk pattern

### 3. CLI Command Implementation
- [ ] **3.1 flotsam edit command**: Create edit command with filtering
  - *Filtering:* Reuse existing repository query patterns
  - *Options:* Interactive mode, force flag, path filtering
  - *Integration:* Connect editor launcher with fuzzy finder
- [ ] **3.2 Safety features**: Implement confirmation and error handling
  - *Confirmation:* Prompt when editing >5 notes
  - *Validation:* Check editor and fzf availability
  - *Errors:* Graceful handling of subprocess failures

### 4. Testing & Documentation
- [ ] **4.1 Integration tests**: Test editor and fuzzy finder workflows
  - *Mocking:* Mock external processes for testing
  - *Scenarios:* Single note, multiple notes, cancellation
  - *Error cases:* Missing dependencies, invalid editor commands
- [ ] **4.2 Documentation**: Document external editor setup and usage
  - *Setup:* Environment variable configuration
  - *Usage:* Command examples and workflows
  - *Troubleshooting:* Common issues and solutions

## Technical Design Notes

### Architecture Pattern
- **Adapter Pattern**: Follow zk's approach with `internal/adapter/editor/` and `internal/adapter/fzf/`
- **Dependency Injection**: Editor and fzf instances created via container
- **Clean Separation**: Editor logic separate from fuzzy finder logic

### Dependencies
- **External Tools**: fzf (fuzzy finder), user's preferred editor
- **Go Packages**: `github.com/kballard/go-shellquote` for safe command construction
- **Integration**: Repository interface for flotsam queries

### Error Handling Strategy
- **Graceful Degradation**: Non-interactive mode when fzf unavailable
- **Clear Messages**: Specific error messages for setup issues
- **Exit Codes**: Proper handling of editor and fzf exit codes

## Roadblocks

*(No roadblocks identified yet)*

## Future Improvements & Refactoring Opportunities

### **Enhanced Features**
1. **Editor Plugins** - Integration with editor-specific plugins (vim-zk, etc.)
2. **Preview Customization** - User-configurable preview commands
3. **Multi-Select Operations** - Bulk operations beyond editing
4. **Search Integration** - Full-text search in fuzzy finder

### **Performance Optimizations**
1. **Lazy Loading** - Load note content only when needed for preview
2. **Caching** - Cache fuzzy finder results for repeated operations
3. **Incremental Updates** - Update fuzzy finder without full reload

## Notes / Discussion Log

### **Task Creation (2025-07-18 - AI)**

**Extracted from T026 Analysis:**
- External editor integration essential for flotsam note editing workflow
- Fuzzy finder required for efficient note selection in large collections
- ZK provides excellent reference implementation patterns to follow

**Technical Approach:**
- **Reference Implementation**: Closely follow zk's proven patterns
- **Adapter Pattern**: Clean separation of external tool integrations
- **Safety First**: Confirmation prompts and error handling
- **Environment Respect**: Standard Unix editor environment variables

**Dependencies Identified:**
- **fzf**: Required for fuzzy finding, clear error if missing
- **shellquote**: Safe command construction for various editors
- **External Editor**: User's preferred editor via environment variables

**Integration Points:**
- **Repository Layer**: Leverage existing flotsam query capabilities
- **CLI Framework**: Extend existing command structure
- **Error Handling**: Consistent with vice's error handling patterns

**Key Benefits:**
- Familiar Unix workflow for note editing
- Efficient note selection for large collections
- Integration with user's existing editor setup
- Safe handling of multiple note editing scenarios