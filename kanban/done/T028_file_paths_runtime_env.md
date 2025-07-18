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

**Overall Status:** `Done`

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

> user: review and create an ascii diagram of the existing structs / interfaces which handle data loading and persistence for user data. Then consider potential
> architectural improvements (e.g. repository pattern); enumerate any options worth consideration and analysis of tradeoffs. note we likely have to track
> state of which data is loaded somewhere in order to load on demand; where will we manage that state and what's the best approach to information hiding?
> consider whether any investigation into UI layer / patterns (bubbletea) is necessary

### Current Data Loading & Persistence Architecture

```
CURRENT DATA ARCHITECTURE:
┌─────────────────────────────────────────────────────────────────────────┐
│                           CLI Entry Point                               │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                       cmd/entry.go                                  ││
│  │  • runEntryMenu(paths) - orchestrates data loading                  ││
│  │  • LoadFromFile() - habit schema loading                           ││
│  │  • loadTodayEntries() - existing entry loading                     ││
│  └─────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         Data Layer                                      │
│  ┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────────┐│
│  │   Parser Layer      │  │   Storage Layer     │  │   Config Layer      ││
│  │ ┌─────────────────┐ │  │ ┌─────────────────┐ │  │ ┌─────────────────┐ ││
│  │ │  HabitParser    │ │  │ │  EntryStorage   │ │  │ │  Paths          │ ││
│  │ │• LoadFromFile() │ │  │ │• Load/Save()    │ │  │ │• File paths     │ ││
│  │ │• SaveToFile()   │ │  │ │• Atomic writes  │ │  │ │• XDG compliance │ ││
│  │ │• ID persistence │ │  │ │• Backup config  │ │  │ │                 │ ││
│  │ └─────────────────┘ │  │ └─────────────────┘ │  │ └─────────────────┘ ││
│  │ ┌─────────────────┐ │  │                     │  │                     ││
│  │ │ChecklistParser  │ │  │                     │ │                     ││
│  │ │ChecklistEntries │ │  │                     │ │                     ││
│  │ └─────────────────┘ │  │                     │ │                     ││
│  └─────────────────────┘  └─────────────────────┘  └─────────────────────┘│
└─────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      Models & Validation                                │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                      models/ Package                                ││
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ ││
│  │  │   Schema    │  │ EntryLog    │  │ Checklist   │  │ HabitEntry  │ ││
│  │  │• Habits[]   │  │• DayEntry[] │  │• Items[]    │  │• Value      │ ││
│  │  │• Validate() │  │• Version    │  │• Template   │  │• Status     │ ││
│  │  │• Version    │  │             │  │• Metadata   │  │• Timestamps │ ││
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘ ││
│  └─────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         UI State Management                             │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                      EntryCollector                                 ││
│  │  ┌─────────────────────────────────────────────────────────────────┐││
│  │  │               Central State Coordinator                        │││
│  │  │  • habitParser: *parser.HabitParser                           │││
│  │  │  • entryStorage: *storage.EntryStorage                        │││
│  │  │  • habits: []models.Habit                                     │││
│  │  │  • entries: map[string]interface{}                            │││
│  │  │  • achievements: map[string]*models.AchievementLevel         │││
│  │  │  • notes: map[string]string                                   │││
│  │  │  • statuses: map[string]models.EntryStatus                   │││
│  │  └─────────────────────────────────────────────────────────────────┘││
│  └─────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        BubbleTea UI Layer                               │
│  ┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────────┐│
│  │  EntryMenuModel     │  │  EntryFormModal     │  │ Field Input         ││
│  │ ┌─────────────────┐ │  │ ┌─────────────────┐ │  │ Components          ││
│  │ │• habits []      │ │  │ │• habit config   │ │  │ ┌─────────────────┐ ││
│  │ │• entries map    │ │  │ │• input factory  │ │  │ │ BooleanEntry    │ ││
│  │ │• entryCollector │ │  │ │• field inputs   │ │  │ │ TextEntry       │ ││
│  │ │• directModal    │ │  │ │• validation     │ │  │ │ NumericEntry    │ ││
│  │ │• selectedID     │ │  │ └─────────────────┘ │  │ │ ChecklistEntry  │ ││
│  │ └─────────────────┘ │  │                     │  │ └─────────────────┘ ││
│  └─────────────────────┘  └─────────────────────┘  └─────────────────────┘│
│                                       │                                   │
│                              ┌─────────────────┐                         │
│                              │   Data Flow     │                         │
│                              │• Key messages   │                         │
│                              │• State sync     │                         │
│                              │• Modal lifecycle│                         │
│                              │• Deferred cmds  │                         │
│                              └─────────────────┘                         │
└─────────────────────────────────────────────────────────────────────────┘

DATA LOADING PATTERNS:
┌─────────────────────────────────────────────────────────────────────────┐
│                        Current Loading Strategy                         │
│                                                                         │
│  1. CLI Orchestration (cmd/entry.go):                                  │
│     • Load habits.yml → Schema validation                              │
│     • Load entries.yml → Today's entries                               │
│     • Create EntryCollector with loaded data                           │
│                                                                         │
│  2. State Management (EntryCollector):                                 │
│     • In-memory maps for all data types                                │
│     • Interface{} values for type flexibility                          │
│     • Separate maps for different data aspects                         │
│                                                                         │
│  3. File-based Persistence:                                            │
│     • Load-modify-save pattern                                         │
│     • Atomic writes with backups                                       │
│     • YAML serialization with custom time handling                     │
│                                                                         │
│  4. UI Data Binding:                                                   │
│     • Factory pattern for field inputs                                 │
│     • Type-specific validation and conversion                          │
│     • Deferred state synchronization                                   │
└─────────────────────────────────────────────────────────────────────────┘
```

### File Structure & Data Flow

```
DATA FILES:                    COMPONENTS:
├── habits.yml                ├── parser/habits.go          (HabitParser)
├── entries.yml               ├── storage/entries.go        (EntryStorage)
├── checklists.yml            ├── parser/checklists.go      (ChecklistParser)
└── checklist_entries.yml     └── parser/checklist_entries.go

STATE MANAGEMENT:              UI COMPONENTS:
├── EntryCollector             ├── entrymenu/model.go        (EntryMenuModel)
│   ├── entries map[string]interface{}   ├── modal/entry_form_modal.go (EntryFormModal)
│   ├── achievements map       └── entry/field_inputs.go     (Field Components)
│   ├── notes map
│   └── statuses map
```

### Current vs Target Architecture

```
CURRENT ARCHITECTURE:
┌─────────────────────────────────────────────────────────────────────────┐
│                              CLI Layer                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │   cmd/      │  │   cmd/      │  │   cmd/      │  │   cmd/      │    │
│  │  root.go    │  │  entry.go   │  │  habit.go   │  │  todo.go    │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │
│         │                 │                 │                 │         │
│         └─────────────────┼─────────────────┼─────────────────┘         │
│                           │                 │                           │
│                           ▼                 ▼                           │
│                    ┌─────────────────────────────────────────────────┐  │
│                    │           GetPaths()                            │  │
│                    │      returns *config.Paths                     │  │
│                    └─────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Config Layer                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                     config.Paths                                    ││
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ ││
│  │  │ ConfigDir   │  │ HabitsFile  │  │ EntriesFile │  │ChecklistsFile││
│  │  │   (XDG)     │  │   (YAML)    │  │   (YAML)    │  │   (YAML)    │ ││
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘ ││
│  └─────────────────────────────────────────────────────────────────────┘│
│                                │                                         │
│                                ▼                                         │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                XDG Path Resolution                                   ││
│  │          ~/.config/vice (ConfigDir only)                            ││
│  └─────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        UI/Storage Layer                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │  UI/Todo    │  │  UI/Entry   │  │  Storage    │  │  Debug      │    │
│  │   (paths)   │  │   (paths)   │  │   (paths)   │  │   (paths)   │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘

TARGET ARCHITECTURE:
┌─────────────────────────────────────────────────────────────────────────┐
│                              CLI Layer                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │   cmd/      │  │   cmd/      │  │   cmd/      │  │   cmd/      │    │
│  │  root.go    │  │  entry.go   │  │  habit.go   │  │  todo.go    │    │
│  │ + --context │  │             │  │             │  │             │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │
│         │                 │                 │                 │         │
│         └─────────────────┼─────────────────┼─────────────────┘         │
│                           │                 │                           │
│                           ▼                 ▼                           │
│                    ┌─────────────────────────────────────────────────┐  │
│                    │           GetEnv()                              │  │
│                    │      returns *config.ViceEnv                   │  │
│                    └─────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Config Layer                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                     config.ViceEnv                                  ││
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ ││
│  │  │   Config    │  │    Data     │  │    State    │  │    Cache    │ ││
│  │  │(XDG_CONFIG) │  │(XDG_DATA)   │  │(XDG_STATE)  │  │(XDG_CACHE)  │ ││
│  │  │config.toml  │  │/context/    │  │vice.yml     │  │             │ ││
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘ ││
│  └─────────────────────────────────────────────────────────────────────┘│
│                                │                                         │
│                                ▼                                         │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                Context Management                                    ││
│  │  ┌─────────────────────────────────────────────────────────────────┐││
│  │  │  config.toml: [core] contexts = ["personal", "work"]            │││
│  │  │  vice.yml: active_context = "personal"                         │││
│  │  │  ENV: VICE_CONTEXT override                                    │││
│  │  │  CLI: --context transient override                             │││
│  │  └─────────────────────────────────────────────────────────────────┘││
│  └─────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        UI/Storage Layer                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │  UI/Todo    │  │  UI/Entry   │  │  Storage    │  │  Debug      │    │
│  │   (env)     │  │   (env)     │  │   (env)     │  │   (env)     │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │
│         │                 │                 │                 │         │
│         └─────────────────┼─────────────────┼─────────────────┘         │
│                           │                 │                           │
│                           ▼                 ▼                           │
│                    ┌─────────────────────────────────────────────────┐  │
│                    │        Context-aware Data Loading              │  │
│                    │   $VICE_DATA/{context}/habits.yml              │  │
│                    │   $VICE_DATA/{context}/entries.yml             │  │
│                    └─────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
```

### Key Architectural Changes

1. **config.Paths → ViceEnv**: Replace single-purpose config paths with comprehensive environment management
2. **XDG Full Compliance**: Extend from CONFIG-only to CONFIG/DATA/STATE/CACHE
3. **Context-aware Data**: User data segregated by context in $VICE_DATA/{context}/
4. **TOML Configuration**: App settings moved from code defaults to config.toml
5. **Layered Overrides**: ENV vars → CLI flags → config.toml → XDG defaults

### Architectural Analysis: Context-Aware Data Loading

#### Current State Issues:
1. **No Context Switching Support**: All data loading assumes single context
2. **Eager Loading**: CLI orchestration loads all data upfront 
3. **In-Memory State**: EntryCollector holds all data in memory maps
4. **File Path Coupling**: Direct dependency on config.Paths throughout

#### Architectural Options for Context-Aware Data Loading

```
OPTION 1: Repository Pattern with Context Manager
┌─────────────────────────────────────────────────────────────────────────┐
│                      DataRepository Interface                           │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │  LoadHabits(ctx Context) (*Schema, error)                          ││
│  │  LoadEntries(ctx Context, date Date) (*EntryLog, error)            ││
│  │  SaveEntries(ctx Context, entries *EntryLog) error                 ││
│  │  LoadChecklists(ctx Context) (*ChecklistSchema, error)             ││
│  └─────────────────────────────────────────────────────────────────────┘│
│                                   │                                     │
│                           ┌─────────────────┐                          │
│                           │ ContextManager  │                          │
│                           │• activeContext  │                          │
│                           │• SwitchContext()│                          │
│                           │• UnloadData()   │                          │
│                           └─────────────────┘                          │
│                                   │                                     │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │              FileRepository Implementation                          ││
│  │  • Uses ViceEnv for context-aware paths                            ││
│  │  • Delegates to existing parsers/storage                           ││
│  │  • Handles context switching and data unloading                    ││
│  └─────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────┘

Pros: Clean abstraction, testable, context-agnostic UI layer
Cons: Major refactoring required, potential over-engineering

``` 
--- 

**User notes**: "over-engineering" feels warranted here, it's core data and I
expect a lot of dependencies inbound - as long as it's not at odds with (good,
appropriate) bubbletea usage patterns or fighting the needs of the UI framework.
We can expect reloading context to be infrequent and a "full page reload"
appropriate in such cases, so perhaps we can make some simplifying assumptions
/ establish some conventions which let us keep things easy to reason about
without fighting the framework? (e.g. in case of context change "turn it off
and on again")

```
OPTION 2: Enhanced EntryCollector with Context Support
┌─────────────────────────────────────────────────────────────────────────┐
│                    ContextAwareEntryCollector                          │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │  • currentContext: string                                          ││
│  │  • viceEnv: *ViceEnv                                               ││
│  │  • dataState: map[string]*ContextData                              ││
│  │  • LoadForContext(context string) error                           ││
│  │  • SwitchContext(context string) error                            ││
│  │  • UnloadCurrentData()                                             ││
│  └─────────────────────────────────────────────────────────────────────┘│
│                                   │                                     │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                        ContextData                                  ││
│  │  • habits: []models.Habit                                          ││
│  │  • entries: map[string]interface{}                                 ││
│  │  • achievements: map[string]*models.AchievementLevel               ││
│  │  • loaded: bool                                                    ││
│  │  • lastAccess: time.Time                                           ││
│  └─────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────┘

Pros: Minimal changes to existing code, maintains BubbleTea patterns
Cons: EntryCollector becomes more complex, mixed responsibilities
```

**User notes**: Nahh. I don't like it. Discount this.

```
OPTION 3: Lazy-Loading Data Services
┌─────────────────────────────────────────────────────────────────────────┐
│                         DataServiceManager                             │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │  • viceEnv: *ViceEnv                                               ││
│  │  • loadedData: map[string]*LazyDataCache                           ││
│  │  • GetHabits() (*Schema, error)                                    ││
│  │  • GetEntries(date Date) (*EntryLog, error)                       ││
│  │  • SaveEntries(entries *EntryLog) error                           ││
│  │  • SwitchContext(context string)                                   ││
│  └─────────────────────────────────────────────────────────────────────┘│
│                                   │                                     │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                        LazyDataCache                                ││
│  │  • habits: *Schema (nil until loaded)                              ││
│  │  • entries: map[Date]*EntryLog                                     ││
│  │  • checklists: *ChecklistSchema                                    ││
│  │  • loadOnDemand(dataType string) error                             ││
│  └─────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────┘

Pros: True lazy loading, minimal memory usage, context isolation
Cons: Complex cache invalidation, potential race conditions
```

***User Notes***: Recent experience with actual race conditions (T024) tempers
my enthusiasm substantially for this approach. See above note about context
change; maybe there's a way to look at usage patterns which lets us keep things
simple if we establish and follow some rules. That said, at least for
analytics, and probably for several other cases involving historical data or
e.g search use cases, use cases I don't think we want to load all data
available into RAM. Consider T026 & fuzzy search over a large corpus of notes,
for example (if we have an in-memory search implementation instead of shelling
out to say fzf running over markdown files; tag/link based search and other
semantically aware discovery operations would probably want a DB with its own
memory management if we're to avoid this kind of memory management complexity
ourselves)

#### State Management Analysis

**Current EntryCollector State:**
- 5 separate maps for different data aspects
- `interface{}` values require type assertions
- No context awareness
- Direct parser/storage dependencies

**State Tracking Requirements for Context Switching:**
1. **Loaded Data Tracking**: Which data is currently in memory per context
2. **Dirty State Tracking**: Which data has unsaved changes  
3. **Context Isolation**: Ensure no data bleeding between contexts
4. **Unload Orchestration**: Clean unload of all data on context switch

**Information Hiding Options:**

```
OPTION A: Internal State Maps (Current Pattern)
• Pro: Familiar pattern, existing code works
• Con: No encapsulation, type safety issues

OPTION B: Typed Data Containers
• Pro: Type safety, clear data boundaries  
• Con: More verbose, requires data mapping

OPTION C: Interface-Based Data Access
• Pro: Abstraction, testability, future-proofing
• Con: Complexity, potential over-engineering for current needs
```

#### BubbleTea Integration Considerations

**Message Flow Impact:**
- Context switches must trigger UI state updates
- Modal lifecycle needs context awareness  
- Deferred state synchronization becomes more complex

**State Synchronization:**
- Current pattern relies on EntryCollector as single source of truth
- Context switching requires careful state invalidation
- UI components need notification of context changes

**Error Handling:**
- Context switches can fail (invalid context, file access issues)
- UI needs graceful degradation for context switch failures
- Partial loading scenarios need handling

#### Recommended Approach: Enhanced EntryCollector (Option 2)

**Rationale:**
1. **Minimal Disruption**: Preserves existing BubbleTea patterns
2. **Incremental Migration**: Can evolve toward repository pattern later
3. **Context Isolation**: Clear separation of data by context
4. **State Management**: Natural fit with existing UI state patterns

#### Migration Path Analysis

> **User Question**: Think about the repository vs lazy loading approaches outlined here in light of my comments. Is there a clear migration path later from the former to the latter if the need becomes apparent?

**Analysis of Migration Path: Repository → Lazy Loading**

The Repository Pattern (Option 1) provides an excellent foundation for later migration to sophisticated lazy loading (Option 3), with clear architectural benefits:

**Migration Advantages:**

1. **Interface Stability**: Repository interface abstracts implementation details
   ```go
   // Current Repository interface remains unchanged
   type DataRepository interface {
       LoadHabits(ctx Context) (*Schema, error)
       LoadEntries(ctx Context, date Date) (*EntryLog, error)
       SaveEntries(ctx Context, entries *EntryLog) error
   }
   
   // Implementation can evolve from simple to sophisticated
   // Phase 1: SimpleFileRepository (full reload on context switch)
   // Phase 2: CachedFileRepository (intelligent caching)
   // Phase 3: LazyLoadingRepository (on-demand loading)
   ```

2. **"Turn Off and On Again" Simplification**: Repository pattern naturally supports complete data unloading
   ```go
   func (r *Repository) SwitchContext(newContext string) error {
       r.UnloadAllData()           // Simple: clear all state
       r.context = newContext      // Switch context
       // Data loads on next access through repository methods
   }
   ```

3. **BubbleTea Integration**: Clean separation allows UI to remain unchanged
   ```go
   // UI components never change - always call repository
   habits, err := repo.LoadHabits(currentContext)
   
   // Repository implementation evolves independently:
   // - Phase 1: Always loads from disk
   // - Phase 2: Caches in memory with TTL
   // - Phase 3: Sophisticated lazy loading with dependency tracking
   ```

**Migration Path Stages:**

```
STAGE 1: Simple Repository (Immediate Implementation)
┌─────────────────────────────────────────────────────────────────────────┐
│                    SimpleFileRepository                                 │
│  • Context switching: complete unload + reload on next access          │
│  • No caching: every call hits disk                                    │
│  • Clear state management: loaded = true/false                         │
│  • Minimal complexity: no race conditions                              │
└─────────────────────────────────────────────────────────────────────────┘

STAGE 2: Cached Repository (Future Enhancement)
┌─────────────────────────────────────────────────────────────────────────┐
│                    CachedFileRepository                                 │
│  • Add simple in-memory cache with TTL                                 │
│  • Context switching: invalidate cache, lazy reload                    │
│  • Cache per context: map[context]*ContextCache                        │
│  • Still simple: cache hit/miss only                                   │
└─────────────────────────────────────────────────────────────────────────┘

STAGE 3: Lazy Loading Repository (Advanced Use Cases)
┌─────────────────────────────────────────────────────────────────────────┐
│                   LazyLoadingRepository                                 │
│  • Dependency-aware loading (e.g., habits before entries)              │
│  • Partial loading for large datasets (T026 fuzzy search)              │
│  • Background preloading for anticipated access                        │
│  • Complex invalidation and consistency management                     │
└─────────────────────────────────────────────────────────────────────────┘
```

**Key Migration Benefits:**

1. **Zero UI Changes**: Repository interface remains constant across all stages
2. **Testability**: Each stage can be thoroughly tested in isolation
3. **Rollback Safety**: Can revert to simpler implementation if complexity issues arise
4. **Incremental Complexity**: Add sophistication only when needed
5. **Clear Boundaries**: Data access logic completely separated from UI concerns

**Implementation Timeline:**

- **Phase 2 (Current Task)**: Implement SimpleFileRepository with "full reload" context switching
- **Future Enhancement**: Add caching when performance needs arise
- **Advanced Use Cases**: Implement lazy loading for T026 (large datasets) or analytics features

**Recommended Approach: Repository Pattern (Option 1) with Staged Implementation**

**Immediate Implementation (Phase 2):**
1. Create DataRepository interface with context-aware methods
2. Implement SimpleFileRepository with ViceEnv integration
3. "Turn off and on again" context switching for simplicity
4. Update EntryCollector to use repository instead of direct parser/storage access
5. Maintain all existing UI interfaces and BubbleTea patterns

**Future Migration Path:**
- Repository interface provides stable foundation for any internal implementation changes
- Can evolve from simple file access to sophisticated caching/lazy loading
- UI layer remains completely unaffected by internal repository evolution
- Clear rollback path if advanced implementations prove problematic

## Implementation Plan & Progress

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

- [x] **Phase 1: Core ViceEnv Structure & XDG Compliance**
  - [x] **1.1: Create ViceEnv struct with full XDG support**
    - *Design:* New `ViceEnv` struct with all XDG directories (CONFIG/DATA/STATE/CACHE), context-aware data paths
    - *Code/Artifacts:* `internal/config/env.go` (new), replace `config.Paths` usage
    - *Testing Strategy:* Unit tests for env variable resolution, XDG path construction
    - *AI Notes:* Priority override: ENV vars → CLI flags → XDG defaults
    - create paths if missing
    - represent default settings in code
    - stub: load default settings into ViceEnv (no config.toml yet)
  - [x] **1.2: Add pelletier/go-toml dependency and TOML parsing**
    - *Design:* Add go-toml to go.mod, parse config.toml from $VICE_CONFIG for app settings
    - *Code/Artifacts:* `internal/config/toml.go` (new), update go.mod
    - *Testing Strategy:* Unit tests for TOML parsing, invalid config handling, defaults
    - *AI Notes:* config.toml = app settings (keybindings, themes, contexts array), not user data
  - [x] **1.3: Create config.toml with default settings if missing**
    - *Design:* Initialize config.toml with default [core] contexts = ["personal", "work"] if missing
    - *Code/Artifacts:* Update `internal/config/env.go`, file initialization logic
    - *Testing Strategy:* Unit tests for config.toml creation, default value handling
    - *AI Notes:* Undo stub loading of defaults into ViceEnv, load from config.toml instead

### Phase 1 Implementation Analysis

**Current State**: `config.Paths` struct provides basic XDG CONFIG directory support with hardcoded YAML file paths

**Target State**: `ViceEnv` struct with full XDG compliance and TOML-based configuration

**Key Changes Required**:

1. **Replace config.Paths entirely** - 11 cmd files + UI components use `GetPaths()`
2. **Add 4 XDG directories** - extend from CONFIG-only to CONFIG/DATA/STATE/CACHE
3. **Context-aware path resolution** - data paths become $VICE_DATA/{context}/file.yml
4. **TOML configuration parsing** - add pelletier/go-toml dependency, parse config.toml
5. **Maintain CLI compatibility** - preserve `--config-dir` flag behavior

**Implementation Strategy**:
- Build ViceEnv alongside config.Paths initially
- Update `cmd/root.go` to use ViceEnv first  
- Phase out config.Paths usage across cmd files
- Update UI constructors to accept ViceEnv instead of paths
- Ensure backward compatibility during transition
- Ensure good Anchor comments and grep codebase to ensure no access bypassing new design. 

- [x] **Phase 2: Context Management System**
  - [x] **2.1: Implement context switching with immediate data unload**
    - *Design:* Context manager unloads all data immediately, loads on-demand per UI needs
    - *Code/Artifacts:* `internal/config/context.go` (new), state file management
    - *Testing Strategy:* Unit tests for context switching, data unloading behavior
    - *AI Notes:* No eager loading - load data only when UI components request it
  - [x] **2.2: Add context-aware data directory management**
    - *Design:* Auto-create $VICE_DATA/$CONTEXT directories, populate with minimal YAML files
    - *Code/Artifacts:* Update `ViceEnv` struct, directory creation logic
    - *Testing Strategy:* Integration tests for context directory creation
    - *AI Notes:* First context in array is default, create dirs on first access
    - **Completed Steps:**
      - [x] Step 2.2.1: Context-Aware FileInitializer - Added `EnsureContextFiles(env *ViceEnv)` and checklist initialization
        - *Artifacts:* `internal/init/files.go` - Added ViceEnv support and checklist file creation
        - *Testing:* Unit tests pass, FileInitializer creates all 4 data files per context
      - [x] Step 2.2.2: CMD Layer Integration - Updated `cmd/root.go` and `cmd/entry.go` to use ViceEnv 
        - *Artifacts:* `cmd/root.go`, `cmd/entry.go` - Replaced `initializePaths` with `initializeViceEnv`
        - *Testing:* CMD tests updated and passing, maintains backward compatibility
      - [x] Step 2.2.3: Context Data File Templates - Implemented checklist file creation methods
        - *Artifacts:* Added `createEmptyChecklistsFile()` and `createEmptyChecklistEntriesFile()` 
        - *Testing:* Integration with parser layer verified
      - [x] Step 2.2.4: CLI Updates - Updated all 9 remaining cmd files to use ViceEnv instead of config.Paths
        - *Artifacts:* Updated `habit_add.go`, `habit_list.go`, `habit_edit.go`, `habit_remove.go`, `list_add.go`, `list_edit.go`, `list_entry.go`, `todo.go`
        - *Testing:* All commands build and run successfully
      - [x] Step 2.2.5: Repository Integration - Integrated FileInitializer with Repository pattern for automatic file creation
        - *Artifacts:* `internal/repository/file_repository.go` - Added FileInitializer to LoadHabits, LoadEntries, LoadChecklists
        - *Testing:* Repository tests updated, automatic file creation verified

### Phase 2.2 Detailed Implementation Steps

**Initial State Analysis:**
- `internal/init/files.go`: Creates habits.yml + entries.yml with hardcoded paths
- 11 cmd files use `GetPaths()` from legacy config.Paths system
- Sample data: 4 comprehensive habits (2 simple, 2 elastic) + empty entries
- Missing: checklist file initialization, context awareness

**Step-by-Step Implementation Completed:**

**Step 2.2.1: Create Context-Aware FileInitializer** ✅
- *Files:* `internal/init/files.go`
- *Detailed Sub-tasks Completed:*
  - [x] Add import for `internal/config` package
  - [x] Add FileInitializer struct fields for checklist parsers (`checklistParser`, `checklistEntriesParser`)
  - [x] Update `NewFileInitializer()` to initialize checklist parsers
  - [x] Create `EnsureContextFiles(env *ViceEnv)` method with ViceEnv integration
  - [x] Create helper methods: `ensureHabitsFile()`, `ensureEntriesFile()`, `ensureChecklistsFile()`, `ensureChecklistEntriesFile()`
  - [x] Implement `createEmptyChecklistsFile()` with proper ChecklistSchema structure
  - [x] Implement `createEmptyChecklistEntriesFile()` using `parser.CreateEmptySchema()`
  - [x] Add ANCHOR comment: T028-context-files for future reference
  - [x] Unit tests pass for file initialization functionality

**Step 2.2.2: Update CMD Layer Integration** ✅
- *Files:* `cmd/root.go`, `cmd/entry.go`
- *Detailed Sub-tasks Completed:*
  - [x] Replace global `paths *config.Paths` with `viceEnv *config.ViceEnv` in root.go
  - [x] Rename `initializePaths` to `initializeViceEnv` with ViceEnv initialization logic
  - [x] Update `GetPaths()` to return legacy config.Paths for backward compatibility
  - [x] Create new `GetViceEnv()` function for modern usage
  - [x] Update `runDefaultCommand()` to use `EnsureContextFiles(env)` instead of hardcoded paths
  - [x] Move `runEntryMenu()` from entry.go to root.go for shared access
  - [x] Move `loadTodayEntries()` function to root.go and add required imports
  - [x] Update `runEntry()` in entry.go to use ViceEnv and `EnsureContextFiles()`
  - [x] Clean up unused imports in entry.go
  - [x] Update cmd tests to use `initializeViceEnv` and fix path expectations for context structure
  - [x] Add ANCHOR comment: T028-cmd-integration for future reference

**Step 2.2.3: Add Context Data File Templates** ✅
- *Files:* `internal/init/files.go` (template methods)
- *Detailed Sub-tasks Completed:*
  - [x] Implement `createEmptyChecklistsFile()` with basic ChecklistSchema (version "1.0.0", empty checklists array)
  - [x] Implement `createEmptyChecklistEntriesFile()` using ChecklistEntriesParser.CreateEmptySchema()
  - [x] Ensure all 4 data files initialized per context: habits.yml, entries.yml, checklists.yml, checklist_entries.yml
  - [x] Maintain consistent sample habits across all contexts (4 comprehensive habits: 2 simple, 2 elastic)
  - [x] Integration testing with parser layer verified

**Step 2.2.4: Integration Testing & CLI Updates** ✅
- *Files:* All remaining cmd files using `GetPaths()` (9 files)
- *Detailed Sub-tasks Completed:*
  - [x] **Core Commands Priority 1:**
    - [x] `habit_add.go`: Replace `GetPaths()` with `GetViceEnv()`, update `EnsureConfigFiles()` to `EnsureContextFiles()`
    - [x] `habit_list.go`: Replace paths with env, update file initialization calls
    - [x] `todo.go`: Replace with ViceEnv (kept backward compatibility for TodoDashboard)
  - [x] **Secondary Commands Priority 2:**
    - [x] `habit_edit.go`: Update to use `env.GetHabitsFile()` and context-aware initialization
    - [x] `habit_remove.go`: Replace paths with env for habit removal operations
    - [x] `list_add.go`: Update to use `env.GetChecklistsFile()` for all checklist operations
    - [x] `list_edit.go`: Replace `paths.ChecklistsFile` with `env.GetChecklistsFile()`
    - [x] `list_entry.go`: Update all checklist file paths to use ViceEnv methods
  - [x] Fix `prototype.go` to use ViceEnv for debug logging paths
  - [x] Resolve compilation errors and unused variables
  - [x] All commands build successfully and maintain CLI compatibility
  - [x] Dry-run functionality preserved across all habit and list commands

**Step 2.2.5: Repository Integration** ✅
- *Files:* `internal/repository/file_repository.go`, `internal/repository/file_repository_test.go`
- *Detailed Sub-tasks Completed:*
  - [x] Add `init_pkg "davidlee/vice/internal/init"` import to repository
  - [x] Add `fileInitializer *init_pkg.FileInitializer` field to FileRepository struct
  - [x] Update `NewFileRepository()` to initialize FileInitializer
  - [x] Add `EnsureContextFiles()` call to `LoadHabits()` method with error handling
  - [x] Add `EnsureContextFiles()` call to `LoadEntries()` method with error handling  
  - [x] Add `EnsureContextFiles()` call to `LoadChecklists()` method with error handling
  - [x] Add ANCHOR comment: T028/2.2-file-init for automatic file creation behavior
  - [x] Update repository test `TestLoadHabitsFileNotFound` to `TestLoadHabitsAutoCreateFiles`
  - [x] Verify automatic file creation behavior in repository tests
  - [x] Add import for `os` package in repository tests for file existence verification
  - [x] All repository tests pass with new automatic file creation behavior

**Integration Points Verified:**
- **ViceEnv Methods:** Successfully using `GetHabitsFile()`, `GetEntriesFile()`, `GetChecklistsFile()`, `GetChecklistEntriesFile()`
- **Directory Creation:** Leveraging `env.EnsureDirectories()` from Phase 2.1 working correctly
- **Context Switching:** Files created automatically when switching to new context via repository
- **Backward Compatibility:** Legacy `GetPaths()` function maintained for transition period
- **Testing:** All unit tests, integration tests, and build verification passing

- [ ] **Phase 3: CLI Integration & Context Flags**
  - [x] **3.1: Update CLI to use ViceEnv throughout and add XDG directory flags**
    - *Design:* Replace config.Paths usage with ViceEnv in cmd/root.go and subcommands, add --data-dir, --cache-dir, --state-dir flags alongside existing --config-dir
    - *Code/Artifacts:* `cmd/root.go`, all subcommand files, flag definitions and ViceEnv integration
    - *Testing Strategy:* CLI integration tests, existing command compatibility, flag override verification
    - *AI Notes:* Maintain --config-dir flag behavior, add --data-dir/--cache-dir/--state-dir flags, environment variable support
    - **Implementation Summary:**
      - Added --data-dir, --cache-dir, --state-dir, --context flags with comprehensive help text
      - Created DirectoryOverrides struct for clean function signatures
      - Updated GetViceEnvWithOverrides() to handle all directory overrides
      - Migrated TodoDashboard from legacy config.Paths to ViceEnv with backward compatibility
      - Fixed all linting issues (errcheck, gosec, revive, staticcheck, unused)
      - Improved security with 0600 file permissions and proper error handling
      - All tests pass, application builds successfully, CLI functionality verified
  - [x] **3.2: Add transient --context CLI flag**
    - *Design:* Global --context flag for all operations, always transient (no state persistence)
    - *Code/Artifacts:* Update root command flags, context resolution logic
    - *Testing Strategy:* CLI tests with --context flag, verify no state persistence
    - *AI Notes:* Transient override for all operations, doesn't modify $VICE_STATE/vice.yml
    - **Already Implemented:** Completed as part of Phase 3.1 - --context flag is fully functional with transient behavior

- [x] **Phase 4: Runtime Context Operations**
  - [x] **4.1: Add context switching commands**
    - *Design:* Add `vice context` subcommand with list/switch operations for persistence
    - *Code/Artifacts:* `cmd/context.go` (new), context CLI interface
    - *Testing Strategy:* Manual testing with multiple contexts, state persistence
    - *AI Notes:* Interactive context switching persists to state file
    - **Implementation Summary:**
      - Created comprehensive `vice context` command with list/show/switch subcommands
      - Context switching validates available contexts and persists changes to state file
      - Added proper error handling for invalid contexts and edge cases
      - Updated main help with context management examples
      - Added ANCHOR comments for future reference: T028/4.1-context-list, T028/4.1-context-switch
      - Full test coverage with context command validation
  - [x] **4.2: Add environment variable override support**
    - *Design:* VICE_CONTEXT override, runtime context switching capability
    - *Code/Artifacts:* Update ViceEnv initialization, context resolution
    - *Testing Strategy:* Environment variable integration tests
    - *AI Notes:* ENV var should override persisted state but not modify it
    - **Already Implemented:** VICE_CONTEXT environment variable support was completed in earlier phases
    - **Implementation Summary:**
      - VICE_CONTEXT environment variable provides transient context override
      - Works alongside --context CLI flag for temporary context switching
      - Updated CLI help to document environment variable usage with examples
      - Priority order: CLI flag → ENV var → persisted state → default context

- [x] **Phase 5: Documentation & Specification Updates**
  - [x] **5.1: Extract implementation details to specification**
    - *Design:* Update `doc/specifications/file_paths_runtime_env.md` with implementation learnings
    - *Code/Artifacts:* Update specification with architectural decisions, data loading patterns
    - *Testing Strategy:* Documentation review, specification accuracy validation
    - *AI Notes:* Include Repository Pattern decision, migration path, BubbleTea integration patterns
    - *Implementation Summary:* 
      - Complete rewrite of specification with comprehensive implementation details
      - Added all architectural decisions: Repository Pattern, "turn off and on again" context switching
      - Documented three context switching methods with priority resolution
      - Included BubbleTea integration patterns and UI state management
      - Added migration path from simple → cached → lazy loading repositories
      - Documented security considerations, operational behavior, and edge cases
      - Specification now serves as authoritative reference for the implemented system

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
- `2025-07-16 - AI:` Completed pre-flight analysis for Phase 1:
  - Added comprehensive ASCII architecture diagram (current vs target)  
  - Analyzed current config.Paths usage across 11 cmd files + UI components
  - Identified key changes: full XDG compliance, context-aware paths, TOML config
  - Implementation strategy: build ViceEnv alongside config.Paths, then phase out
- `2025-07-16 - AI:` Completed Phase 1 implementation:
  - 1.1: Created ViceEnv struct with full XDG support (CONFIG/DATA/STATE/CACHE)
  - 1.2: Added pelletier/go-toml dependency and TOML parsing infrastructure
  - 1.3: Integrated config.toml creation and loading into ViceEnv initialization
  - All functionality tested with comprehensive unit and integration tests
- `2025-07-16 - AI:` Completed architectural analysis of data loading/persistence:
  - Created comprehensive ASCII diagrams of current architecture
  - Analyzed EntryCollector state management and BubbleTea integration patterns
  - Evaluated 3 architectural options for context-aware data loading
  - Based on user feedback, updated recommendation to Repository Pattern (Option 1)
  - Documented clear migration path from simple repository to lazy loading
  - Recommended "turn off and on again" context switching for simplicity
- `2025-07-16 - User:` Accepted Repository Pattern approach
- `2025-07-16 - AI:` Added specification cross-reference and Phase 5 documentation task
- `2025-07-16 - AI:` Completed Phase 2.1 implementation:
  - Created DataRepository interface with context-aware operations
  - Implemented FileRepository with "turn off and on again" context switching  
  - Added context state persistence in vice.yml with priority handling
  - Integrated context initialization into ViceEnv setup
  - Full test coverage for repository and context management
- `2025-07-16 - AI:` Added comprehensive ANCHOR comments for future reference:
  - T028-repository-interface, T028-race-condition-avoidance, T028-context-validation
  - T028-xdg-compliance, T028-priority-resolution, T028-initialization-flow
  - T028-toml-config, T028-config-integration, T028-state-yaml-structure
- `2025-07-16 - AI:` Completed pre-flight analysis and detailed Phase 2.2 planning:
  - Analyzed current file initialization patterns (internal/init/files.go)
  - Identified 11 cmd files using legacy config.Paths system
  - Broke down Phase 2.2 into 5 detailed implementation steps
  - Defined integration points with ViceEnv and Repository pattern
- `2025-07-16 - AI:` Completed Phase 2.2 implementation:
  - **Step 2.2.1:** Extended FileInitializer with context-aware `EnsureContextFiles(env *ViceEnv)` method
    - Added checklist parser integration for complete 4-file initialization per context
    - ANCHOR: T028-context-files in internal/init/files.go
  - **Step 2.2.2:** Updated core CMD layer (root.go, entry.go) to use ViceEnv 
    - Replaced `initializePaths` with `initializeViceEnv` and `GetPaths()` with `GetViceEnv()`
    - Moved `runEntryMenu` to root.go for shared access, added backward compatibility
  - **Step 2.2.3:** Implemented checklist file creation templates
    - Added `createEmptyChecklistsFile()` and `createEmptyChecklistEntriesFile()` methods
    - Integrated with existing parser.ChecklistParser and parser.ChecklistEntriesParser
  - **Step 2.2.4:** Updated all remaining cmd files to use ViceEnv
    - Migrated 9 remaining cmd files from config.Paths to ViceEnv pattern
    - Maintained CLI compatibility and dry-run functionality throughout
  - **Step 2.2.5:** Integrated FileInitializer with Repository pattern
    - Added automatic file creation to LoadHabits, LoadEntries, and LoadChecklists methods
    - Updated repository tests to verify automatic file creation behavior
    - ANCHOR: T028/2.2-file-init in internal/repository/file_repository.go
- `2025-07-17 - AI:` Completed Phase 3.1 implementation:
  - Added comprehensive XDG directory flags (--data-dir, --cache-dir, --state-dir, --context)
  - Completed ViceEnv migration throughout application, including TodoDashboard
  - Fixed all linting issues for production-ready code quality
  - All functionality tested and verified working correctly
  - Ready to proceed with Phase 3.2 or Phase 4 as requested
- `2025-07-17 - AI:` Completed Phase 4 implementation:
  - **Context Management System:** Comprehensive context switching with three methods:
    1. **Persistent:** `vice context switch <name>` - persists to state file
    2. **CLI Flag:** `--context <name>` - transient override per command
    3. **Environment:** `VICE_CONTEXT=<name>` - transient override via env var
  - **Priority Resolution:** CLI flag → ENV var → persisted state → default context
  - **Data Isolation:** Each context maintains completely separate data directories
  - **User Experience:** Clear help documentation with examples for all methods
  - **Quality:** Full test coverage, proper error handling, ANCHOR comments added

## Implementation Details & System Behavior

### Directory Structure & File Management
- **XDG Compliance:** Full implementation of XDG Base Directory Specification
  - CONFIG: `$XDG_CONFIG_HOME/vice` or `~/.config/vice` - stores config.toml
  - DATA: `$XDG_DATA_HOME/vice` or `~/.local/share/vice` - stores user data by context
  - STATE: `$XDG_STATE_HOME/vice` or `~/.local/state/vice` - stores vice.yml (active context)
  - CACHE: `$XDG_CACHE_HOME/vice` or `~/.cache/vice` - reserved for future caching

### Context Data Isolation
- **Per-context directories:** `$VICE_DATA/{context}/` containing:
  - `habits.yml` - habit definitions (4 sample habits created automatically)
  - `entries.yml` - daily habit completion data  
  - `checklists.yml` - checklist templates
  - `checklist_entries.yml` - checklist completion data
- **Automatic file creation:** Repository pattern ensures files exist before data operations
- **Sample data consistency:** All contexts get identical 4-habit sample set (2 simple, 2 elastic)

### Configuration System Architecture
- **config.toml structure:** 
  ```toml
  [core]
  contexts = ["personal", "work"]  # defines available contexts
  ```
- **State persistence in vice.yml:**
  ```yaml
  version: "1.0"
  active_context: "personal"  # last context set via 'vice context switch'
  ```
- **Priority resolution order:** CLI --context flag → VICE_CONTEXT env var → vice.yml state → first context in config.toml → "personal" fallback

### Context Switching Behavior
- **Persistent switching** (`vice context switch`):
  - Validates context exists in config.toml contexts array
  - Updates ViceEnv.Context and recomputes ContextData path
  - Calls EnsureDirectories() to create context directory structure
  - Persists change to $VICE_STATE/vice.yml for future sessions
  - Auto-creates data files via Repository.LoadHabits/LoadEntries/LoadChecklists
- **Transient overrides** (--context flag, VICE_CONTEXT env):
  - Override active context for single command execution only
  - Do NOT modify vice.yml state file
  - CLI flag takes precedence over environment variable
  - Temporary ContextData path computation for the override context

### Repository Pattern & Data Loading
- **"Turn off and on again" approach:** Complete data unload on context switch for simplicity
- **Lazy loading:** Data loaded on-demand when UI components request it
- **Automatic initialization:** FileInitializer.EnsureContextFiles() called by repository methods
- **Error handling:** Repository.Error type provides context-aware error messages

### ViceEnv Migration Strategy
- **Backward compatibility:** Legacy GetPaths() function maintained during transition
- **TodoDashboard migration:** Updated to use ViceEnv with NewTodoDashboardLegacy() for compatibility
- **Gradual replacement:** All cmd files migrated from config.Paths to ViceEnv pattern
- **DirectoryOverrides struct:** Clean abstraction for CLI flag handling

### Security & File Permissions
- **Configuration files:** 0600 permissions (user read/write only) for config.toml and vice.yml
- **Data directories:** 0750 permissions (user read/write/execute, group read/execute)
- **Controlled file paths:** All os.ReadFile operations use validated, controlled paths

### Testing Strategy
- **Unit tests:** ViceEnv creation, configuration loading, context switching logic
- **Integration tests:** Complete environment setup with temporary directories
- **Command tests:** Context command validation and error handling
- **Manual verification:** All flag combinations tested with real data operations

### Performance Considerations
- **Minimal overhead:** Directory flag parsing only affects startup initialization
- **Efficient context switching:** No data migration, just path recomputation
- **File system operations:** Directories created lazily only when needed
- **Memory usage:** No eager loading of data across contexts

### Operational Behavior & Edge Cases
- **Missing config.toml:** Automatically created with default contexts ["personal", "work"]
- **Invalid context in vice.yml:** Falls back to first context in config.toml array
- **Context not in config.toml:** `vice context switch` validates and rejects with clear error
- **Empty contexts array:** Falls back to "personal" default context
- **Directory creation failures:** Clear error messages with specific paths and permissions info
- **File permission issues:** Error messages guide user to check directory permissions

### CLI Integration Points
- **Global flags:** All directory flags (--config-dir, --data-dir, etc.) available to all commands
- **Flag precedence:** CLI flags override environment variables override XDG defaults
- **Help system:** Context examples integrated into main help and context-specific help
- **Command discovery:** `vice context` appears in main command list with proper description
- **Error consistency:** All context-related errors follow repository.Error pattern for consistency

### Migration & Compatibility Notes
- **Legacy support:** GetPaths() function maintained for any remaining legacy code
- **TodoDashboard bridge:** NewTodoDashboardLegacy() provides compatibility during transition
- **No breaking changes:** All existing functionality preserved while adding new capabilities
- **Gradual adoption:** New features can be adopted incrementally without disrupting existing workflows

## Git Commit History

- `9c87759` - docs(kanban)[T028]: create file paths & runtime environment implementation plan
- `b8feecc` - docs(kanban)[T028]: update plan based on user clarifications
- `7486b46` - docs(kanban)[T028]: add comprehensive Phase 1 pre-flight analysis
- `1f80ede` - feat(config)[T028/1.1-1.3]: implement ViceEnv with full XDG compliance and TOML config
- `0e06438` - docs(kanban)[T028]: detailed Phase 2.2 pre-flight analysis and step-by-step plan
- `f922930` - docs(anchor)[T028]: add comprehensive ANCHOR comments for future reference
- `bc71812` - feat(init)[T028/2.2]: implement context-aware data directory management
- `56655cd` - feat(cli)[T028/3.1]: add comprehensive XDG directory flags and complete ViceEnv migration
- `a7cda7d` - feat(context)[T028/4.1-4.2]: implement comprehensive context management system
- `8ec3e4b` - docs(spec)[T028/5.1]: extract comprehensive implementation details to specification
- `0016c8f` - docs(kanban)[T026,T027]: update flotsam tasks with T028 integration and storage options