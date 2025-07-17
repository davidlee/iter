# Flotsam Package Documentation

## Overview

The `flotsam` package provides a comprehensive data layer for the flotsam note system, implementing a files-first architecture with ZK-compatible parsing and spaced repetition system (SRS) functionality. This package combines components from the [ZK note-taking system](https://github.com/zk-org/zk) and [go-srs](https://github.com/revelaction/go-srs) to create a unified system for managing notes with spaced repetition learning.

## Architecture

### Files-First Design

The flotsam system uses individual markdown files with YAML frontmatter as the source of truth, with optional SQLite performance caching. This design ensures:

- **Data Portability**: All data travels with markdown files
- **ZK Compatibility**: Works seamlessly with existing ZK notebooks
- **Performance**: SQLite cache for fast SRS queries
- **Reliability**: Source files are always recoverable

### Component Integration

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           FLOTSAM PACKAGE COMPONENTS                           │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐              │
│  │   ZK Components │    │ go-srs Components│    │ Integration     │              │
│  │                 │    │                 │    │ Components      │              │
│  │ • ID Generation │    │ • SM-2 Algorithm│    │ • Data Models   │              │
│  │ • Frontmatter   │    │ • SRS Interfaces│    │ • Serialization │              │
│  │ • Link Parsing  │    │ • Review System │    │ • Validation    │              │
│  │ • AST Processing│    │ • Quality Scale │    │ • Integration   │              │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘              │
│           │                       │                       │                     │
│           ▼                       ▼                       ▼                     │
│  ┌─────────────────────────────────────────────────────────────────┐            │
│  │                    UNIFIED API LAYER                            │            │
│  │                                                                 │            │
│  │  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐       │            │
│  │  │ Note Creation │  │ Link Extraction│  │ SRS Processing│       │            │
│  │  │ & Parsing     │  │ & Resolution   │  │ & Scheduling  │       │            │
│  │  └───────────────┘  └───────────────┘  └───────────────┘       │            │
│  └─────────────────────────────────────────────────────────────────┘            │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Core Data Structures

### FlotsamNote

The primary data structure representing a complete flotsam note:

```go
type FlotsamNote struct {
    // Core note data
    ID       string    `yaml:"id" json:"id"`
    Title    string    `yaml:"title" json:"title"`
    Type     string    `yaml:"type" json:"type"`         // idea, flashcard, script, log
    Tags     []string  `yaml:"tags" json:"tags"`
    Created  time.Time `yaml:"created-at" json:"created-at"`
    Modified time.Time `yaml:"-" json:"-"`               // File modification time
    
    // Content
    Body      string   `yaml:"-" json:"-"`               // Markdown body content
    Links     []string `yaml:"-" json:"-"`               // Extracted [[wikilinks]]
    Backlinks []string `yaml:"-" json:"-"`               // Computed reverse links
    FilePath  string   `yaml:"-" json:"-"`               // Absolute file path
    
    // SRS data (optional)
    SRS *SRSData `yaml:"srs,omitempty" json:"srs,omitempty"`
}
```

### SRSData

Spaced repetition data stored in frontmatter:

```go
type SRSData struct {
    // Easiness factor (default 2.5, minimum 1.3)
    Easiness float64 `yaml:"easiness" json:"easiness"`
    // Number of consecutive correct answers
    ConsecutiveCorrect int `yaml:"consecutive_correct" json:"consecutive_correct"`
    // Unix timestamp when the card is due for review
    Due int64 `yaml:"due" json:"due"`
    // Total number of reviews performed
    TotalReviews int `yaml:"total_reviews" json:"total_reviews"`
    // Review history for debugging/analysis
    ReviewHistory []ReviewRecord `yaml:"review_history,omitempty" json:"review_history,omitempty"`
}
```

### Link

Represents a link found in note content:

```go
type Link struct {
    Title        string        // Display text
    Href         string        // Target reference
    Type         LinkType      // Wikilink, Markdown, or Implicit
    IsExternal   bool          // Whether link points outside the note collection
    Rels         []LinkRelation // Relationship types (up/down)
    Snippet      string        // Surrounding context
    SnippetStart int          // Context start position
    SnippetEnd   int          // Context end position
}
```

## API Reference

### ID Generation

#### `NewIDGenerator(opts IDOptions) func() string`

Creates a new ID generator function using the specified options.

```go
generator := NewIDGenerator(IDOptions{
    Length:  4,
    Case:    CaseLower,
    Charset: CharsetAlphanum,
})
noteID := generator() // Returns "a1b2"
```

#### `NewFlotsamIDGenerator() IDGenerator`

Creates an ID generator with flotsam-specific defaults (4-char alphanum lowercase, ZK-compatible).

```go
generator := NewFlotsamIDGenerator()
noteID := generator() // Returns ZK-compatible ID like "3k9z"
```

### Frontmatter Parsing

#### `parseFrontmatter(content string) (map[string]interface{}, string, error)`

Parses YAML frontmatter from markdown content.

```go
frontmatter, body, err := parseFrontmatter(noteContent)
if err != nil {
    return err
}
// frontmatter contains parsed YAML data
// body contains markdown content without frontmatter
```

### Link Extraction

#### `ExtractLinks(content string) []Link`

Extracts all links from markdown content using goldmark AST parsing.

```go
links := ExtractLinks(markdownContent)
for _, link := range links {
    fmt.Printf("Link: %s -> %s (type: %v)\n", link.Title, link.Href, link.Type)
}
```

#### `ExtractWikiLinkTargets(content string) []string`

Extracts only the target hrefs from wikilinks for backlink processing.

```go
targets := ExtractWikiLinkTargets(content)
// Returns ["target1", "target2"] for [[target1]] and [[target2]]
```

#### `BuildBacklinkIndex(notes map[string]string) map[string][]string`

Builds a map of note targets to their source notes for backlink computation.

```go
backlinks := BuildBacklinkIndex(noteContents)
// backlinks["target"] = ["source1", "source2"]
```

### SRS Processing

#### `NewSM2Calculator() *SM2Calculator`

Creates a new SM-2 algorithm calculator.

```go
calc := NewSM2Calculator()
```

#### `ProcessReview(oldData *SRSData, quality Quality) (*SRSData, error)`

Updates SRS data based on a review session.

```go
updatedSRS, err := calc.ProcessReview(currentSRS, CorrectHard)
if err != nil {
    return err
}
// updatedSRS contains new scheduling information
```

#### `IsDue(data *SRSData) bool`

Checks if a card is due for review.

```go
if calc.IsDue(note.SRS) {
    // Present note for review
}
```

### Quality Scale

The SRS system uses a 0-6 quality scale:

```go
const (
    NoReview          Quality = 0  // No review performed
    IncorrectBlackout Quality = 1  // Total failure to recall
    IncorrectFamiliar Quality = 2  // Incorrect but familiar
    IncorrectEasy     Quality = 3  // Incorrect but seemed easy
    CorrectHard       Quality = 4  // Correct with difficulty
    CorrectEffort     Quality = 5  // Correct with some effort
    CorrectEasy       Quality = 6  // Perfect recall
)
```

## Usage Examples

### Basic Note Creation and Processing

```go
// Generate a unique ID
generator := NewFlotsamIDGenerator()
noteID := generator()

// Create note content
noteContent := fmt.Sprintf(`---
id: %s
title: My First Note
created-at: %s
tags: [example, test]
---

# My First Note

This note links to [[another note]] and demonstrates SRS functionality.
`, noteID, time.Now().Format(time.RFC3339))

// Parse the note
frontmatter, body, err := parseFrontmatter(noteContent)
if err != nil {
    log.Fatal(err)
}

// Extract links
links := ExtractLinks(body)
fmt.Printf("Found %d links\n", len(links))

// Create flotsam note structure
note := &FlotsamNote{
    ID:    frontmatter["id"].(string),
    Title: frontmatter["title"].(string),
    Body:  body,
    Links: make([]string, len(links)),
}

for i, link := range links {
    note.Links[i] = link.Href
}
```

### SRS Integration Workflow

```go
// Initialize SRS for a note
srsData := &SRSData{
    Easiness:           2.5,
    ConsecutiveCorrect: 0,
    Due:                time.Now().Unix(),
    TotalReviews:       0,
    ReviewHistory:      []ReviewRecord{},
}

note.SRS = srsData

// Check if due for review
calc := NewSM2Calculator()
if calc.IsDue(note.SRS) {
    // Present note for review
    fmt.Printf("Review note: %s\n", note.Title)
    
    // Process review result
    quality := CorrectHard // User's assessment
    updatedSRS, err := calc.ProcessReview(note.SRS, quality)
    if err != nil {
        log.Fatal(err)
    }
    
    // Update note with new SRS data
    note.SRS = updatedSRS
    
    // Save back to file (implementation depends on storage layer)
}
```

### Link Processing and Backlinks

```go
// Collection of notes for backlink processing
noteContents := map[string]string{
    "note1": "This links to [[note2]] and [[note3]]",
    "note2": "This links back to [[note1]]",
    "note3": "Standalone note",
}

// Build backlink index
backlinks := BuildBacklinkIndex(noteContents)

// backlinks["note1"] = ["note2"]
// backlinks["note2"] = ["note1"]
// backlinks["note3"] = ["note1"]

// Extract specific link types
for noteID, content := range noteContents {
    links := ExtractLinks(content)
    for _, link := range links {
        if link.Type == LinkTypeWikiLink && !link.IsExternal {
            fmt.Printf("%s -> %s\n", noteID, link.Href)
        }
    }
}
```

## Performance Considerations

### Search Operations

When processing large collections of notes:

- **Batch Processing**: Process multiple notes in a single operation
- **Link Extraction**: Goldmark AST parsing is efficient but cache results for repeated operations
- **Frontmatter Parsing**: YAML parsing is lightweight but consider caching for frequently accessed notes

Performance benchmark: ~19µs per note for complete processing (ID generation + frontmatter parsing + link extraction + SRS processing).

### Bulk SRS Processing

For processing many due cards:

- **Batch Calculations**: Process multiple reviews in a single transaction
- **Due Date Queries**: Use efficient date range queries when implementing SQLite cache
- **Memory Management**: Process cards in chunks to avoid memory pressure

### Directory Scanning

When scanning large note collections:

- **Incremental Processing**: Only process changed files using timestamp + checksum
- **Parallel Processing**: Parse notes concurrently where possible
- **Cache Invalidation**: Implement efficient cache invalidation strategies

### Cache Synchronization

For SQLite performance cache:

- **Atomic Operations**: Ensure file writes and cache updates are atomic
- **Change Detection**: Use timestamp + SHA256 checksum for efficient change detection
- **Recovery**: Implement cache rebuild from source files on corruption

## Integration with Vice Architecture

### Repository Pattern Integration

The flotsam package is designed to integrate with Vice's Repository Pattern (T028):

```go
// Extend DataRepository interface
type DataRepository interface {
    // Existing methods...
    
    // Flotsam methods
    LoadFlotsam(ctx string) (*FlotsamCollection, error)
    SaveFlotsam(ctx string, flotsam *FlotsamCollection) error
    CreateFlotsamNote(ctx string, flotsam *FlotsamNote) error
    GetFlotsamNote(ctx string, id string) (*FlotsamNote, error)
    UpdateFlotsamNote(ctx string, flotsam *FlotsamNote) error
    DeleteFlotsamNote(ctx string, id string) error
    SearchFlotsam(ctx string, query string) ([]*FlotsamNote, error)
}
```

### Context Isolation

All flotsam operations respect Vice's context system:

- **Directory Structure**: `$VICE_DATA/{context}/flotsam/`
- **Link Resolution**: Wikilinks resolved within context boundaries
- **Cache Isolation**: Separate cache databases per context
- **Backlink Computation**: Computed within context scope

### ZK Notebook Integration

For ZK notebook compatibility:

- **Hybrid Architecture**: Separate metadata stores prevent conflicts
- **Directory Bridge**: Vice operates on ZK notebooks without modification
- **Frontmatter Extensions**: ZK preserves Vice-specific fields in metadata
- **Database Coexistence**: Vice tables added to ZK database are ignored by ZK

## Error Handling

The package provides comprehensive error handling:

```go
// Quality validation
if err := quality.Validate(); err != nil {
    // Handle invalid quality rating
}

// SRS processing errors
updatedSRS, err := calc.ProcessReview(srsData, quality)
if err != nil {
    // Handle SRS calculation errors
}

// Link extraction errors are handled internally
links := ExtractLinks(content) // Returns empty slice on error
```

## Testing

The package includes comprehensive test coverage:

- **Unit Tests**: Individual component testing
- **Integration Tests**: Cross-component workflow testing
- **Performance Tests**: Benchmarking for optimization
- **Compatibility Tests**: ZK interoperability validation

Integration test coverage includes:
- Complete note lifecycle (creation → parsing → linking → SRS → serialization)
- Cross-component data flow validation
- Performance benchmarking (19µs per note average)
- Round-trip data integrity verification

## License and Attribution

This package incorporates code from two external projects:

### ZK Components (GPLv3)
- **Source**: https://github.com/zk-org/zk
- **License**: GNU General Public License v3.0
- **Components**: Frontmatter parsing, link extraction, ID generation
- **Files**: `zk_parser.go`, `zk_links.go`, `zk_id.go`

### go-srs Components (Apache-2.0)
- **Source**: https://github.com/revelaction/go-srs
- **License**: Apache License 2.0
- **Components**: SM-2 algorithm, SRS interfaces, review system
- **Files**: `srs_sm2.go`, `srs_interfaces.go`, `srs_review.go`

All components are properly attributed with copyright headers and license compliance documentation.

## Future Enhancements

### Planned Features
- **Template System**: Note creation templates using handlebars
- **Advanced Scheduling**: Alternative SRS algorithms (SM-18, FSRS)
- **Bulk Operations**: Batch processing for large note collections
- **Query Language**: Advanced search and filtering capabilities

### Performance Optimizations
- **Streaming Processing**: Process large note collections without loading all into memory
- **Incremental Updates**: More efficient change detection and processing
- **Cache Optimization**: Advanced caching strategies for frequently accessed data
- **Parallel Processing**: Concurrent note processing for bulk operations

### Integration Enhancements
- **Task Management**: Integration with Vice's task-oriented workflows
- **Incremental Writing**: Support for progressive note development
- **Context Switching**: Efficient context migration and synchronization
- **Export/Import**: Data portability between different note systems