# UI libraries: Bubbletea & ecosystem

**AIDEV-NOTE**: UI libraries with example code are checked out as git submodules at [[/charmbracelet]]. Consult these for idiomatic patterns and usage examples.
**AIDEV-NOTE** collect lessons learned, recommended practices, concise directives and resources here**

Vice uses many charmbracelet libraries (bubbletea, huh, lipgloss & bubbles, x). 
- Bubbletea: UI platform based on Elm
- Huh: forms
- Lipgloss: styling
- Bubbles: community widgets.
- X: extensions; contains teatest (testing).

This file contains a summary of guidance for building bubbletea apps effectively.

## Rules of thumb

*provide sources!*

- from [elm](https://guide.elm-lang.org/webapps/structure)
  - Preferring shorter files: considered harmful
  - Don't split `Model` `Update` and `View` into multiple files
  - Build all functionality around a type 
  - Don't actively try to make "components" (encapsulate all functionality into modules) 

- from [hn](https://news.ycombinator.com/item?id=41369065)
  - consider storing your data the Elm way into one big, flat Model. Elm
    webapps turn into headaches when you split state among submodels.


## Example applications

Perhaps the devs knew something
- https://github.com/leg100/pug

## Potentially useful libraries not yet used

- https://github.com/rmhubbert/bubbletea-overlay : an easy way to create modal windows
- https://github.com/NimbleMarkets/ntcharts : report widgets

## Significant Files & Patterns

Files & patterns in this project worth knowing about:

### Examples to follow

**Modal System Architecture** (`internal/ui/modal/`):
- Clean separation of modal lifecycle and parent state management
- Proper use of BubbleTea command pattern for deferred operations
- Modal interface design for polymorphic modal types

**Deferred State Synchronization** (`internal/ui/entrymenu/model.go:482-547`):
- Custom message types for timing-sensitive operations
- Command pattern to defer heavy operations until next BubbleTea cycle
- Prevents timing conflicts between UI state changes and data processing

### Known issues / technical debt

**Modal Lifecycle Timing** (Resolved in T024):
- Critical: Always get data from modal BEFORE nulling reference
- Pattern: `result := modal.GetData(); modal = nil; return ProcessData(result)`
- Anti-pattern: `modal = nil; return ProcessData(modal.GetData())` (nil pointer!)

**State Synchronization Complexity**:
- Entry menu state sync between file storage, EntryCollector, and UI display
- Multiple type conversions (interface{} → HabitEntry) can hide bugs
- Consider simplifying the state management chain

### BubbleTea Patterns & Best Practices

**1. Modal Data Flow Pattern**
```go
// CORRECT: Get data before nulling reference
if modal.IsClosed() {
    result := modal.GetData()
    modal = nil
    return ProcessData(result)
}

// INCORRECT: Null reference before getting data  
if modal.IsClosed() {
    modal = nil
    return ProcessData(modal.GetData()) // nil pointer!
}
```

**2. Deferred Command Pattern**
Use for operations that must happen after current BubbleTea cycle:
```go
// Return command to defer operation
return tea.Cmd(func() tea.Msg {
    return CustomMsg{data: complexData}
})

// Handle in Update method
case CustomMsg:
    m.processComplexOperation(msg.data)
    return m, nil
```

**3. Debug Infrastructure Requirements**
- Comprehensive logging for UI lifecycle events
- Clear categorization: [MODAL], [ENTRYMENU], [FIELD], etc.
- Conditional debug logging that can be easily disabled
- Centralized debug system (see `internal/debug/logger.go`)

**4. Investigation Methodology for Complex UI Bugs**
- Build incremental prototypes to isolate problems
- Systematic elimination of architectural layers
- Granular testing (individual vs. combined operations)
- Comprehensive debug logging with clear categories

# Sources / further reading

the elm architecture, inspiration for bubbletea and likely a good proxy for useful approaches
- https://guide.elm-lang.org/architecture/

online docs:
- [huh API reference](https://pkg.go.dev/github.com/charmbracelet/huh) - Complete API documentation
- [bubbletea API reference](https://pkg.go.dev/github.com/charmbracelet/bubbletea) - complete API documentation
- [bubbletea with huh reference example](https://github.com/charmbracelet/huh/blob/main/examples/bubbletea/main.go) - idiomatic example of huh + bubbletea

guidance:
- https://leg100.github.io/en/posts/building-bubbletea-programs/ - some tips. TODO: summarise here
- https://news.ycombinator.com/item?id=41369065 - nerds parading opinions, may contain gems of wisdom
