---
title: "Implement file paths & runtime environment system"
tags: ["config", "runtime", "xdg"]
related_tasks: ["T026_flotsam_note_system"]
context_windows: ["internal/config/**/*.go", "cmd/*.go", "doc/specifications/file_paths_runtime_env.md", "CLAUDE.md"]
---
# Implement file paths & runtime environment system

**Context (Background)**:
Current configuration system uses basic XDG compliance for config directory only, with YAML-based configuration. The specification requires full XDG Base Directory compliance, TOML configuration, and context switching capabilities for compartmentalized data management.

**Type**: `feature`

**Overall Status:** `In Progress`

## Reference (Relevant Files / URLs)

### Significant Code (Files / Functions)
- `internal/config/paths.go`: Current XDG config path handling (`Paths` struct, `GetDefaultPaths()`)
- `cmd/root.go`: CLI configuration (`--config-dir` flag, `initializePaths()`)
- `internal/parser/habits.go`: YAML parsing infrastructure 
- `internal/storage/entries.go`: File storage with YAML
- `internal/init/files.go`: File initialization system
- `internal/debug/logger.go`: Debug logging system

### Relevant Documentation
- `doc/specifications/file_paths_runtime_env.md`: Complete specification
- XDG Base Directory Specification: https://specifications.freedesktop.org/basedir-spec/latest/
- TOML specification: https://toml.io/en/
- Go TOML library: https://github.com/pelletier/go-toml 

### Related Tasks / History
- T026: Flotsam note system (may benefit from context switching)

## Habit / User Story

**As a user**, I want vice to support multiple contexts (personal/work) with TOML configuration and proper XDG directory compliance so that I can:
- Keep personal and work habits completely separate
- Use standard Unix configuration conventions
- Configure keybindings, themes, and emoji mappings via config.toml
- Switch contexts at runtime or via environment variables

This enables professional users to maintain clear boundaries between different aspects of their habit tracking while following Unix filesystem conventions.

## Acceptance Criteria (ACs)

- [ ] TOML configuration support via config.toml in $VICE_CONFIG
- [ ] Full XDG Base Directory compliance (CONFIG, DATA, STATE, CACHE)
- [ ] Context system with configurable contexts array in config.toml
- [ ] Environment variable overrides (VICE_CONFIG, VICE_DATA, etc.)
- [ ] Active context persistence in $VICE_STATE/vice.yml  
- [ ] Runtime context switching capability
- [ ] Context-specific data directories auto-created
- [ ] Backward compatibility with existing YAML configuration
- [ ] Command line flag override support (--config-dir)
- [ ] ViceEnv struct available throughout application

## Architecture

*To be completed during planning phase*

## Implementation Plan & Progress

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

- [ ] **Phase 1: Core ViceEnv Structure & XDG Compliance**
  - [ ] **1.1: Create ViceEnv struct with full XDG support**
    - *Design:* Extend current `config.Paths` to `ViceEnv` struct with XDG_DATA_HOME, XDG_STATE_HOME, XDG_CACHE_HOME support
    - *Code/Artifacts:* `internal/config/env.go` (new), update `internal/config/paths.go`
    - *Testing Strategy:* Unit tests for env variable resolution, XDG path construction
    - *AI Notes:* Priority override: ENV vars → CLI flags → XDG defaults
  - [ ] **1.2: Add pelletier/go-toml dependency and TOML parsing**
    - *Design:* Add go-toml to go.mod, create TOML config parser with contexts support
    - *Code/Artifacts:* `internal/config/toml.go` (new), update go.mod
    - *Testing Strategy:* Unit tests for TOML parsing, invalid config handling
    - *AI Notes:* Start with minimal [core] contexts = ["personal", "work"] support

- [ ] **Phase 2: Context Management System**
  - [ ] **2.1: Implement context switching logic**
    - *Design:* Context manager with active context tracking, state persistence
    - *Code/Artifacts:* `internal/config/context.go` (new), state file management
    - *Testing Strategy:* Unit tests for context switching, state persistence
    - *AI Notes:* Must handle missing contexts gracefully, create data dirs on demand
  - [ ] **2.2: Add context-aware data directory management**
    - *Design:* Extend ViceEnv with context_data path, auto-create context directories
    - *Code/Artifacts:* Update `ViceEnv` struct, directory creation logic
    - *Testing Strategy:* Integration tests for context directory creation
    - *AI Notes:* Create minimal YAML files in new context directories

- [ ] **Phase 3: CLI Integration & Backward Compatibility**
  - [ ] **3.1: Update CLI to use ViceEnv throughout**
    - *Design:* Replace config.Paths usage with ViceEnv in cmd/root.go and subcommands
    - *Code/Artifacts:* `cmd/root.go`, all subcommand files
    - *Testing Strategy:* CLI integration tests, existing command compatibility
    - *AI Notes:* Maintain --config-dir flag behavior, add environment variable support
  - [ ] **3.2: Add backward compatibility layer**
    - *Design:* Auto-migrate existing YAML configs, graceful degradation for missing TOML
    - *Code/Artifacts:* `internal/config/migration.go` (new), update initialization
    - *Testing Strategy:* Migration tests with existing config directories
    - *AI Notes:* Don't break existing users, warn about missing config.toml

- [ ] **Phase 4: Runtime Context Operations**
  - [ ] **4.1: Add context switching commands**
    - *Design:* Add `vice context` subcommand with list/switch operations
    - *Code/Artifacts:* `cmd/context.go` (new), context CLI interface
    - *Testing Strategy:* Manual testing with multiple contexts, state persistence
    - *AI Notes:* Consider adding context indicator to existing commands
  - [ ] **4.2: Add environment variable override support**
    - *Design:* VICE_CONTEXT override, runtime context switching capability
    - *Code/Artifacts:* Update ViceEnv initialization, context resolution
    - *Testing Strategy:* Environment variable integration tests
    - *AI Notes:* ENV var should override persisted state but not modify it

## Roadblocks

*None currently identified*

## Notes / Discussion Log

- `2025-07-16 - AI:` Created task based on doc/specifications/file_paths_runtime_env.md
- `2025-07-16 - AI:` Current system has basic XDG config support but lacks TOML, contexts, and full XDG compliance
- `2025-07-16 - AI:` Moved to in-progress and developed implementation plan. Key questions pending:
  1. Migration strategy for existing YAML configs
  2. Context switching scope (immediate vs lazy reload)
  3. Default context behavior 
  4. CLI integration for context switching

## Git Commit History

*No commits yet - task is in backlog*