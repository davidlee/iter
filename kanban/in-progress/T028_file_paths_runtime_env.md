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
    - *Design:* New `ViceEnv` struct with all XDG directories (CONFIG/DATA/STATE/CACHE), context-aware data paths
    - *Code/Artifacts:* `internal/config/env.go` (new), replace `config.Paths` usage
    - *Testing Strategy:* Unit tests for env variable resolution, XDG path construction
    - *AI Notes:* Priority override: ENV vars → CLI flags → XDG defaults
  - [ ] **1.2: Add pelletier/go-toml dependency and TOML parsing**
    - *Design:* Add go-toml to go.mod, parse config.toml from $VICE_CONFIG for app settings
    - *Code/Artifacts:* `internal/config/toml.go` (new), update go.mod
    - *Testing Strategy:* Unit tests for TOML parsing, invalid config handling, defaults
    - *AI Notes:* config.toml = app settings (keybindings, themes, contexts array), not user data

- [ ] **Phase 2: Context Management System**
  - [ ] **2.1: Implement context switching with immediate data unload**
    - *Design:* Context manager unloads all data immediately, loads on-demand per UI needs
    - *Code/Artifacts:* `internal/config/context.go` (new), state file management
    - *Testing Strategy:* Unit tests for context switching, data unloading behavior
    - *AI Notes:* No eager loading - load data only when UI components request it
  - [ ] **2.2: Add context-aware data directory management**
    - *Design:* Auto-create $VICE_DATA/$CONTEXT directories, populate with minimal YAML files
    - *Code/Artifacts:* Update `ViceEnv` struct, directory creation logic
    - *Testing Strategy:* Integration tests for context directory creation
    - *AI Notes:* First context in array is default, create dirs on first access

- [ ] **Phase 3: CLI Integration & Context Flags**
  - [ ] **3.1: Update CLI to use ViceEnv throughout**
    - *Design:* Replace config.Paths usage with ViceEnv in cmd/root.go and subcommands
    - *Code/Artifacts:* `cmd/root.go`, all subcommand files
    - *Testing Strategy:* CLI integration tests, existing command compatibility
    - *AI Notes:* Maintain --config-dir flag behavior, add environment variable support
  - [ ] **3.2: Add transient --context CLI flag**
    - *Design:* Global --context flag for non-interactive operations, no state persistence
    - *Code/Artifacts:* Update root command flags, context resolution logic
    - *Testing Strategy:* CLI tests with --context flag, verify no state persistence
    - *AI Notes:* Transient override for CLI ops, doesn't modify $VICE_STATE/vice.yml

- [ ] **Phase 4: Runtime Context Operations**
  - [ ] **4.1: Add context switching commands**
    - *Design:* Add `vice context` subcommand with list/switch operations for persistence
    - *Code/Artifacts:* `cmd/context.go` (new), context CLI interface
    - *Testing Strategy:* Manual testing with multiple contexts, state persistence
    - *AI Notes:* Interactive context switching persists to state file
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
- `2025-07-16 - User:` Clarified design decisions:
  1. config.toml = app settings ($VICE_CONFIG), YAML = user data ($VICE_DATA) - no migration needed
  2. Context switching: unload all data immediately, load on-demand per UI needs
  3. First context in array is default
  4. --context CLI flag should be transient (no state persistence) for non-interactive ops
- `2025-07-16 - AI:` Updated implementation plan to reflect correct separation of concerns
- `2025-07-16 - User:` Noted migration concern: contexts will change user data paths but user will handle manual migration

## Git Commit History

- `9c87759` - docs(kanban)[T028]: create file paths & runtime environment implementation plan