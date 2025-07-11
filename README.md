# iter - CLI Habit Tracker

A command-line habit tracker that supports flexible goal types and stores data in human-readable YAML files.

## Features

- **Simple Goals**: Boolean pass/fail tracking (did you meditate today?)
- **Elastic Goals**: Multi-level achievement tracking with mini/midi/maxi levels
- **Informational Goals**: Data collection without pass/fail scoring
- **Automatic Scoring**: Goals can be automatically scored based on defined criteria
- **Local Storage**: All data stored in local YAML files for version control and portability
- **Interactive CLI**: User-friendly forms with field-specific input validation

## Installation

```bash
# Build from source
git clone <repository>
cd iter
go build -o iter .

# Install to PATH
go install .
```

## Quick Start

1. **Initialize configuration**:
   ```bash
   iter entry
   ```
   This creates sample configuration files in `~/.config/iter/` on first run.

2. **Record today's habits**:
   ```bash
   iter entry
   ```
   Answer the interactive prompts to record your progress.

3. **Use custom config directory**:
   ```bash
   iter --config-dir /path/to/config entry
   ```

## Configuration

iter stores configuration in two files:

- `goals.yml` - defines your habit goals and criteria
- `entries.yml` - stores your daily progress entries

Default location: `~/.config/iter/` (follows XDG Base Directory specification)

## Goal Types

### Simple Goals

Boolean goals with pass/fail tracking:

```yaml
version: "1.0.0"
goals:
  - title: "Morning Exercise"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Did you exercise this morning?"
    help_text: "Any movement counts - stretching, walking, gym, sports, etc."
```

### Elastic Goals

Multi-level achievement goals with mini/midi/maxi levels:

```yaml
goals:
  - title: "Exercise Duration"
    goal_type: "elastic"
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

### Informational Goals

Data collection without scoring:

```yaml
goals:
  - title: "Sleep Quality"
    goal_type: "informational"
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

## Goal Schema Structure

Each goal supports these fields:

```yaml
title: "Goal Title"                    # Required: Human-readable name
id: "goal_id"                         # Optional: auto-generated from title
description: "Goal description"        # Optional: markdown supported
goal_type: "simple|elastic|informational"  # Required
field_type:                           # Required: see field types above
  type: "boolean"
scoring_type: "manual|automatic"      # Required for simple/elastic goals
prompt: "Custom prompt text"          # Optional: CLI prompt
help_text: "Additional guidance"      # Optional: help text
```

### Elastic Goal Specific Fields

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
goals:
  # Simple boolean goal
  - title: "Morning Meditation"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Did you meditate this morning?"

  # Elastic goal with automatic scoring
  - title: "Exercise Duration"
    goal_type: "elastic"
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

  # Numeric goal with units
  - title: "Water Intake"
    goal_type: "elastic"
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
    goal_type: "informational"
    field_type:
      type: "unsigned_int"
      unit: "rating"
      min: 1
      max: 10
    prompt: "Rate your sleep quality (1-10):"
```

## Commands

### `iter entry`

Record today's habit completion. Presents an interactive form for each defined goal.

**Options:**
- `--config-dir PATH` - Use custom configuration directory

**Examples:**
```bash
iter entry                           # Use default config directory
iter --config-dir ~/habits entry    # Use custom config directory
```

## Data Storage

### Goals File (`goals.yml`)

Contains your goal definitions and scoring criteria.

### Entries File (`entries.yml`)

Stores daily progress entries:

```yaml
version: "1.0.0"
entries:
  - date: "2024-01-15"
    goals:
      - goal_id: "morning_exercise"
        value: true
        completed_at: "2024-01-15T07:30:00Z"
        notes: "Great workout!"
      - goal_id: "exercise_duration"
        value: "45m"
        achievement_level: "midi"
        completed_at: "2024-01-15T07:30:00Z"
```

## Specification

For complete technical details, see [Goal Schema Specification](doc/specifications/goal_schema.md).

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