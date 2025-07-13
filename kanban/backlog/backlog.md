# Backlog

High-level list of future features / improvements. Break into detailed task cards before implementation.

## Goal Management

## Schema / add / edit

- add optional default for boolean fields. This will set the default state for form fields during entry.

### List

  - [ ] List goals by title; select goal to view details
  - [ ] Rich table display with goal summaries
  - [ ] Interactive filtering and sorting
  - [ ] Goal status indicators (manual/automatic scoring, completeness)
  - [ ] Search functionality for large goal sets

## reporting 

- https://github.com/NimbleMarkets/ntcharts

# inspired by Harsh
Looking at Harsh's unique features that Iter lacks, here's a prioritized list of improvements for Iter:

## High Priority

**1. Skip functionality with visual tracking**
- Add "skip" option alongside current goal responses
- Track skips separately from failures in analytics
- Visual indicators for skipped days vs failed days
- Essential for real-world habit tracking where circumstances prevent completion

**2. Consistency graph visualization**
- Terminal-based visual habit chains (Seinfeld method)
- Character-based progress indicators (━, •, ·, etc.)
- Rolling window display (last 30-100 days)
- Provides immediate visual feedback that's core to habit psychology

**3. Streak break warnings**
- Proactive alerts when habits are at risk
- Warning indicators in upcoming habit displays
- Configurable warning periods based on habit frequency
- Prevents accidental streak breaks through awareness

## Medium Priority

**4. Flexible habit frequencies**
- Support for "X times per Y days" patterns (e.g., 3/7 for 3 times per week)
- Rolling time windows vs fixed calendar periods
- Custom interval tracking (every N days)
- More realistic than daily-only tracking for many habits

**5. Enhanced CLI workflow commands**
- `todo` command showing today's incomplete habits
- `ask` command for quick habit entry without full forms
- Substring filtering for specific habits (e.g., `iter ask gym`)
- Streamlines daily usage vs current form-based approach

**6. Habit categorization and organization**
- Visual groupings (Dailies, Weeklies, etc.)
- Section headers in config and display
- Organizational comments in config files
- Better management of multiple habits

## Lower Priority

**7. Date-specific entry capabilities**
- Backfill missed days with specific dates
- Yesterday shortcuts (`yday`, `yd`)
- ISO date targeting for habit entry
- Historical data correction without manual file editing

**8. Terminal display enhancements**
- Dynamic terminal width adaptation
- Color-coded output with no-color option
- Sparkline overview graphs
- Better visual hierarchy and scanning

**9. Advanced data annotation**
- In-line quantity tracking (@ symbol equivalent)
- Comment support for context (# symbol equivalent)
- Notes field enhancement beyond current basic notes
- Richer data collection for pattern analysis

**10. Configuration quality-of-life improvements**
- Environment variable config path override
- Automatic sample habit generation
- Configuration validation and helpful error messages
- Easier habit file management

The top 3 features (skip functionality, consistency graphs, and streak warnings) would transform Iter from a data collection tool into a true habit formation system by adding the psychological reinforcement mechanisms that make habit tracking effective.