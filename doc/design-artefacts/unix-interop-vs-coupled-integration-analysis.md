# Unix Interop vs Coupled Integration Analysis

**Status**: Design exploration - Multiple decisions to be extracted

**Purpose**: Comprehensive analysis of architectural alternatives for flotsam functionality, comparing T027's coupled approach with Unix interop patterns inspired by zk.

**Strategic Insight**: This analysis reveals a fundamental repositioning of vice from a monolithic habit tracker to a **TUI front-end that orchestrates Unix tools** for productivity workflows.

## Context

After implementing T027 (flotsam data layer with repository patterns, backlink computation, and SRS scheduling), we need to evaluate whether this tightly coupled approach is optimal for vice's flotsam functionality.

Examining zk's documentation reveals a Unix philosophy approach that achieves similar outcomes through:
- CLI composability and piping
- External tool integration via standard interfaces
- Minimal coupling between components
- Delegation to specialized tools

### Current T027 Implementation
- **Repository Layer**: Full CRUD operations for flotsam notes
- **Backlink Computation**: Context-scoped automatic backlink resolution
- **SRS Integration**: Direct scheduling and due date computation
- **Data Models**: Rich Go structs with validation and serialization
- **Storage**: Direct file system operations with caching

### Unix Interop Alternative (zk-inspired)
- **CLI as Primary Interface**: `vice flotsam list|new|edit` with rich output formatting
- **External Tool Delegation**: Editor integration, fuzzy finding, file processing
- **Composable Commands**: Pipeline-friendly output formats (JSON, delimited, templates)
- **Minimal Coupling**: Each command does one thing well
- **Standard Interfaces**: File paths, stdin/stdout, environment variables

## Key Insights from zk Documentation

### 1. External Processing (`external-processing.md`)
```sh
# Get file paths for external processing
zk list --format path --delimiter " "
zk list --format "'{{path}}'" --delimiter " "  # spaces in paths
zk list --format path --delimiter0 | xargs -0 git log --patch --

# Process note content directly
zk list --format {{raw-content}} --limit 1

# Compose zk commands
zk list --exclude "`zk list -q -f path -d "," --orphan`"
```

### 2. External Tool Integration (`external-call.md`)
```sh
# Background-friendly options
zk command --no-input --quiet
```

### 3. Workflow Automation (`daily-journal.md`)
```sh
# Simple alias for complex workflow
[alias]
daily = 'zk new --no-input "$ZK_NOTEBOOK_DIR/journal/daily"'
```

## Decision Framework

We need to evaluate whether vice's flotsam functionality can achieve the same user value through Unix interop patterns instead of the coupled implementation.

### Core Use Cases to Evaluate

1. **Note Creation & Editing**
   - Current: Repository.CreateFlotsamNote() + Editor integration
   - Alternative: `vice flotsam new` + external editor via $EDITOR

2. **Note Discovery & Navigation**
   - Current: Repository.LoadFlotsam() + backlink computation
   - Alternative: `vice flotsam list` with filtering + fuzzy finder

3. **SRS Scheduling**
   - Current: Repository.GetDueFlotsamNotes() + direct scheduling
   - Alternative: `vice flotsam due` + external scheduling tools

4. **Backlink Management**
   - Current: Automatic backlink computation and storage
   - Alternative: `vice flotsam backlinks` + external link analysis

5. **Bulk Operations**
   - Current: Repository batch operations
   - Alternative: `vice flotsam list --format path | xargs ...`

### Evaluation Criteria

#### Unix Philosophy Alignment
- **Do one thing well**: Each command has single responsibility
- **Composable**: Commands work together via pipes and standard interfaces
- **Text streams**: Data flows through text-based interfaces
- **External tools**: Leverage existing ecosystem (fzf, editors, etc.)

#### Implementation Complexity
- **Coupling**: How tightly are components connected?
- **Maintenance**: How much code to maintain and test?
- **Dependencies**: Internal vs external tool dependencies
- **Extensibility**: How easy to add new functionality?

#### User Experience
- **Discoverability**: Can users find and understand functionality?
- **Composability**: Can users create custom workflows?
- **Performance**: Response time and resource usage
- **Integration**: How well does it fit existing workflows?

## Proposed Investigation

### Phase 1: CLI Command Prototyping
Create minimal CLI commands that demonstrate Unix interop:

```sh
# Note management
vice flotsam new [path]         # Create new note, open in $EDITOR
vice flotsam edit [filter]      # Edit existing notes via fuzzy finder
vice flotsam list [filter]      # List notes with rich formatting options

# Discovery and navigation
vice flotsam backlinks <note>   # Show backlinks to note
vice flotsam links <note>       # Show outbound links from note
vice flotsam due                # Show notes due for review

# Formatting and integration
vice flotsam list --format json
vice flotsam list --format path --delimiter0
vice flotsam list --format template --template "{{path}}: {{title}}"
```

### Phase 2: External Tool Integration
Demonstrate integration with existing tools:

```sh
# Editor integration
vice flotsam edit --interactive  # Use fzf for selection
vice flotsam edit daily-notes    # Edit matching notes

# Pipeline integration
vice flotsam list --format path | xargs grep "keyword"
vice flotsam due --format json | jq '.[] | select(.priority > 3)'

# Workflow automation
alias daily='vice flotsam new journal/$(date +%Y-%m-%d)'
```

### Phase 3: Comparison Analysis
Compare the Unix interop approach against T027 implementation:

1. **Functionality Coverage**: Can all T027 use cases be achieved?
2. **Code Complexity**: Lines of code, test coverage, maintenance burden
3. **User Experience**: Learning curve, discoverability, workflow integration
4. **Performance**: Command execution time, resource usage
5. **Extensibility**: Adding new features, third-party integration

## Decisions to Extract

This analysis has identified several discrete decisions that should become separate ADRs:

1. **ADR-008: SRS Storage Strategy** - Database vs frontmatter vs hybrid approaches
2. **ADR-009: ZK Dependency Management** - How to handle zk as external dependency
3. **ADR-010: Unix Interop vs Coupled Integration** - Primary architectural approach
4. **ADR-011: Tag-based Note Behaviors** - Using zk tags for vice-specific functionality
5. **ADR-012: Cache Consistency Strategy** - mtime vs file watching approaches

## Tasks to Extract

1. **T041: Unix Interop Prototype** - Minimal CLI demonstrating shell-out patterns
2. **T042: ZK Dependency Installation** - `vice install-deps` and `vice doctor` commands
3. **T043: SRS Database Implementation** - Separate SQLite database with mtime validation
4. **T044: Tag-based Behavior System** - `vice:*` tag patterns and integration
5. **T045: ZK Configuration Management** - Write zk config options during notebook initialization
6. **T046: User ZK Customization** - Allow users to configure zk behavior via vice config

## Summary & Decision

**Analysis Conclusion**: The Unix interop approach is superior to T027's coupled integration for flotsam functionality.

**Key Findings**:
1. **Reduced Complexity**: ~500 lines vs ~2000+ lines of implementation
2. **Better Performance**: Cache + mtime validation vs loading everything fresh
3. **Leveraged Capabilities**: zk provides sophisticated search, linking, and editor integration
4. **User Customization**: Configuration through vice config, zk behavior control
5. **Strategic Positioning**: Enables evolution to Unix tool orchestrator rather than monolithic tracker

**Critical Discoveries**:
- zk supports custom filenames, external processing, and rich tag systems
- Combined search (zk + SRS) is manageable with targeted caching
- External editor integration can be fully delegated to zk
- Tool orchestration architecture enables future integrations (remind, taskwarrior, MCP)

**Decision**: Proceed with Unix interop approach for flotsam functionality.

## Next Steps

1. **Extract ADRs**: Create focused decision documents for each identified choice
2. **Prototype validation**: Build minimal Unix interop demo to validate approach
3. **Migration planning**: Plan T027 migration strategy
4. **Implementation**: Begin with core Unix interop tasks

## Strategic Repositioning: Vice as Unix Tool Orchestrator

**Vision Evolution**: From monolithic habit tracker to **TUI front-end for Unix productivity tools**

### Current Positioning vs Future Vision

**Current (T027 approach)**:
- Vice as self-contained habit tracker
- Custom implementations for everything
- Monolithic architecture
- Limited extensibility

**Future Vision (Unix interop)**:
- Vice as **orchestration layer** for Unix tools
- **TUI front-end** providing unified workflows
- **Tool integration** rather than reimplementation
- **Extensible architecture** for adding new tool integrations

### Tool Integration Roadmap

**Phase 1: Knowledge Management (zk integration)**
- Flotsam notes via zk
- SRS scheduling database
- Tag-based behaviors
- Editor integration

**Phase 2: Task Management (remind/taskwarrior integration)**
```bash
# remind integration for recurring tasks
vice remind add "Review flotsam notes" --daily
vice remind list --today

# taskwarrior integration for project management
vice task add "Implement SRS database" project:vice
vice task list project:vice
```

**Phase 3: Unified Workflows**
```bash
# Cross-tool workflows
vice workflow "weekly-review" --combine remind,zk,taskwarrior
  # 1. Show overdue tasks (taskwarrior)
  # 2. Review SRS notes (zk + vice SRS)
  # 3. Plan next week (remind + zk)
```

### Benefits of Tool Orchestration Approach

**1. Leverage Best-in-Class Tools**
- **zk**: Note management, search, linking
- **remind**: Sophisticated recurring task scheduling
- **taskwarrior**: Project management, contexts, reporting
- **vice**: TUI interface + workflow orchestration

**2. Reduced Implementation Burden**
- Don't reimplement what tools do well
- Focus on integration and user experience
- Maintain smaller, more focused codebase

**3. User Investment Protection**
- Users can use tools independently
- Existing tool configurations/data preserved
- Gradual adoption possible

**4. Extensibility**
- Easy to add new tool integrations
- Community can contribute tool connectors
- Modular architecture supports experimentation

### Implementation Strategy

**Core Architecture**:
```go
// Tool abstraction layer
type Tool interface {
    Name() string
    IsAvailable() bool
    Execute(cmd string, args ...string) (Result, error)
}

// Workflow orchestration
type Workflow struct {
    Name string
    Steps []WorkflowStep
}

type WorkflowStep struct {
    Tool     Tool
    Command  string
    Args     []string
    OnResult func(Result) error
}
```

**Example Integration**:
```go
// remind integration
type RemindTool struct{}

func (r *RemindTool) Execute(cmd string, args ...string) (Result, error) {
    switch cmd {
    case "add":
        return r.addReminder(args...)
    case "list":
        return r.listReminders(args...)
    }
}

// Cross-tool workflow
func weeklyReviewWorkflow() *Workflow {
    return &Workflow{
        Name: "weekly-review",
        Steps: []WorkflowStep{
            {Tool: &TaskwarriorTool{}, Command: "list", Args: []string{"status:pending"}},
            {Tool: &ViceTool{}, Command: "flotsam", Args: []string{"due", "--this-week"}},
            {Tool: &ZkTool{}, Command: "list", Args: []string{"--tag", "review"}},
        },
    }
}
```

### Value Proposition

**For Users**:
- **Unified interface** for disparate productivity tools
- **Workflow automation** across tool boundaries
- **Consistent experience** while leveraging specialized tools
- **Reduced context switching** between different interfaces

**For Vice Development**:
- **Sustainable architecture** - less code to maintain
- **Faster iteration** - delegate complex features to specialized tools
- **Community leverage** - benefit from improvements in integrated tools
- **Clear differentiation** - focus on orchestration and UX, not reimplementation

This repositioning transforms vice from "yet another habit tracker" to a **unique productivity orchestration platform** that makes Unix tools more accessible and powerful through unified workflows.

## Critical Questions & Potential Issues

### 1. Cross-Note Relationship Management
**Issue**: Backlinks and SRS scheduling require understanding relationships between notes

#### What Replaces In-Memory Structs?

**Current T027 Approach**:
```go
// Load everything into memory
notes := repo.LoadFlotsam(ctx)
for _, note := range notes {
    // Complex in-memory operations
    backlinks := computeBacklinks(note, notes)
    srsData := computeSRS(note)
}
```

**Unix Interop Replacement Strategy**:

**1. Bulk Operations → Shell Commands to ZK**
```bash
# Instead of loading all notes into memory
vice flotsam list --format json | jq '.[] | select(.tags[] == "project")'

# Delegate to zk for complex queries
zk list --format json --tag project | vice flotsam srs-filter
```

**2. Single Note Operations → Direct .md Parsing**
```go
// Only when we need rich structs for single notes
func editNote(path string) error {
    note, err := parseFlotsamNote(path)  // Parse just this one
    if err != nil {
        return err
    }
    // Work with single note struct
    return openEditor(path)
}
```

**3. Relationships → External Tools + Caching**
```bash
# Backlink discovery via ripgrep
rg -l "\[\[$(basename "$note")\]\]" --type md

# Or delegate to zk's link analysis
zk list --linked-by "$note" --format path
```

**4. Cross-Note State → Database Only**
```go
// SRS data lives in database, not in-memory structs
func getDueNotes() ([]string, error) {
    return db.Query("SELECT note_path FROM srs_reviews WHERE next_due <= ?", time.Now())
}
```

#### Architecture Comparison

**T027 In-Memory Approach**:
- Load entire collection into `[]FlotsamNote` structs
- Compute relationships in memory
- High memory usage, fast operations
- Complex consistency management

**Unix Interop Approach**:
- **Bulk queries**: Shell out to zk/external tools
- **Single operations**: Parse only needed .md files
- **Relationships**: External tools + simple caching
- **State**: Database for SRS, file system for everything else

#### Specific Replacements

| T027 Operation | Unix Interop Replacement |
|----------------|--------------------------|
| `repo.LoadFlotsam(ctx)` | `zk list --format json` |
| `computeBacklinks(note, allNotes)` | `rg -l "\[\[$note\]\]" --type md` |
| `findRelated(note, allNotes)` | `zk list --linked-by $note` |
| `filterByTag(notes, tag)` | `zk list --tag $tag` |
| `getSRSData(note)` | `sqlite3 srs.db "SELECT * WHERE note_path=..."` |
| `note.Content` | `cat $note.md` or single file parse |

#### When Do We Still Need Structs?

**Single Note Operations** (parse .md directly):
- Note editing/creation
- Template rendering
- Frontmatter manipulation
- Content validation

**Never For**:
- Bulk operations
- Collection-wide queries
- Cross-note relationships
- Multi-note aggregations

This eliminates the "load everything into memory" pattern entirely, replacing it with:
1. **Shell commands** for bulk operations
2. **Single file parsing** for individual note work
3. **Database queries** for SRS state
4. **External tools** for relationships

#### The Combined Search Challenge

**Problem**: zk's powerful filtering can't directly query SRS data, and our SRS database doesn't have zk's rich metadata.

**ZK's Rich Filtering Capabilities**:
- Full-text search with FTS database
- Complex tag queries (`--tag "inbox OR todo"`)
- Link relationship queries (`--linked-by`, `--link-to`)
- Date-based filtering (`--created-after`, `--modified-before`)
- Interactive fuzzy finding (`--interactive`)
- Sorting by various criteria (`--sort created-`)

**SRS Database Contains**:
- `note_path`, `last_reviewed`, `next_due`, `quality`, `interval_days`
- But lacks: title, tags, content, links, creation date

**Combined Query Examples That Would Be Difficult**:
```bash
# Find overdue notes tagged "important" created this week
zk list --tag important --created-after "last monday" | vice flotsam srs-filter --overdue

# Show notes due today that link to a specific concept
zk list --linked-by concepts/spaced-repetition.md | vice flotsam srs-filter --due-today

# Interactive selection of high-quality review notes
zk list --interactive | vice flotsam srs-filter --quality ">3"
```

#### Two Solutions for Combined Search

**Solution 1: Denormalized SRS Database (Recommended)**
Store essential zk metadata in SRS database for efficient joins:

```sql
CREATE TABLE srs_reviews (
    note_path TEXT PRIMARY KEY,
    title TEXT,           -- Denormalized from zk
    tags TEXT,            -- JSON array of tags
    created_date DATE,    -- Denormalized from zk
    modified_date DATE,   -- Denormalized from zk
    -- SRS-specific fields
    last_reviewed DATE,
    next_due DATE,
    quality INTEGER,
    interval_days INTEGER,
    context TEXT
);
```

**Cache Consistency Strategy**:
- **Persistent processes**: Background sync with file system watching (fswatch)
- **CLI invocations**: mtime checking on startup for cache invalidation
- **Lazy refresh**: Update cache entries when stale data is detected

**Advantages**:
- Single database query for combined filters
- Fast performance for complex queries
- Standard SQL for all operations

**Disadvantages**:
- Cache invalidation complexity
- Potential inconsistency between zk and SRS data
- Storage duplication

**Solution 2: Pipeline Composition (Unix Way)**
Chain commands for combined queries:

```bash
# Find notes, then filter by SRS
zk list --tag important --format path | vice flotsam srs-filter --overdue

# Find SRS notes, then filter by zk
vice flotsam due --format path | zk list --include-file - --interactive

# Two-stage filtering
zk list --tag project --created-after "last week" --format json > /tmp/notes.json
vice flotsam srs-query --input /tmp/notes.json --quality ">3"
```

**Advantages**:
- True Unix composition
- No cache consistency issues
- Each tool does what it does best

**Disadvantages**:
- More complex user experience
- Potential performance overhead
- Requires temp files or complex piping

#### Practical Impact Assessment

**Common Use Cases**:
1. **"Show me overdue notes"** - Simple SRS query, no zk needed
2. **"Edit today's review notes"** - SRS query → file paths → editor
3. **"Find high-quality notes on topic X"** - Combined query needed
4. **"Review notes I haven't seen in a while"** - Combined query needed

**Frequency Analysis**:
- **Pure SRS queries**: 70% of use cases (due notes, review scheduling)
- **Pure zk queries**: 20% of use cases (content search, navigation)
- **Combined queries**: 10% of use cases (advanced workflows)

**Recommendation**: 
Use **Solution 1 (Denormalized SRS Database)** because:
- Most practical queries need combined data
- Cache consistency is a solved problem (file watching)
- Performance is critical for daily SRS workflows
- SQL is more powerful than shell pipeline composition

The 10% of advanced combined queries justify the complexity of cache management.

#### The Algorithm Access Tradeoff

**Critical Insight**: Using a denormalized SRS database gives us **data access** but not **algorithm access**.

**What We Gain**:
- Combined search and SRS data in SQL
- Fast queries for complex filters
- Cache consistency through file watching

**What We Lose**:
- zk's sophisticated FTS algorithms
- zk's link analysis algorithms  
- zk's relationship computation
- zk's optimized indexing strategies

**Specific Capabilities We'd Need to Reimplement**:
- **Full-text search**: zk's FTS database with stemming, tokenization
- **Link parsing**: Extract `[[wikilinks]]` and `[markdown](links)`
- **Backlink computation**: Bidirectional relationship mapping
- **Tag extraction**: Parse frontmatter and inline tags
- **Content analysis**: Word counts, mention detection

**Alternative: Hybrid Data Access**
Instead of full denormalization, use **targeted caching** for common queries:

```sql
-- Minimal SRS database (no zk data duplication)
CREATE TABLE srs_reviews (
    note_path TEXT PRIMARY KEY,
    last_reviewed DATE,
    next_due DATE,
    quality INTEGER,
    interval_days INTEGER
);

-- Cache only for performance-critical combined queries
CREATE TABLE search_cache (
    note_path TEXT PRIMARY KEY,
    title TEXT,
    tags TEXT,
    last_indexed DATE
);
```

#### SRS Database Location Strategy

**Database Placement**: Separate from zk's database to avoid conflicts
- **Option A**: `.zk/vice.db` (alongside zk's database)
- **Option B**: `.vice/flotsam.db` (separate vice directory)

**Advantages of separate database**:
- **No schema conflicts**: Independent of zk's database evolution
- **Clear ownership**: Vice manages SRS data, zk manages note metadata
- **Backup isolation**: Can backup/restore SRS data independently
- **Tool independence**: Works even if zk database is corrupted/missing

**Recommended structure**:
```
notebook/
├── .zk/
│   ├── config.toml          # zk configuration
│   ├── notebook.db          # zk's database
│   └── templates/           # note templates
├── .vice/                   # or .zk/vice/
│   ├── flotsam.db          # SRS database
│   └── config.toml         # vice-specific config
└── flotsam/                # actual note files
    ├── concept-1.md
    └── concept-2.md
```

#### Frontmatter Cache Question

**Current assumption**: Store SRS data in note frontmatter as cache
**Alternative**: Database-only approach with no frontmatter

**Frontmatter advantages**:
- Self-contained notes (survives copy/move)
- Git-trackable review history
- Human-readable scheduling info

**Frontmatter disadvantages**:
- Frontmatter pollution
- Consistency complexity (database ↔ file sync)
- Parsing overhead for every note read

**Recommendation**: Start with **database-only** approach
- Simpler consistency model
- No frontmatter pollution
- Can add frontmatter caching later if portability becomes important

The main ADR topic (Unix interop vs coupled integration) is more fundamental than the frontmatter detail.

**Query Strategy**:
- **Pure SRS**: Direct database queries
- **Pure zk**: Shell out to zk commands
- **Combined**: Use cache for hot queries, fall back to pipeline composition

**Example Workflows**:
```bash
# Pure SRS (70% of use cases) - Direct SQL
vice flotsam due --today

# Pure zk (20% of use cases) - Shell out
zk list --tag important --interactive

# Combined (10% of use cases) - Cache + fallback
vice flotsam due --tag important  # Uses cache if available
# Falls back to: vice flotsam due --format path | zk list --include-file -

# Editor integration - delegate to zk entirely
vice flotsam edit --tag project    # Resolves to: zk edit --tag project
vice flotsam edit overdue-notes.md # Resolves to: zk edit overdue-notes.md
```

**Advantages of Hybrid Approach**:
- Leverage zk's algorithms for complex operations
- Minimal cache surface area (less consistency complexity)
- Performance optimization only where needed
- Graceful degradation when cache is stale
- **Delegate editor integration entirely to zk** (no need to reimplement)

**Disadvantages**:
- More complex query planning
- Some operations still require pipeline composition
- Cache miss performance penalty

This approach acknowledges that zk's algorithms are valuable and hard to replicate, while still enabling efficient combined queries for the most common use cases.

#### Using Tags for Note Behaviors

**Insight**: zk's tag system can handle "flags" for note behaviors, making them searchable

**Tag Syntax Options** (all supported by zk):
- `#hashtags` - Simple inline tags
- `:colon:separated:tags:` - Hierarchical organization
- `#multi-word tags#` - Bear-style multi-word tags
- YAML frontmatter - `tags: [vice:srs, vice:task]`

**Vice-specific Tag Patterns**:
```markdown
---
tags: [vice:srs, vice:task, concept, important]
---
# Note with SRS and task behaviors

This note is tracked by SRS and has task-like properties.
```

**Search Integration**:
```bash
# Find all SRS-enabled notes
zk list --tag "vice:srs"

# Find overdue SRS notes (combining zk tags + SRS database)
zk list --tag "vice:srs" --format path | vice flotsam srs-filter --overdue

# Find important SRS notes interactively
zk list --tag "vice:srs AND important" --interactive

# Exclude task notes from general search
zk list --tag "NOT vice:task" --match "productivity"
```

**Benefits of Tag-based Behaviors**:
- **Searchable**: Full integration with zk's powerful tag filtering
- **Discoverable**: `zk tag list` shows all vice-specific tags
- **Composable**: Can combine vice tags with content tags
- **Standard**: Uses zk's existing tag infrastructure
- **Hierarchical**: `:vice:srs:active` vs `:vice:srs:archived`

**Example Tag Hierarchies**:
```
vice:srs           # Basic SRS tracking
vice:srs:active    # Currently being reviewed
vice:srs:suspended # Temporarily disabled
vice:task          # Task-like notes
vice:task:project  # Project-related tasks
vice:daily         # Daily journal integration
```

**Implementation**:
```go
// Check if note has SRS behavior
func hasSRSBehavior(notePath string) bool {
    cmd := exec.Command("zk", "list", "--tag", "vice:srs", "--format", "path", notePath)
    output, _ := cmd.Output()
    return len(output) > 0
}

// Find notes with specific vice behaviors
func findViceNotes(behavior string) ([]string, error) {
    cmd := exec.Command("zk", "list", "--tag", fmt.Sprintf("vice:%s", behavior), "--format", "path")
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    return strings.Split(string(output), "\n"), nil
}
```

This eliminates the need for separate behavior configuration files - everything is discoverable through zk's tag system.

#### ZK Notebook Configuration Management

**Key Insight**: Each zk notebook has its own `.zk/config.toml` that vice should manage

**Notebook Initialization Strategy**:
```go
func initializeNotebook(path string) error {
    // Create .zk directory if it doesn't exist
    zkDir := filepath.Join(path, ".zk")
    os.MkdirAll(zkDir, 0755)
    
    // Write/update .zk/config.toml with vice-specific settings
    config := `
[note]
filename = "{{slug title}}"
extension = "md"

[group.flotsam]
paths = ["flotsam"]

[group.flotsam.note]
filename = "{{slug title}}"
template = "flotsam.md"
default-title = "New Flotsam Note"

[group.journal]
paths = ["journal"]

[group.journal.note]
filename = "{{format-date now '%Y-%m-%d'}}"
template = "journal.md"
`
    
    return writeZkConfig(zkDir, config)
}
```

**Template Management**:
```bash
# vice init creates appropriate templates
.zk/templates/flotsam.md:
---
title: "{{title}}"
tags: [vice:srs, flotsam]
created: {{format-date now}}
---

# {{title}}

Content goes here...

.zk/templates/journal.md:
---
title: "{{format-date now 'long'}}"
tags: [journal, daily]
date: {{format-date now}}
---

# {{format-date now 'long'}}

What happened today?
```

**Benefits**:
- **Consistent setup**: Every vice notebook has proper zk configuration
- **Template management**: Appropriate templates for different note types
- **User customization**: Users can override defaults via vice config
- **Seamless integration**: zk commands work immediately after `vice init`

**Implementation Tasks**:
- **T045: ZK Configuration Management** - Write zk config during notebook init
- **T046: User ZK Customization** - Allow vice config to override zk settings

#### ZK Dependency Management

**Challenge**: zk as a prerequisite complicates installation for many users

**Installation Strategies**:

**1. Flake Dependencies (Nix)**
```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    zk.url = "github:zk-org/zk";
  };
  
  outputs = { self, nixpkgs, zk }: {
    packages.x86_64-linux.default = pkgs.buildGoModule {
      # vice build with zk automatically available
      buildInputs = [ zk.packages.x86_64-linux.default ];
    };
  };
}
```

**2. Git Submodule / "Hostage" Installation**
```bash
# Include zk as submodule, build during vice installation
git submodule add https://github.com/zk-org/zk.git vendor/zk
# Build script compiles zk alongside vice
make build-deps && make build
```

**3. Invokable Shell Script**
```bash
#!/bin/bash
# install-deps.sh - One-command dependency setup
if ! command -v zk &> /dev/null; then
    echo "Installing zk..."
    go install github.com/zk-org/zk@latest
fi
```

**4. Vice Utility Command**
```bash
# Built-in dependency management
vice install-deps    # Installs zk and other dependencies
vice doctor          # Checks for required tools
```

**5. Embedded zk (Go Module Integration)**
```go
// Import zk as library, embed core functionality
import "github.com/zk-org/zk/internal/core"

// Fallback: shell out to zk binary if available, 
// otherwise use embedded functionality
func zkList(args ...string) ([]byte, error) {
    if zkBinary := findZkBinary(); zkBinary != "" {
        return exec.Command(zkBinary, args...).Output()
    }
    // Fallback to embedded zk core
    return embeddedZkList(args...)
}
```

**6. Graceful Degradation**
```go
func checkZkAvailable() bool {
    _, err := exec.LookPath("zk")
    return err == nil
}

func flotsamList(args []string) error {
    if checkZkAvailable() {
        // Full functionality with zk
        return shellOutToZk(args)
    } else {
        // Reduced functionality without zk
        fmt.Println("Enhanced features require zk installation")
        return basicFlotsamList(args)
    }
}
```

**Recommendations**:
- **Primary**: Utility command (`vice install-deps`) for easy setup
- **Advanced**: Nix flake for reproducible environments
- **Fallback**: Graceful degradation with helpful error messages
- **Documentation**: Clear installation instructions with multiple paths

**Benefits of Both Being Go Modules**:
- **Consistent toolchain**: Both use `go install`
- **Version compatibility**: Can specify compatible zk versions
- **Cross-platform**: Go's build system handles platform differences
- **Library integration**: Potential to embed zk functionality if needed

**RTFM Lesson**: This architectural insight only became clear after reading zk's documentation thoroughly. The `external-processing.md` guidance about piping to external apps was key to understanding the interop possibilities.

#### ZK Custom Filename Configuration

**Critical Discovery**: zk supports fully configurable filenames, not just short IDs

**From `config-note.md`**: zk's filename generation is completely customizable through templates:

```toml
[note]
# Custom filename patterns
filename = "{{slug title}}"           # Creates: "my-note-title.md"
filename = "{{format-date now}}"      # Creates: "2025-07-18.md"
filename = "{{id}}-{{slug title}}"    # Creates: "abc123-my-note-title.md"

# Context-specific filenames via groups
[group.flotsam]
paths = ["flotsam"]
[group.flotsam.note]
filename = "{{slug title}}"           # Semantic filenames for flotsam
template = "flotsam.md"

[group.daily]
paths = ["journal/daily"]
[group.daily.note]
filename = "{{format-date now}}"      # Date-based filenames for journal
```

**Impact on Unix Interop Approach**:
- **Human-readable filenames**: Notes can have meaningful names
- **Semantic organization**: Filenames reflect content, not just IDs
- **Grep-friendly**: Can search by filename patterns
- **Editor-friendly**: Easier to navigate in file explorers

**Flotsam-Specific Configuration**:
```toml
# .zk/config.toml
[group.flotsam]
paths = ["flotsam"]

[group.flotsam.note]
filename = "{{slug title}}"
extension = "md"
template = "flotsam.md"

# Template file: .zk/templates/flotsam.md
---
title: "{{title}}"
tags: [vice:srs, flotsam]
created: {{format-date now}}
---

# {{title}}

Content here...
```

**Benefits**:
- **Discoverable filenames**: `concept-of-learning.md` vs `abc123.md`
- **File system browsing**: Easy to find notes by name
- **Git-friendly**: Meaningful commit diffs
- **Backup clarity**: Clear what each file contains

**This removes another assumption** that was limiting the Unix interop approach - we're not stuck with cryptic short IDs as filenames.

#### User Configuration of ZK Behaviors

**Additional Unix Interop Advantage**: Users can configure zk behavior through vice's configuration

**Vice Configuration Integration**:
```toml
# vice config.toml
[flotsam]
zk_flags = ["--no-input", "--quiet"]
zk_editor = "nvim"
zk_fzf_options = "--height 50% --border"

[flotsam.zk_config]
# Options to write to .zk/config.toml
filename_template = "{{slug title}}"
template = "flotsam.md"
default_tags = ["vice:srs", "flotsam"]
```

**Dynamic ZK Command Construction**:
```go
func buildZkCommand(baseCmd string, args []string) *exec.Cmd {
    // Add user-configured flags
    flags := viper.GetStringSlice("flotsam.zk_flags")
    zkArgs := append([]string{baseCmd}, flags...)
    zkArgs = append(zkArgs, args...)
    
    cmd := exec.Command("zk", zkArgs...)
    
    // Set environment variables for zk
    if editor := viper.GetString("flotsam.zk_editor"); editor != "" {
        cmd.Env = append(os.Environ(), "ZK_EDITOR="+editor)
    }
    
    return cmd
}
```

**Benefits**:
- **User customization**: Each user can configure zk behavior via vice
- **Consistent experience**: vice manages zk configuration across notebooks
- **Power user features**: Advanced zk options accessible through vice config
- **Environment management**: vice can set zk environment variables

**Example User Workflows**:
```bash
# User prefers vim with specific options
vice config set flotsam.zk_editor "vim +startinsert"

# User wants custom fzf behavior
vice config set flotsam.zk_fzf_options "--height 80% --preview-window right:50%"

# User wants different filename patterns per context
vice config set flotsam.filename_template "{{format-date now}}-{{slug title}}"
```

**Installation Complexity vs Feature Richness Tradeoff**:
- **T027 approach**: Complex code, no external dependencies
- **Unix interop**: Simple code, zk dependency
- **Hybrid**: Graceful degradation provides middle ground

#### Cache Consistency Implementation

**For CLI Invocations (Typical Usage)**:
```go
func checkCacheConsistency() error {
    // Check mtime of cache database vs source files
    cacheTime := getCacheModTime()
    
    // Quick mtime check on key directories
    if dirModTime("flotsam/") > cacheTime {
        return refreshCache()
    }
    
    // Or check individual files if cache has per-file entries
    cachedFiles := getCachedFilePaths()
    for _, file := range cachedFiles {
        if fileModTime(file) > cache.GetFileTime(file) {
            refreshCacheEntry(file)
        }
    }
    return nil
}
```

**For Persistent Processes (Future TUI/daemon)**:
```go
func watchForChanges() {
    watcher, _ := fsnotify.NewWatcher()
    watcher.Add("flotsam/")
    
    for {
        select {
        case event := <-watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                invalidateCacheEntry(event.Name)
            }
        }
    }
}
```

**Advantages of mtime approach**:
- **No external dependencies**: No fswatch/inotify setup required
- **Startup cost**: One-time check on CLI invocation
- **Simple implementation**: Standard file system operations
- **Cross-platform**: Works on all operating systems

**Performance characteristics**:
- **Cold start**: ~1-5ms for directory mtime checks
- **Warm cache**: Zero overhead once validated
- **Selective refresh**: Only update changed files

**Comparison with T027**:
- **T027**: Always loads everything fresh (high startup cost)
- **Unix interop**: Cache + mtime validation (low startup cost)
- **Result**: Better performance for repeated operations

### 2. SRS Scheduling Complexities
**Issue**: Spaced repetition requires stateful scheduling across time

#### SRS Storage Options Analysis

**Option 1: ZK Notebook Integration (Cribbed from ZK)**
- **Approach**: Extend zk's SQLite database to store SRS metadata
- **Feasibility**: High - zk already has proven SQLite integration patterns
- **Advantages**: 
  - Leverages existing, battle-tested infrastructure
  - Atomic transactions and ACID properties
  - Efficient querying and indexing
  - Familiar to zk users
- **Disadvantages**:
  - Requires zk notebook structure
  - Adds dependency on zk's database schema
  - May conflict with zk's own SRS features if they exist
- **Implementation**: Extend zk's `metadata_dao.go` patterns for SRS fields

**Option 2: Separate SRS Database**
- **Approach**: Dedicated SQLite/other database for SRS data only
- **Feasibility**: High - standard database operations
- **Advantages**:
  - Independent of zk notebook structure
  - Optimized schema for SRS use cases
  - No conflicts with zk functionality
  - Easy backup and migration
- **Disadvantages**:
  - Additional database to manage
  - Synchronization challenges with file system
  - Potential consistency issues
- **Implementation**: Simple table: `(note_path, last_reviewed, next_due, quality, interval)`

**Option 3: Frontmatter Embedding**
- **Approach**: Store SRS data in YAML frontmatter of each note
- **Feasibility**: Medium - requires consistent frontmatter management
- **Advantages**:
  - Self-contained notes (no external dependencies)
  - Git-trackable SRS history
  - Human-readable and editable
  - Survives note moves/copies
- **Disadvantages**:
  - Requires parsing every note for queries
  - Potential frontmatter corruption
  - No atomic batch operations
  - Performance issues with large collections
- **Implementation**: Extend existing frontmatter parsing

**Option 4: External SRS Tool Integration**
- **Approach**: Delegate to existing SRS tools (Anki, Mnemosyne, etc.)
- **Feasibility**: Low-Medium - depends on tool APIs
- **Advantages**:
  - Leverages specialized SRS algorithms
  - Mature scheduling implementations
  - Rich ecosystem of SRS tools
- **Disadvantages**:
  - Complex integration and data synchronization
  - Tool-specific dependencies
  - Limited control over scheduling logic
- **Implementation**: Export/import bridges to external tools

**Option 5: Git-Based State Tracking**
- **Approach**: Use git history and timestamps for SRS calculations
- **Feasibility**: Medium - requires git repository
- **Advantages**:
  - Leverages existing git infrastructure
  - Natural history tracking
  - Distributed and backed up
- **Disadvantages**:
  - Complex scheduling algorithm implementation
  - Requires git repository
  - Performance issues with large histories
- **Implementation**: Analyze git log for review patterns

#### Comparison with T027 "Strangler" Approach

**T027 Current Approach**:
- **Storage**: In-memory structs with file system persistence
- **Consistency**: Single-process consistency only
- **Performance**: Fast in-memory operations
- **Complexity**: High coupling with repository layer

**Feasibility Comparison**:

| Aspect | T027 Strangler | ZK Integration | Separate DB | Frontmatter | External Tools |
|--------|---------------|----------------|-------------|-------------|----------------|
| **Implementation Complexity** | High | Medium | Low | Low | High |
| **Consistency Guarantees** | Limited | High | High | Low | Variable |
| **Performance** | High | High | High | Low | Variable |
| **Data Portability** | Low | Medium | Medium | High | Low |
| **Maintenance Burden** | High | Medium | Low | Low | High |
| **Unix Interop Friendly** | No | Yes | Yes | Yes | Variable |

#### Recommended Approach for Unix Interop

**Hybrid: Separate SRS Database + Frontmatter Cache**
- **Primary Storage**: Lightweight SQLite database for SRS metadata
- **Cache Layer**: Essential SRS data in frontmatter for standalone operation
- **Synchronization**: Background process to sync database ↔ frontmatter

```sql
-- SRS Database Schema
CREATE TABLE srs_reviews (
    note_path TEXT PRIMARY KEY,
    last_reviewed DATE,
    next_due DATE,
    quality INTEGER,
    interval_days INTEGER,
    ease_factor REAL,
    context TEXT
);
```

```yaml
# Frontmatter Cache
---
srs:
  next_due: 2025-07-25
  quality: 4
  interval: 7
---
```

**CLI Integration**:
```bash
# Query operations use database for performance
vice flotsam due --format json
vice flotsam review <note> --quality 4

# Standalone notes work via frontmatter
cp note.md /tmp/  # Still contains SRS data
```

**Advantages over T027**:
- **Simpler Architecture**: Database operations instead of complex repository patterns
- **Unix Friendly**: Standard database tools and SQL queries
- **Portable**: Notes contain essential SRS data
- **Consistent**: ACID properties from SQLite
- **Testable**: Easy to mock database operations

This approach is **more feasible** than T027's strangler pattern because:
1. **Simpler Mental Model**: Database operations vs complex repository abstractions
2. **Standard Tools**: SQL instead of custom query languages
3. **Better Separation**: SRS concerns isolated from note management
4. **Easier Testing**: Database mocking vs repository mocking

### 3. Performance & Scalability Concerns
**Issue**: Process spawning and file parsing overhead
- **Large Collections**: How to handle thousands of notes efficiently?
- **Repeated Operations**: Is re-parsing metadata on each command acceptable?
- **Search Performance**: How to provide fast full-text and metadata search?
- **Memory Usage**: How to avoid loading entire collection into memory?

*Possible Solutions*:
- Incremental indexing strategies
- Metadata caching in separate files
- Database-backed metadata store
- Lazy loading patterns

### 4. Atomic Operations & Consistency
**Issue**: Multi-step operations across process boundaries
- **Note Creation**: How to ensure template + metadata + file creation is atomic?
- **Bulk Updates**: How to handle partial failures in batch operations?
- **Concurrent Access**: How to prevent corruption when multiple processes operate?
- **Transaction Semantics**: How to rollback failed multi-step operations?

*Possible Solutions*:
- File locking mechanisms
- Temporary file + atomic rename patterns
- Process coordination via lock files
- Git-based versioning and rollback

### 5. Configuration & Environment Management
**Issue**: Flotsam-specific behavior and settings
- **Context Configuration**: How to maintain different settings per context?
- **Tool Discovery**: How to find and configure external tools (fzf, editors)?
- **Template Management**: How to handle note templates and generation?
- **Path Resolution**: How to resolve relative paths across different invocation contexts?

*Possible Solutions*:
- Environment variable conventions
- Configuration file hierarchy
- Tool auto-detection patterns
- Notebook-relative path resolution

### 6. Error Handling & User Experience
**Issue**: Debugging across process boundaries
- **Error Attribution**: Which tool/command caused a failure?
- **Partial Results**: How to handle operations that partially succeed?
- **User Feedback**: How to provide meaningful error messages?
- **Recovery Guidance**: How to help users fix problems?

*Possible Solutions*:
- Structured error output (JSON)
- Verbose/debug modes
- Error code conventions
- Self-diagnostic commands

### 7. Testing & Development Complexity
**Issue**: External dependencies and integration testing
- **Tool Availability**: How to test when external tools are missing?
- **Process Mocking**: How to mock external process interactions?
- **Integration Testing**: How to test complete workflows reliably?
- **CI/CD Complexity**: How to ensure consistent test environments?

*Possible Solutions*:
- Test doubles for external tools
- Container-based testing
- Graceful degradation modes
- Minimal external dependencies

### 8. Rich UI Integration Challenges
**Issue**: Building future TUI/GUI interfaces
- **API Surface**: How to expose flotsam operations to rich UIs?
- **Real-time Updates**: How to show live updates without polling?
- **Performance**: Can UI interactions be responsive with process spawning?
- **State Synchronization**: How to keep UI state consistent with file system?

*Possible Solutions*:
- Hybrid approach: CLI for automation, Go APIs for UI
- File watching for real-time updates
- Local caching layers
- Background process coordination

### 9. Use Case Coverage Analysis
**Issue**: Ensuring all T027 functionality remains available
- **Repository Operations**: Can all CRUD operations be expressed via CLI?
- **Complex Queries**: How to handle multi-criteria searches and joins?
- **Batch Processing**: How to efficiently process large result sets?
- **Workflow Integration**: How to maintain seamless editor/tool integration?

*Possible Solutions*:
- Rich query language for CLI
- Streaming result processing
- Plugin/extension architecture
- Workflow orchestration commands

## Impact Assessment

### If Unix Interop is Chosen
- **Simplification**: Significant reduction in vice's codebase
- **Flexibility**: Users can integrate with any external tool
- **Maintenance**: Less code to maintain, fewer internal dependencies
- **Learning**: Users familiar with Unix tools get immediate value

### If Coupled Integration is Retained
- **Performance**: Faster operations, no process spawning overhead
- **Rich APIs**: Internal Go APIs for future UI development
- **Control**: Full control over user experience and error handling
- **Complexity**: More code to maintain and test

## References
- T027: Flotsam data layer implementation
- zk documentation: Unix interop patterns
- Unix Philosophy: "Do one thing and do it well"
- Command-line interface design best practices