# ZK CLI Interop Specification

## Overview

This specification defines how vice integrates with the [zk](https://github.com/zk-org/zk) CLI tool for Unix interop patterns. Vice delegates complex note operations to zk while maintaining SRS scheduling functionality through its own SQLite database.

## Reference

### ZK Documentation
- **ZK Source**: `zk/` - Local zk source code and documentation
- **ZK CLI Help**: `zk --help`, `zk <command> --help` for command reference
- **ZK Docs**: `zk/docs/` - Comprehensive documentation including:
  - `zk/docs/notes/` - Note management, formatting, frontmatter
  - `zk/docs/config/` - Configuration, tools, aliases
  - `zk/docs/tips/` - Integration patterns, automation, external processing
- **Online**: [zk-org.github.io/zk](https://zk-org.github.io/zk)
- **Repository**: [github.com/zk-org/zk](https://github.com/zk-org/zk)

### Related Vice Specifications
- `doc/specifications/flotsam.md` - Flotsam package and Unix interop architecture
- `doc/specifications/file_paths_runtime_env.md` - Vice directory structure and environment
- `doc/decisions/ADR-002-flotsam-files-first-architecture.md` - Original flotsam decisions


## Design Principles

### Tool Specialization
- **zk**: Handles note content, search, linking, editing, metadata management
- **vice**: Handles SRS scheduling, review workflows, habit integration
- **Clean Separation**: `.zk/` for zk data, `.vice/` for vice data

### Command Patterns
- **Composable**: Commands designed for Unix pipe composition
- **Structured Output**: JSON and template formats for programmatic consumption  
- **Error Handling**: Clear distinction between missing tools, command failures, parsing errors
- **Configuration**: Respect user's editor preferences and zk configuration

## ZK Command Reference

### Core Operations

#### Note Listing and Filtering
```bash
# List all notes with vice:srs tag
zk list --tag "vice:srs" --format json

# List flashcard notes (complex tag queries)
zk list --tag "vice:srs,vice:type:flashcard" --format json

# List notes by title/content pattern
zk list --match "concept" --format json

# List notes in specific directory
zk list concepts/ --format json

# List recent notes  
zk list --created-after "1 week ago" --format json

# Interactive selection with fzf
zk list --tag "vice:srs" --interactive

# Limit results
zk list --tag "vice:srs" --limit 10 --format json
```

#### Link Analysis
```bash
# Get notes that link to target note
zk list --linked-by "path/to/note.md" --format path

# Get notes that target note links to
zk list --link-to "path/to/note.md" --format path

# Get all links (both directions)
zk list --linked-by "path/to/note.md" --link-to "path/to/note.md" --format json
```

#### Note Editing
```bash
# Edit single note
zk edit "path/to/note.md"

# Edit multiple notes (interactive selection)
zk edit --tag "vice:type:flashcard" --interactive

# Edit notes matching criteria
zk list --tag "vice:srs" --format path | xargs zk edit
```

#### Note Creation
```bash
# Create new note with title
zk new --title "My Note"

# Create note in specific directory
zk new concepts/ --title "New Concept"

# Create note with template and extra variables
zk new --title "My Note" --template flotsam --extra 'tags=vice:srs,vice:type:flashcard'

# Create note with custom ID
zk new --title "My Note" --id "abc1"

# Dry run to see generated content
zk new --title "My Note" --dry-run

# Print path instead of editing
zk new --title "My Note" --print-path
```

#### Metadata Operations
```bash
# Show note metadata
zk show "path/to/note.md" --format json

# Update note tags (via zk edit or direct manipulation)
zk edit "path/to/note.md"
```

### Output Formats

zk supports multiple predefined formats: `oneline`, `short`, `medium`, `long`, `full`, `json`, `jsonl`

#### JSON Format
```json
{
  "filename": "concept-1.md",
  "file-path": "/path/to/concept-1.md", 
  "title": "My Concept Note",
  "lead": "Brief excerpt...",
  "body": "Full note content...",
  "raw-content": "---\nid: abc1\n...",
  "word-count": 150,
  "tags": ["vice:srs", "vice:type:flashcard", "concept"],
  "metadata": {
    "id": "abc1",
    "created-at": "2025-07-18T10:30:00Z"
  },
  "created": "2025-07-18T10:30:00Z",
  "modified": "2025-07-18T11:15:00Z"
}
```

#### Path Format
```
/path/to/concept-1.md
/path/to/concept-2.md
/path/to/flashcard-1.md
```

#### Template Format
```
{{title}} - {{path}} ({{word-count}} words)
{{lead}}
```

## Vice Integration Patterns

### Command Mapping

| Vice Operation | ZK Command | Post-processing |
|----------------|------------|-----------------|
| `vice flotsam list` | `zk list --tag "vice:srs" --format json` | Merge with SRS due dates |
| `vice flotsam due` | Direct SRS DB query | Filter by zk tag validation |
| `vice flotsam edit <note>` | `zk edit <path>` | Path resolution |
| `vice flotsam search <term>` | `zk list --match <term> --tag "vice:srs"` | Format results |
| `vice flotsam links <note>` | `zk list --linked-by <path> --format json` | Format backlinks |
| `vice flotsam new <title>` | `zk new --title <title> --template flotsam` | Add to SRS DB |

### Composition Patterns

#### Interactive Review Workflow
```bash
# Get overdue flashcards, review with vice
zk list --tag "vice:srs AND vice:type:flashcard" --format path | \
  vice flotsam due --stdin --overdue --interactive
```

#### Batch Operations
```bash
# Edit all script notes
zk list --tag "vice:type:script" --format path | \
  xargs zk edit

# Review all concept notes  
zk list --tag "vice:type:idea" --format path | \
  vice flotsam review --stdin --batch
```

#### Search and Filter
```bash
# Find and edit notes matching criteria
zk list --match-title "algorithm*" --tag "vice:srs" --format path | \
  head -5 | \
  xargs zk edit
```

## Go Integration Layer

### Tool Abstraction Interface

```go
// Tool represents an external CLI tool for delegation
type Tool interface {
    Name() string
    IsAvailable() bool
    Execute(ctx context.Context, args ...string) (*ToolResult, error)
    Version() (string, error)
}

// ToolResult encapsulates command execution results
type ToolResult struct {
    Stdout   []byte
    Stderr   []byte
    ExitCode int
    Duration time.Duration
}

// ZKTool implements Tool interface for zk operations
type ZKTool struct {
    path       string
    workingDir string
    env        []string
}
```

### Command Builder Pattern

```go
// ZKCommand builds zk commands fluently
type ZKCommand struct {
    tool *ZKTool
    args []string
}

// Example usage
cmd := zk.List().
    Tag("vice:srs AND vice:type:flashcard").
    Format("json").
    CreatedAfter("1 week ago")

result, err := cmd.Execute(ctx)
```

### Output Parsing

```go
// ParseZKListJSON parses zk list output in JSON format
func ParseZKListJSON(data []byte) ([]ZKNote, error) {
    var notes []ZKNote
    if err := json.Unmarshal(data, &notes); err != nil {
        return nil, fmt.Errorf("failed to parse zk JSON output: %w", err)
    }
    return notes, nil
}

// ParseZKPaths parses zk output in path format
func ParseZKPaths(data []byte) ([]string, error) {
    content := strings.TrimSpace(string(data))
    if content == "" {
        return []string{}, nil
    }
    return strings.Split(content, "\n"), nil
}
```

## Error Handling Strategy

### Error Classification

```go
const (
    ErrToolNotFound      ToolErrorType = "tool_not_found"
    ErrToolNotExecutable ToolErrorType = "tool_not_executable"  
    ErrCommandFailed     ToolErrorType = "command_failed"
    ErrInvalidOutput     ToolErrorType = "invalid_output"
    ErrPermissionDenied  ToolErrorType = "permission_denied"
)

type ToolError struct {
    Type    ToolErrorType
    Tool    string
    Command []string
    Message string
    Cause   error
    ExitCode int
}
```

### Error Messages

```go
// User-friendly error messages with actionable guidance
var errorMessages = map[ToolErrorType]string{
    ErrToolNotFound: `zk command not found. Please install zk:
    
    # Using Homebrew (macOS/Linux)
    brew install zk
    
    # Or download from: https://github.com/zk-org/zk/releases
    
    After installation, ensure zk is in your PATH.`,
    
    ErrCommandFailed: `zk command failed. Check that you're in a zk notebook directory.
    
    Initialize a zk notebook with:
    zk init`,
}
```

### Graceful Degradation

```go
// Fallback strategies when zk is unavailable
func (s *FlotsamService) SearchNotes(query string) ([]FlotsamNote, error) {
    // Try zk first for rich search capabilities
    if s.zk.IsAvailable() {
        return s.searchViaZK(query)
    }
    
    // Fallback to in-memory collection search
    log.Warn("zk not available, using fallback search")
    collection, err := s.LoadAllNotes()
    if err != nil {
        return nil, err
    }
    return collection.SearchByTitle(query), nil
}
```

## Directory and Environment Handling

### Notebook Directory Management

ZK relies on notebook discovery and configuration. Vice must ensure consistent behavior:

```go
// Always specify notebook directory explicitly
tool := NewZKTool()
tool.SetNotebookDir("/path/to/vice/context/flotsam")

// This ensures all commands use the correct notebook
result, err := tool.List().Tag("vice:srs").Execute(ctx)
```

### Global Flags Strategy

Vice applies global flags consistently across all zk commands:

```bash
# Every zk command gets these flags
zk <command> --notebook-dir=/path/to/flotsam --no-input [command-args...]
```

**Global Flags Applied**:
- `--notebook-dir=<path>` - Ensure correct notebook (overrides ZK env vars)
- `--working-dir=<path>` - Set working directory if needed  
- `--no-input` - Prevent interactive prompts in automated contexts

### Environment Variable Handling

ZK respects these environment variables that vice may need to manage:
- `ZK_NOTEBOOK_DIR` - Default notebook directory
- `ZK_EDITOR` - Editor for note editing (falls back to VISUAL, EDITOR)
- `VISUAL`, `EDITOR` - Standard editor environment variables

## Configuration Integration

### ZK Configuration Respect

```go
// Respect user's zk configuration
type ZKConfig struct {
    NotebookDir string // From zk config or .zk detection
    Editor      string // From ZK_EDITOR, VISUAL, or EDITOR
    Shell       string // From SHELL env var
    Format      string // Default output format preference
}

// Load configuration from zk and environment
func LoadZKConfig(contextDir string) (*ZKConfig, error) {
    config := &ZKConfig{
        NotebookDir: filepath.Join(contextDir, "flotsam"),
        Editor:      getEditor(),
        Shell:       getShell(), 
        Format:      "json",
    }
    
    // Override with zk-specific config if available
    if zkConfig, err := parseZKConfig(config.NotebookDir); err == nil {
        mergeZKConfig(config, zkConfig)
    }
    
    return config, nil
}
```

### Vice Configuration Extension

```go
// Vice-specific zk integration settings
type ZKIntegration struct {
    Enabled          bool     `toml:"enabled"`
    Path             string   `toml:"path"`              // Custom zk binary path
    DefaultFormat    string   `toml:"default_format"`    // json, path, template
    TagPrefix        string   `toml:"tag_prefix"`        // "vice:" prefix for tags
    TemplateDir      string   `toml:"template_dir"`      // Custom template location
    TimeoutSeconds   int      `toml:"timeout_seconds"`   // Command timeout
    RetryAttempts    int      `toml:"retry_attempts"`    // Retry failed commands
    FallbackMode     string   `toml:"fallback_mode"`     // "memory", "none"
}
```

## Performance Considerations

### Command Caching

```go
// Cache frequently used command results
type CommandCache struct {
    mu    sync.RWMutex
    cache map[string]CachedResult
    ttl   time.Duration
}

type CachedResult struct {
    Data      []byte
    ExpiresAt time.Time
}

// Use for expensive operations like full note listing
func (zk *ZKTool) ListAllNotes(useCache bool) ([]ZKNote, error) {
    if useCache {
        if cached := zk.cache.Get("list_all_notes"); cached != nil {
            return ParseZKListJSON(cached.Data)
        }
    }
    
    result, err := zk.List().Format("json").Execute(context.Background())
    if err != nil {
        return nil, err
    }
    
    if useCache {
        zk.cache.Set("list_all_notes", result.Stdout, 5*time.Minute)
    }
    
    return ParseZKListJSON(result.Stdout)
}
```

### Batch Operations

```go
// Batch multiple operations to reduce zk process spawning
func (zk *ZKTool) BatchEdit(notePaths []string) error {
    // Single zk edit command with multiple paths
    args := append([]string{"edit"}, notePaths...)
    _, err := zk.Execute(context.Background(), args...)
    return err
}

// Pipeline operations for efficiency
func (s *FlotsamService) GetDueNotesForReview() ([]FlotsamNote, error) {
    // Get due note paths from SRS database (fast)
    duePaths, err := s.srsDB.GetDueNotePaths()
    if err != nil {
        return nil, err
    }
    
    // Batch verify notes still have vice:srs tag (zk validation)
    validPaths, err := s.zk.FilterNotesByTag(duePaths, "vice:srs")
    if err != nil {
        return nil, err
    }
    
    // Batch load note metadata
    return s.zk.LoadNotesByPaths(validPaths)
}
```

### Resource Management

```go
// Context-aware execution with proper cleanup
func (zk *ZKTool) Execute(ctx context.Context, args ...string) (*ToolResult, error) {
    cmd := exec.CommandContext(ctx, zk.path, args...)
    cmd.Dir = zk.workingDir
    cmd.Env = zk.env
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    start := time.Now()
    err := cmd.Run()
    duration := time.Since(start)
    
    result := &ToolResult{
        Stdout:   stdout.Bytes(),
        Stderr:   stderr.Bytes(),
        ExitCode: cmd.ProcessState.ExitCode(),
        Duration: duration,
    }
    
    if err != nil {
        return result, NewToolError(ErrCommandFailed, zk.Name(), args, err, result.ExitCode)
    }
    
    return result, nil
}
```

## Testing Strategy

### Mock Implementation

```go
// MockZKTool for testing without external zk dependency
type MockZKTool struct {
    responses map[string]*ToolResult
    available bool
}

func (m *MockZKTool) Execute(ctx context.Context, args ...string) (*ToolResult, error) {
    key := strings.Join(args, " ")
    if result, exists := m.responses[key]; exists {
        return result, nil
    }
    return nil, NewToolError(ErrCommandFailed, "zk", args, nil, 1)
}

// Test helper to setup mock responses
func (m *MockZKTool) SetResponse(command string, stdout []byte, stderr []byte, exitCode int) {
    m.responses[command] = &ToolResult{
        Stdout:   stdout,
        Stderr:   stderr,
        ExitCode: exitCode,
        Duration: time.Millisecond * 10,
    }
}
```

### Integration Tests

```go
func TestZKIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Verify zk is available for integration tests
    zk := NewZKTool()
    if !zk.IsAvailable() {
        t.Skip("zk not available for integration tests")
    }
    
    // Test real zk commands
    notes, err := zk.List().Tag("vice:srs").Format("json").Execute(context.Background())
    require.NoError(t, err)
    
    parsed, err := ParseZKListJSON(notes.Stdout)
    require.NoError(t, err)
    assert.IsType(t, []ZKNote{}, parsed)
}
```

## Security Considerations

### Command Injection Prevention

```go
// Sanitize user input to prevent command injection
func sanitizeZKArg(arg string) string {
    // Remove potentially dangerous characters
    arg = strings.ReplaceAll(arg, ";", "")
    arg = strings.ReplaceAll(arg, "&", "")
    arg = strings.ReplaceAll(arg, "|", "")
    arg = strings.ReplaceAll(arg, "`", "")
    arg = strings.ReplaceAll(arg, "$", "")
    return arg
}

// Use exec.Command with separate arguments (not shell execution)
func (zk *ZKTool) Execute(ctx context.Context, args ...string) (*ToolResult, error) {
    // exec.Command prevents shell injection by not using shell
    cmd := exec.CommandContext(ctx, zk.path, args...)
    // ... rest of implementation
}
```

### Path Validation

```go
// Validate paths are within notebook directory
func (zk *ZKTool) validateNotePath(path string) error {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return err
    }
    
    notebookAbs, err := filepath.Abs(zk.workingDir)
    if err != nil {
        return err
    }
    
    if !strings.HasPrefix(absPath, notebookAbs) {
        return fmt.Errorf("path %s outside notebook directory %s", path, notebookAbs)
    }
    
    return nil
}
```

## Migration from T027

### Function Mapping

| T027 Function | ZK Equivalent | Notes |
|---------------|---------------|-------|
| `LoadFlotsam()` | `zk list --format json` | Full collection loading |
| `SearchNotes(query)` | `zk list --match query` | Fuzzy search delegation |
| `GetBacklinks(note)` | `zk list --linked-by path` | Link analysis delegation |
| `EditNote(note)` | `zk edit path` | Editor delegation |
| `CreateNote(template)` | `zk new --template name` | Template-based creation |

### Preserved Components

- **SRS Database**: SQLite operations remain unchanged
- **FlotsamNote Structs**: Simplified but core structure preserved  
- **File I/O**: Basic read/write operations for SRS data
- **Cache Management**: mtime-based validation for performance

### Deprecated Components

- **In-memory Collections**: Use only as fallback when zk unavailable
- **Backlink Computation**: Delegate entirely to zk
- **Complex Repository Patterns**: Replace with simple tool delegation
- **Coupled Models**: Simplify to minimal data structures

## Future Extensions

### Tool Orchestration Framework

```go
// Generic tool interface for future extensions
type ProductivityTool interface {
    Tool
    Domain() string // "notes", "tasks", "calendar", "time"
    Capabilities() []string // ["search", "edit", "create", "link"]
}

// Support for additional tools
type RemindTool struct {} // Calendar and recurring tasks
type TaskWarriorTool struct {} // GTD task management  
type ObsidianTool struct {} // Alternative note system
```

### Workflow Engine

```go
// Cross-tool workflow coordination
type Workflow struct {
    Name  string
    Steps []WorkflowStep
}

type WorkflowStep struct {
    Tool    string
    Command string
    Args    []string
    Output  string // Pass to next step
}

// Example: "Review overdue flashcards and schedule follow-up"
workflow := Workflow{
    Name: "flashcard_review_followup",
    Steps: []WorkflowStep{
        {Tool: "vice", Command: "flotsam", Args: []string{"due", "--overdue"}},
        {Tool: "zk", Command: "edit", Args: []string{"--stdin"}},
        {Tool: "remind", Command: "add", Args: []string{"follow-up in 3 days"}},
    },
}
```

## Conclusion

This specification establishes vice as a Unix tool orchestrator that leverages zk's strengths while adding sophisticated SRS functionality. The design emphasizes:

- **Tool Specialization**: Each tool excels in its domain
- **Clean Integration**: Minimal coupling between tools
- **Graceful Degradation**: Useful functionality when tools unavailable
- **Future Extensibility**: Framework for additional productivity tools

The implementation prioritizes simplicity, performance, and maintainability while providing a foundation for advanced productivity workflows.