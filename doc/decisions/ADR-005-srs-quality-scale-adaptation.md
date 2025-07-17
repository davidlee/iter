# ADR-005: SRS Quality Scale Adaptation

**Status**: Accepted

**Date**: 2025-07-17

## Related Reading

**Related ADRs**: 
 - [ADR-003: ZK-go-srs Integration Strategy](/doc/decisions/ADR-003-zk-gosrs-integration-strategy.md) - Component integration requiring quality scale choice
 - [ADR-004: SQLite Cache Strategy](/doc/decisions/ADR-004-flotsam-sqlite-cache-strategy.md) - Cache implementation storing quality ratings
 - (Future) ADR-006: Context Isolation Model - How quality ratings interact with context boundaries

**Related Specifications**: 
 - [Flotsam Package Documentation](/doc/specifications/flotsam.md) - Quality scale API reference and usage examples

**Related Tasks**: 
 - [T027/1.2] - go-srs component integration requiring quality scale decision
 - [T027/1.3.3] - Integration testing validating quality scale implementation

## Context

The flotsam SRS system requires users to self-assess their recall performance to drive the SM-2 spaced repetition algorithm. The quality scale choice significantly impacts both user experience and algorithmic effectiveness.

### Quality Scale Requirements
1. **Algorithmic Compatibility**: Must work effectively with SM-2 algorithm calculations
2. **User Experience**: Should be intuitive and quick for users to assess
3. **Research Backing**: Should be based on established spaced repetition research
4. **Granularity**: Must provide sufficient detail for meaningful scheduling adjustments
5. **Consistency**: Should align with broader SRS ecosystem practices

### Available Options Evaluated

#### Option A: Anki 1-4 Scale (Again/Hard/Good/Easy)
- **Pros**: Widely used, simple, familiar to many users
- **Cons**: Limited granularity, designed for Anki's modified SM-2, not pure SM-2
- **Research**: Modified from original SuperMemo research
- **User Experience**: Very simple but potentially lacks nuance

#### Option B: go-srs 0-6 Scale (Research-Based)
- **Pros**: Based on original SuperMemo research, designed for pure SM-2, good granularity
- **Cons**: Less familiar to users, requires explanation
- **Research**: Directly based on Piotr Wozniak's SuperMemo research
- **User Experience**: More detailed assessment but requires learning

#### Option C: Custom Vice Scale (1-5 or similar)
- **Pros**: Could optimize for Vice's specific use cases and user base
- **Cons**: No research backing, algorithm would need adjustment, ecosystem fragmentation
- **Research**: None - would require developing our own
- **User Experience**: Could be tailored but lacks proven effectiveness

#### Option D: Multiple Scale Support
- **Pros**: Flexibility for different user preferences
- **Cons**: Complexity in implementation, algorithm confusion, testing burden
- **Research**: Mixed - different scales have different research backing
- **User Experience**: Choice paralysis, inconsistent ecosystem

### Quality Scale Analysis

#### go-srs 0-6 Scale (Detailed)
```
0 = No Review      - Review not performed (skip/postpone)
1 = Blackout       - Complete failure to recall (total memory blank)
2 = Incorrect      - Wrong answer but familiar when shown correct answer
3 = Incorrect Easy - Wrong answer but seemed easy upon seeing correct answer
4 = Correct Hard   - Correct answer but required significant effort/time
5 = Correct        - Correct answer with some hesitation or effort
6 = Perfect        - Immediate perfect recall with confidence
```

#### Algorithmic Implications
- **SM-2 Compatibility**: Scale designed specifically for original SM-2 algorithm
- **Easiness Calculation**: go-srs implements proven easiness factor adjustments
- **Threshold Logic**: Quality ≥4 considered "correct" for consecutive count
- **Interval Growth**: Different qualities produce different interval multipliers

#### User Experience Considerations
- **Learning Curve**: Users need to understand 7-point scale initially
- **Decision Speed**: Once learned, can be quickly assessed
- **Cognitive Load**: More options but clearer distinctions than binary choices
- **Consistency**: Research-backed distinctions help users be more consistent

## Decision

**We adopt the go-srs 0-6 Quality Scale** with enhanced user experience through clear documentation, examples, and progressive disclosure.

### Rationale:

#### 1. Research Foundation
- **SuperMemo Compatibility**: Based on original Piotr Wozniak research that developed SM-2
- **Proven Effectiveness**: Scale designed specifically for the algorithm we're using
- **Mathematical Basis**: Quality ratings directly map to easiness factor calculations
- **Long-term Success**: Scale has been validated through decades of SuperMemo usage

#### 2. Algorithmic Optimality
- **Pure SM-2**: Our implementation uses original SM-2, not Anki's modifications
- **Easiness Calculation**: go-srs easiness formula designed for this scale
- **Interval Scheduling**: Quality thresholds optimized for this rating system
- **Statistical Validity**: Larger scale provides better data for algorithm adjustments

#### 3. Integration Benefits
- **Component Consistency**: Maintains consistency with copied go-srs components
- **Testing Alignment**: Integration tests already validate this scale
- **Code Simplicity**: No translation layer needed between user input and algorithm
- **Ecosystem Compatibility**: Aligns with SuperMemo ecosystem practices

### User Experience Enhancements:

#### 1. Progressive Disclosure
```go
// Beginner mode: Simplified 3-choice mapping
const (
    BeginnerIncorrect = 2  // Maps to "Incorrect"
    BeginnerHard      = 4  // Maps to "Correct Hard" 
    BeginnerEasy      = 6  // Maps to "Perfect"
)

// Advanced mode: Full 0-6 scale available
// Users can graduate to full scale when comfortable
```

#### 2. Contextual Guidance
- **Hover Help**: Detailed descriptions available for each rating
- **Examples**: Concrete examples for each quality level
- **Quick Reference**: Keyboard shortcuts and mnemonics
- **Visual Aids**: Color coding and icons to reinforce meaning

#### 3. Adaptive Interface
- **Usage Patterns**: Track user preferences to suggest appropriate options
- **Time Pressure**: Quick binary fallback for timed reviews
- **Review Context**: Different interfaces for different note types
- **Statistics**: Show how ratings affect scheduling to educate users

## Consequences

### Positive

- **Algorithmic Effectiveness**: Optimal performance from SM-2 algorithm using intended scale
- **Research Backing**: Decades of SuperMemo research validating this approach
- **Code Simplicity**: Direct integration with go-srs components without translation
- **Long-term Learning**: More nuanced feedback improves algorithm adaptation over time
- **Ecosystem Alignment**: Consistent with broader SuperMemo/SRS research community
- **Statistical Quality**: 7-point scale provides better data for algorithm improvements
- **Future Compatibility**: Positions flotsam for integration with other research-based tools

### Negative

- **Initial Learning Curve**: Users must learn 7-point scale vs simpler alternatives
- **Decision Fatigue**: More options may slow down review process initially
- **User Confusion**: Scale distinctions may not be immediately intuitive
- **Documentation Burden**: Requires comprehensive explanation and examples
- **Support Complexity**: More options mean more potential user questions

### Neutral

- **Industry Deviation**: Different from popular Anki approach but aligned with research
- **UI Complexity**: More interface elements but can be progressively disclosed
- **Training Time**: Initial investment in user education for long-term benefits
- **Cultural Shift**: May require changing user expectations from simpler scales

## Implementation Details

### Quality Enum Implementation
```go
// Quality represents the user's self-evaluation of their recall performance.
// Based on original SuperMemo research and go-srs implementation.
type Quality int

const (
    // NoReview indicates no review was performed (0)
    NoReview Quality = iota
    // IncorrectBlackout indicates total failure to recall (1)
    IncorrectBlackout
    // IncorrectFamiliar indicates incorrect but familiar upon seeing answer (2)
    IncorrectFamiliar
    // IncorrectEasy indicates incorrect but seemed easy upon seeing answer (3)
    IncorrectEasy
    // CorrectHard indicates correct but required significant difficulty (4)
    CorrectHard
    // CorrectEffort indicates correct after some hesitation (5)
    CorrectEffort
    // CorrectEasy indicates correct with perfect recall (6)
    CorrectEasy
)
```

### User Interface Patterns

#### 1. Full Scale Interface
```
┌─────────────────────────────────────────────────────────────┐
│ How well did you recall this information?                   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ [0] No Review     - Skip this card                         │
│ [1] Complete Blank - Total failure to recall               │
│ [2] Wrong/Familiar - Incorrect but familiar when shown     │
│ [3] Wrong/Easy     - Incorrect but seemed easy             │
│ [4] Hard/Correct   - Correct but difficult                 │
│ [5] Some Effort    - Correct with hesitation               │
│ [6] Perfect        - Immediate confident recall            │
│                                                             │
│ Keyboard: 0-6 or ←→ arrows then Enter                      │
└─────────────────────────────────────────────────────────────┘
```

#### 2. Simplified Interface (Beginner Mode)
```
┌─────────────────────────────────────────────────────────────┐
│ How well did you recall this information?                   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ [1] Incorrect  - Wrong or couldn't recall                  │
│ [2] Hard       - Correct but difficult                     │
│ [3] Easy       - Correct and confident                     │
│                                                             │
│ Keyboard: 1-3 or ←→ arrows then Enter                      │
│ Advanced mode available in settings                        │
└─────────────────────────────────────────────────────────────┘
```

### Documentation Strategy

#### 1. Quality Scale Guide
```markdown
# SRS Quality Scale Guide

## Understanding the 0-6 Scale

The flotsam SRS system uses a 7-point quality scale based on SuperMemo research:

### When to Use Each Rating:

**0 - No Review**: You skipped or postponed this card
- Example: Interrupted session, will review later

**1 - Complete Blank**: Total failure to recall anything
- Example: No memory of ever seeing this information

**2 - Wrong but Familiar**: Incorrect answer but familiar when shown
- Example: "Oh right, I remember now but got it completely wrong"

**3 - Wrong but Easy**: Incorrect but seemed easy upon seeing answer  
- Example: "I should have known that, obvious mistake"

**4 - Hard but Correct**: Correct answer but required significant effort
- Example: Took 10+ seconds, had to think hard, uncertain

**5 - Correct with Effort**: Correct answer with some hesitation
- Example: Took 3-5 seconds, minor uncertainty

**6 - Perfect Recall**: Immediate confident correct answer
- Example: Instant recognition, completely confident
```

#### 2. Algorithm Impact Explanation
```markdown
# How Quality Ratings Affect Scheduling

## The Algorithm Connection

Your quality ratings directly influence when you'll see cards again:

### Correct Answers (4-6):
- **Quality 4**: Shorter intervals, more frequent review
- **Quality 5**: Standard intervals
- **Quality 6**: Longer intervals, less frequent review

### Incorrect Answers (1-3):
- **Any incorrect**: Resets progress, back to frequent review
- **Pattern tracking**: Algorithm learns your difficulty patterns

### Why This Matters:
- **Honest ratings** = Better scheduling
- **Consistent ratings** = Algorithm learns your patterns
- **Over/under rating** = Suboptimal learning schedule
```

### Validation and Error Handling
```go
// Validate checks if the quality value is within valid range
func (q Quality) Validate() error {
    if q > CorrectEasy || q < NoReview {
        return fmt.Errorf("invalid quality %d: must be between %d and %d", 
            int(q), int(NoReview), int(CorrectEasy))
    }
    return nil
}

// IsCorrect returns true if the quality represents a correct answer
func (q Quality) IsCorrect() bool {
    return q >= CorrectHard
}

// String provides human-readable description
func (q Quality) String() string {
    descriptions := map[Quality]string{
        NoReview:          "No Review",
        IncorrectBlackout: "Complete Blank",
        IncorrectFamiliar: "Wrong but Familiar", 
        IncorrectEasy:     "Wrong but Easy",
        CorrectHard:       "Hard but Correct",
        CorrectEffort:     "Correct with Effort",
        CorrectEasy:       "Perfect Recall",
    }
    return descriptions[q]
}
```

### Migration and Compatibility

#### Future Scale Support
```go
// QualityMapper allows future support for other scales
type QualityMapper interface {
    ToSM2Quality(input interface{}) (Quality, error)
    FromSM2Quality(q Quality) interface{}
}

// AnkiQualityMapper converts Anki 1-4 scale to 0-6 scale
type AnkiQualityMapper struct{}

func (m *AnkiQualityMapper) ToSM2Quality(ankiRating int) (Quality, error) {
    mapping := map[int]Quality{
        1: IncorrectFamiliar, // Again
        2: CorrectHard,       // Hard  
        3: CorrectEffort,     // Good
        4: CorrectEasy,       // Easy
    }
    if q, exists := mapping[ankiRating]; exists {
        return q, nil
    }
    return NoReview, fmt.Errorf("invalid Anki rating: %d", ankiRating)
}
```

### Analytics and Improvement

#### Usage Tracking
```go
// QualityStats tracks user rating patterns for UX improvement
type QualityStats struct {
    TotalReviews     int64
    QualityHistogram map[Quality]int64
    AverageTime      time.Duration
    ConsistencyScore float64
}

// Track user patterns to improve UX
func (stats *QualityStats) SuggestSimplifiedMode() bool {
    // If user only uses 3 qualities, suggest simplified mode
    usedQualities := 0
    for _, count := range stats.QualityHistogram {
        if count > 0 {
            usedQualities++
        }
    }
    return usedQualities <= 3 && stats.TotalReviews > 50
}
```

---
*ADR format based on [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)*