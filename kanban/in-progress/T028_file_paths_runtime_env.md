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

**Implementation Strategy:**
1. Add context awareness to EntryCollector
2. Implement ContextData containers for type safety
3. Add lazy loading with on-demand data access
4. Integrate with ViceEnv for context-aware file paths
5. Preserve existing UI interfaces during transition

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
    - NEEDS CONSIDERATION: non-interactive commands must be explicitly understood as such by code handling --context flag
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
  - Recommended Enhanced EntryCollector approach (Option 2) for Phase 2

## Git Commit History

- `9c87759` - docs(kanban)[T028]: create file paths & runtime environment implementation plan
- `b8feecc` - docs(kanban)[T028]: update plan based on user clarifications
- `7486b46` - docs(kanban)[T028]: add comprehensive Phase 1 pre-flight analysis
- `1f80ede` - feat(config)[T028/1.1-1.3]: implement ViceEnv with full XDG compliance and TOML config