# vice - CLI Habit Tracker

A command-line habit tracker that supports flexible habit types and stores data in human-readable YAML files.

## Features

- **Simple Habits**: Boolean pass/fail tracking (did you meditate today?)
- **Elastic Habits**: Multi-level achievement tracking with mini/midi/maxi levels
- **Informational Habits**: Data collection without pass/fail scoring
- **Automatic Scoring**: Habits can be automatically scored based on defined criteria
- **Context Management**: Separate personal/work contexts with isolated data
- **XDG Compliance**: Follows Unix filesystem conventions with proper directory structure
- **Local Storage**: All data stored in local YAML files for version control and portability
- **Interactive CLI**: User-friendly forms with field-specific input validation

## Installation

```bash
# Build from source
git clone <repository>
cd vice
go build -o vice .

# Install to PATH
go install .
```

## Quick Start

1. **Initialize configuration**:
   ```bash
   vice entry
   ```
   This creates sample configuration and data files in XDG directories on first run.

2. **Record today's habits**:
   ```bash
   vice entry
   ```
   Answer the interactive prompts to record your progress.

3. **Use context switching**:
   ```bash
   # Switch context persistently
   vice context switch work
   
   # Use temporary context for one command
   vice --context personal entry
   
   # Use environment variable
   VICE_CONTEXT=work vice entry
   ```

4. **Use custom directories**:
   ```bash
   vice --config-dir /path/to/config --data-dir /path/to/data entry
   ```

## Configuration

vice uses XDG Base Directory specification for configuration and data storage:

### Configuration Files
- **Application config**: `config.toml` in `$XDG_CONFIG_HOME/vice/` (default: `~/.config/vice/`)
- **Context state**: `vice.yml` in `$XDG_STATE_HOME/vice/` (default: `~/.local/state/vice/`)

### Data Files (per context)
- **Habit definitions**: `habits.yml` in `$XDG_DATA_HOME/vice/{context}/` 
- **Daily entries**: `entries.yml` in `$XDG_DATA_HOME/vice/{context}/`
- **Checklists**: `checklists.yml` and `checklist_entries.yml` in `$XDG_DATA_HOME/vice/{context}/`

Default data location: `~/.local/share/vice/{context}/` (where context is "personal" or "work" by default)

### Context Management

Contexts allow complete isolation of habit data for different life areas:

```toml
# config.toml
[core]
contexts = ["personal", "work"]  # Define available contexts
```

Each context maintains completely separate data files, enabling users to track personal habits separately from work habits.

## Habit Types

### Simple Habits

Boolean habits with pass/fail tracking:

```yaml
version: "1.0.0"
habits:
  - title: "Morning Exercise"
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Did you exercise this morning?"
    help_text: "Any movement counts - stretching, walking, gym, sports, etc."
```

### Elastic Habits

Multi-level achievement habits with mini/midi/maxi levels:

```yaml
habits:
  - title: "Exercise Duration"
    habit_type: "elastic"
    field_type:
      type: "duration"
    scoring_type: "automatic"
    prompt: "How long did you exercise today?"
    help_text: "Enter duration like: 30m, 1h15m, or 1:30:00"
    mini_criteria:
      condition:
        greater_than_or_equal: 15  # 15 minutes = mini achievement
    midi_criteria:
      condition:
        greater_than_or_equal: 30  # 30 minutes = midi achievement
    maxi_criteria:
      condition:
        greater_than_or_equal: 60  # 60 minutes = maxi achievement
```

### Informational Habits

Data collection without scoring:

```yaml
habits:
  - title: "Sleep Quality"
    habit_type: "informational"
    field_type:
      type: "unsigned_int"
      unit: "rating"
      min: 1
      max: 10
    prompt: "Rate your sleep quality (1-10):"
```

## Field Types

| Type | Description | Example Input |
|------|-------------|---------------|
| `boolean` | Yes/no questions | true, false, yes, no |
| `unsigned_int` | Positive integers | 5, 10, 100 |
| `unsigned_decimal` | Positive decimals | 2.5, 10.75 |
| `decimal` | Any decimal number | -1.5, 0, 3.14 |
| `duration` | Time duration | 30m, 1h15m, 1:30:00 |
| `time` | Time of day | 14:30, 09:00 |
| `text` | Free-form text | Any string |

### Field Type Options

```yaml
field_type:
  type: "unsigned_int"
  unit: "glasses"        # Display unit (optional)
  min: 1                 # Minimum value (optional)
  max: 20                # Maximum value (optional)
```

## Scoring Criteria

### Simple Comparisons

```yaml
condition:
  greater_than: 10
  greater_than_or_equal: 15
  less_than: 100
  less_than_or_equal: 50
```

### Range Constraints

```yaml
condition:
  range:
    min: 10
    max: 100
    min_inclusive: true   # default: true
    max_inclusive: false  # default: false
```

### Time Constraints

```yaml
condition:
  before: "10:00"  # Before 10 AM
  after: "06:00"   # After 6 AM
```

### Boolean Matching

```yaml
condition:
  equals: true
```

### Logical Operators

```yaml
condition:
  and:
    - greater_than_or_equal: 30
    - less_than: 120
  or:
    - equals: true
    - greater_than: 50
  not:
    less_than: 10
```

## Habit Schema Structure

Each habit supports these fields:

```yaml
title: "Habit Title"                    # Required: Human-readable name
id: "habit_id"                         # Optional: auto-generated from title
description: "Habit description"        # Optional: markdown supported
habit_type: "simple|elastic|informational"  # Required
field_type:                           # Required: see field types above
  type: "boolean"
scoring_type: "manual|automatic"      # Required for simple/elastic habits
prompt: "Custom prompt text"          # Optional: CLI prompt
help_text: "Additional guidance"      # Optional: help text
```

### Elastic Habit Specific Fields

```yaml
mini_criteria:      # Required for automatic scoring
  description: "Minimum achievement"
  condition:
    greater_than_or_equal: 15
midi_criteria:      # Required for automatic scoring
  condition:
    greater_than_or_equal: 30  
maxi_criteria:      # Required for automatic scoring
  condition:
    greater_than_or_equal: 60
```

## Example Complete Configuration

```yaml
version: "1.0.0"
habits:
  # Simple boolean habit
  - title: "Morning Meditation"
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Did you meditate this morning?"

  # Elastic habit with automatic scoring
  - title: "Exercise Duration"
    habit_type: "elastic"
    field_type:
      type: "duration"
    scoring_type: "automatic"
    prompt: "How long did you exercise today?"
    help_text: "Enter duration like: 30m, 1h15m, or 1:30:00"
    mini_criteria:
      condition:
        greater_than_or_equal: 15
    midi_criteria:
      condition:
        greater_than_or_equal: 30
    maxi_criteria:
      condition:
        greater_than_or_equal: 60

  # Numeric habit with units
  - title: "Water Intake"
    habit_type: "elastic"
    field_type:
      type: "unsigned_int"
      unit: "glasses"
    scoring_type: "automatic"
    prompt: "How many glasses of water did you drink?"
    mini_criteria:
      condition:
        greater_than_or_equal: 4
    midi_criteria:
      condition:
        greater_than_or_equal: 6
    maxi_criteria:
      condition:
        greater_than_or_equal: 8

  # Informational data collection
  - title: "Sleep Quality"
    habit_type: "informational"
    field_type:
      type: "unsigned_int"
      unit: "rating"
      min: 1
      max: 10
    prompt: "Rate your sleep quality (1-10):"
```

## Commands

### `vice entry`

Record today's habit completion. Presents an interactive form for each defined habit.

**Examples:**
```bash
vice entry                              # Use current context
vice --context work entry              # Use work context temporarily
VICE_CONTEXT=personal vice entry       # Use environment variable
```

### `vice context`

Manage contexts for data isolation.

**Examples:**
```bash
vice context list                       # List available contexts
vice context show                       # Show current context
vice context switch work               # Switch to work context (persistent)
```

### Global Options

All commands support these global flags:

```bash
--config-dir PATH      # Override $XDG_CONFIG_HOME/vice
--data-dir PATH        # Override $XDG_DATA_HOME/vice  
--state-dir PATH       # Override $XDG_STATE_HOME/vice
--cache-dir PATH       # Override $XDG_CACHE_HOME/vice
--context NAME         # Use context temporarily (no state change)
```

**Examples:**
```bash
vice entry                                              # Use defaults
vice --config-dir ~/custom entry                       # Custom config
vice --data-dir /work/habits --context work entry      # Custom data + context
```

## Data Storage

### Directory Structure

```
XDG directories:
├── ~/.config/vice/
│   └── config.toml              # Application configuration
├── ~/.local/state/vice/
│   └── vice.yml                 # Active context state
├── ~/.local/share/vice/
│   ├── personal/                # Personal context data
│   │   ├── habits.yml          # Habit definitions
│   │   ├── entries.yml         # Daily entries
│   │   ├── checklists.yml      # Checklist templates
│   │   └── checklist_entries.yml # Checklist completions
│   └── work/                   # Work context data
│       ├── habits.yml          # Separate habit definitions
│       ├── entries.yml         # Separate daily entries
│       ├── checklists.yml      # Separate checklists
│       └── checklist_entries.yml # Separate completions
└── ~/.cache/vice/              # Future: performance caching
```

### Context Isolation

Each context maintains completely separate data:
- **Personal habits** tracked in `~/.local/share/vice/personal/`
- **Work habits** tracked in `~/.local/share/vice/work/`
- **No data bleeding** between contexts
- **Context switching** changes which data files are active

### Entries File Format

Daily progress entries in each context:

```yaml
version: "1.0.0"
entries:
  - date: "2024-01-15"
    habits:
      - habit_id: "morning_exercise"
        value: true
        completed_at: "2024-01-15T07:30:00Z"
        notes: "Great workout!"
      - habit_id: "exercise_duration"
        value: "45m"
        achievement_level: "midi"
        completed_at: "2024-01-15T07:30:00Z"
```

## Specifications

For complete technical details, see:

- [Habit Schema Specification](doc/specifications/habit_schema.md) - YAML structure and validation
- [File Paths & Runtime Environment](doc/specifications/file_paths_runtime_env.md) - XDG compliance and context management

## Development

```bash
# Run tests
go test ./...

# Format code  
gofumpt -w .

# Lint code
golangci-lint run

# Build
go build .
```