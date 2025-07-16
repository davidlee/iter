# Specification: File paths & runtime environment

**Status**: Implemented ✅  
**Dependencies**: pelletier/go-toml library  
**Implementation**: [T028 File Paths & Runtime Environment](../kanban/in-progress/T028_file_paths_runtime_env.md)

## Overview

Vice supports context-switched habit tracking through TOML configuration and XDG Base Directory compliance. Each context maintains completely isolated data directories, enabling users to separate personal/work habits while following Unix filesystem conventions.

## Configuration System

### config.toml Structure
Located in `$VICE_CONFIG/config.toml`, defines application settings:

```toml
[core]
contexts = ["personal", "work"]  # defines available contexts
```

**Auto-creation**: If missing, config.toml is created with default contexts.  
**Future settings**: keybindings, themes, emoji mappings will be added here.

### State Persistence  
Active context stored in `$VICE_STATE/vice.yml`:

```yaml
version: "1.0"
active_context: "personal"  # last context set via 'vice context switch'
```

## XDG Base Directory Compliance

Full implementation of XDG specification with all four directory types:

```bash
$VICE_CONFIG = ($XDG_CONFIG_HOME || $HOME/.config)/vice     # ~/.config/vice
$VICE_DATA   = ($XDG_DATA_HOME || $HOME/.local/share)/vice  # ~/.local/share/vice  
$VICE_STATE  = ($XDG_STATE_HOME || $HOME/.local/state)/vice # ~/.local/state/vice
$VICE_CACHE  = ($XDG_CACHE_HOME || $HOME/.cache)/vice       # ~/.cache/vice
```

**Directory Creation**: All paths expanded to absolute paths, created recursively (0750 permissions) if missing.  
**File Security**: Configuration files use 0600 permissions (user read/write only).

## Context Management

### Data Isolation
Each context maintains separate data in `$VICE_DATA/{context}/`:

```
$VICE_DATA/
├── personal/
│   ├── habits.yml           # habit definitions
│   ├── entries.yml          # daily completion data
│   ├── checklists.yml       # checklist templates  
│   └── checklist_entries.yml # checklist completions
└── work/
    ├── habits.yml
    ├── entries.yml
    ├── checklists.yml
    └── checklist_entries.yml
```

**Auto-initialization**: Files created automatically with sample data (4 habits: 2 simple, 2 elastic).

### Context Switching Methods

Three methods for context switching with clear precedence:

1. **Persistent** (`vice context switch <name>`):
   - Validates context exists in config.toml
   - Updates `$VICE_STATE/vice.yml`  
   - Persists for future sessions

2. **CLI Flag** (`--context <name>`):
   - Transient override per command
   - Higher precedence than persisted state
   - Does NOT modify state file

3. **Environment Variable** (`VICE_CONTEXT=<name>`):
   - Transient override via environment
   - Lower precedence than CLI flag
   - Does NOT modify state file

**Priority Resolution**: CLI flag → ENV var → persisted state → first context in config.toml → "personal" fallback

## Runtime Environment

### ViceEnv Struct
Central configuration available throughout application:

```go
type ViceEnv struct {
    Config      string  // $VICE_CONFIG  
    Data        string  // $VICE_DATA
    State       string  // $VICE_STATE
    Cache       string  // $VICE_CACHE
    Context     string  // active context name
    ContextData string  // $VICE_DATA/{context}
}
```

**Runtime Updates**: Context changes trigger ContextData recomputation and data reload.

### CLI Integration
Global flags available to all commands:

```bash
--config-dir <path>  # override $VICE_CONFIG
--data-dir <path>    # override $VICE_DATA  
--state-dir <path>   # override $VICE_STATE
--cache-dir <path>   # override $VICE_CACHE
--context <name>     # transient context override
```

**Override Priority**: CLI flags → Environment variables → XDG defaults

## Architecture

### Repository Pattern
**Decision**: Repository Pattern chosen over lazy loading for simplicity and clear migration path.

```go
type DataRepository interface {
    LoadHabits(ctx string) (*models.Schema, error)
    LoadEntries(ctx string, date time.Time) (*models.EntryLog, error)
    SaveEntries(ctx string, entries *models.EntryLog) error
    LoadChecklists(ctx string) (*models.ChecklistSchema, error)
    SwitchContext(newContext string) error
}
```

**Key Benefits**:
- Clean abstraction isolates UI from data access details
- "Turn off and on again" context switching avoids race conditions
- Clear migration path to sophisticated caching/lazy loading
- Interface stability supports implementation evolution

### Data Loading Strategy

**"Turn Off and On Again" Approach**:
- Complete data unload on context switch
- Lazy loading: data loaded on-demand when UI requests it
- Race condition avoidance through simplicity
- Memory efficiency through context isolation

**FileInitializer Integration**:
- Automatic file creation via `EnsureContextFiles(env *ViceEnv)`
- Repository methods ensure files exist before data operations
- Consistent sample data across all contexts

### BubbleTea Integration

**UI State Management**:
- Repository abstracts data access from UI components
- Context switches trigger complete UI state refresh
- Modal lifecycle remains unaware of context boundaries
- EntryCollector pattern preserved for UI state coordination

**Message Flow**:
- Context switching handled at application level
- UI components receive refreshed data transparently  
- Deferred state synchronization works unchanged
- Error handling consistent via repository.Error pattern

## Migration Path

### Evolution Strategy
Repository interface enables progressive sophistication:

1. **Phase 1**: SimpleFileRepository (current)
   - Complete reload on context switch
   - No caching, every call hits disk
   - Minimal complexity, no race conditions

2. **Phase 2**: CachedFileRepository (future)
   - In-memory cache with TTL
   - Context-aware cache invalidation
   - Performance optimization for repeated access

3. **Phase 3**: LazyLoadingRepository (advanced)
   - Dependency-aware loading
   - Partial loading for large datasets
   - Background preloading capabilities

**Migration Benefits**:
- Zero UI changes across all phases
- Testability: each phase thoroughly testable in isolation
- Rollback safety: can revert to simpler implementation
- Incremental complexity: add sophistication only when needed

## Security Considerations

**File Permissions**:
- Config files (config.toml, vice.yml): 0600 (user read/write only)
- Data directories: 0750 (user read/write/execute, group read/execute)
- Controlled file paths: all operations use validated, controlled paths

**Error Handling**:
- Clear error messages for permission issues
- Graceful degradation for context switch failures
- Validation of context names and directory creation

## Operational Behavior

### Edge Cases
- **Missing config.toml**: Auto-created with default contexts ["personal", "work"]
- **Invalid context in vice.yml**: Falls back to first context in config.toml
- **Context not in config.toml**: `vice context switch` validates and rejects
- **Empty contexts array**: Falls back to "personal" default
- **Directory creation failures**: Clear error messages with paths and permissions

### Performance
- **Minimal overhead**: Directory flag parsing only affects startup
- **Efficient context switching**: No data migration, just path recomputation  
- **Lazy file operations**: Directories created only when needed
- **Memory usage**: No eager loading across contexts

### Compatibility
- **Legacy support**: `config.Paths` compatibility maintained during transition
- **TodoDashboard migration**: Bridge functions for backward compatibility
- **No breaking changes**: All existing functionality preserved
- **Gradual adoption**: New features adoptable incrementally
