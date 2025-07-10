  # Goal Schema Format Specification

  ## Overview

This specification defines the syntax and data structures for the habit tracker
goal schema. The schema enables:

1. Validation: Ensuring schema correctness and entry compatibility
2. Parser Implementation: Structured data extraction from schema files
3. CLI Generation: Dynamic prompt creation for user entry
4. Entry Validation: Checking daily log entries against schema requirements

## Design Principles

- Resilience: Schema changes shouldn't break historical data
- Flexibility: Support diverse goal types and data formats
- Clarity: Human-readable format with clear validation rules
- Extensibility: Easy to add new field types and criteria

## File Format

Schema files use YAML format for human readability while maintaining structured
parsing capabilities.

# Data Type Specifications

## Base Field Types

### Text field for free-form comments

```  
  text:
    type: "text"
    multiline: boolean (default: false)
```

### Boolean field for yes/no questions  

``` 
  boolean:
    type: "boolean"
    default: boolean (optional)
``` 

### Numeric fields with units

``` 
  numeric:
    type: "unsigned_int" | "unsigned_decimal" | "decimal"
    unit: string (e.g., "kg", "hours", "count")
    min?: number (optional constraint)
    max?: number (optional constraint)

  # Time of day in HH:MM format
  time:
    type: "time"
    format: "HH:MM" (24-hour format)

  # Duration in multiple formats
  duration:
    type: "duration"
    format: "HH:MM:SS" | "minutes" | "seconds"
``` 

## Validation Rules

- text: Any string, newlines allowed if multiline=true
- boolean: Accepts true/false, yes/no, y/n, 1/0 (case-insensitive)
- numeric: Must be valid number of specified type, within min/max if specified
- time: Must match HH:MM format, 00:00-23:59 range
- duration: Must match specified format, non-negative values

## Goal Type Specifications

### Simple Goals

Boolean pass/fail goals with manual or automatic scoring.

``` 
  goal_type: "simple"
  scoring_type: "manual" | "automatic"
  criteria: # Required if scoring_type="automatic"
    # Criteria specification (see below)
``` 

### Elastic Goals

Three-tier achievement goals (mini/midi/maxi) with manual or automatic scoring.

``` 
  goal_type: "elastic"
  scoring_type: "manual" | "automatic"
  mini_criteria: # Required if scoring_type="automatic"
    # Criteria for minimum achievement level
  midi_criteria: # Required if scoring_type="automatic"  
    # Criteria for medium achievement level
  maxi_criteria: # Required if scoring_type="automatic"
    # Criteria for maximum achievement level
``` 

### Informational Goals

Data collection without success/failure scoring.

``` 
  goal_type: "informational"
  direction: "higher_better" | "lower_better" | "neutral" # Optional, for display
``` 

## Criteria Specification

### Numeric/Duration Criteria

Simple comparisons

```
  greater_than: number
  greater_than_or_equal: number
  less_than: number
  less_than_or_equal: number
```

Range constraints
```
  range:
    min: number
    max: number
    min_inclusive: boolean (default: true)
    max_inclusive: boolean (default: false)
```

Periodicity (requires historical data)
```
  periodicity:
    count: integer
    operator: "at_least" | "at_most"
    period: integer
    unit: "days" | "weeks" | "months"
```

### Time Criteria

Time constraints (HH:MM format)
```
  before: "HH:MM"
  after: "HH:MM"
```

Boolean Criteria
```
  equals: true | false
```

### Composite Criteria

Logical operators
  and:
    - # Array of criteria (all must pass)
  or:
    - # Array of criteria (any must pass)
  not:
    # Single criteria to negate

## Schema Structure

### Top-Level Schema

```
  version: "1.0.0" # Semantic version
  created_date: "2024-01-01" # ISO8601 date
  goals:
    - # Array of Goal objects
```

## Goal Object

```
  title: "Goal Title" # Human-readable name
  id: "goal_id" # Optional unique identifier (auto-generated if missing)
  position: 1 # Unique integer for display order
  description: | # Optional markdown description
    Multi-line description
    supports **markdown**
  goal_type: "simple" | "elastic" | "informational"
  field_type:
    # Field type specification (see above)
  scoring_type: "manual" | "automatic" # Required for simple/elastic
  criteria: # Required for automatic scoring
    description: "Optional criteria description"
    condition:
      # Criteria condition (see above)
  # Elastic-specific fields
  mini_criteria: # Elastic goals only
  midi_criteria: # Elastic goals only  
  maxi_criteria: # Elastic goals only
  # Informational-specific fields
  direction: "higher_better" | "lower_better" | "neutral" # Informational only
  prompt: "Enter your value:" # CLI prompt text
  help_text: "Optional additional guidance" # Optional
```

## Identifier System

### ID Generation

- If id is omitted, generate from title: lowercase, replace spaces/special chars
  with underscores
- Ensure uniqueness within schema
- Example: "Daily Exercise" → "daily_exercise"

### Change Resilience

- Entry validation matches fields by goal ID
- Position changes don't affect historical data
- Missing goals in entries are preserved as "orphaned"
- Schema validation warns about orphaned fields

## Validation Requirements

### Schema Validation

1. Structure: Valid YAML matching specification
2. Uniqueness: All goal IDs and positions must be unique
3. Completeness: Required fields present based on goal_type and scoring_type
4. Consistency: Field types compatible with criteria
5. References: All criteria reference valid field types

### Entry Validation

1. Field Matching: Entry fields match goal IDs in schema
2. Type Checking: Values conform to field type specifications
3. Criteria Evaluation: Automatic scoring based on criteria
4. Orphan Detection: Flag fields without matching goals
5. Historical Context: Periodicity criteria require entry history

## Example Schema

```
  version: "1.0.0"
  created_date: "2024-01-01"
  goals:
    - title: "Daily Exercise"
      id: "daily_exercise"
      position: 1
      description: "Track daily physical activity"
      goal_type: "elastic"
      field_type:
        type: "duration"
        format: "minutes"
      scoring_type: "automatic"
      mini_criteria:
        description: "Minimum 15 minutes"
        condition:
          greater_than_or_equal: 15
      midi_criteria:
        description: "Target 30 minutes"
        condition:
          greater_than_or_equal: 30
      maxi_criteria:
        description: "Excellent 60+ minutes"
        condition:
          greater_than_or_equal: 60
      prompt: "How many minutes did you exercise today?"
      help_text: "Include any physical activity: walking, gym, sports, etc."

    - title: "Morning Meditation"
      position: 2
      goal_type: "simple"
      field_type:
        type: "boolean"
      scoring_type: "manual"
      prompt: "Did you meditate this morning?"

    - title: "Sleep Quality"
      position: 3
      goal_type: "informational"
      field_type:
        type: "unsigned_int"
        unit: "rating"
        min: 1
        max: 10
      direction: "higher_better"
      prompt: "Rate your sleep quality (1-10):"
```

This specification provides the foundation for implementing a robust validator
and parser that can handle the complexity of the habit tracking system while
maintaining resilience to schema changes.