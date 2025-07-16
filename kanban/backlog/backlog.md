# Backlog

High-level list of future features / improvements. Break into detailed task cards before implementation.

If dumb ideas were worth shit, we'd all be rich.

# bugs

xx None I'm especially conscious of
-: UI is 1 line higher than screen (progressbar or key legend cut off); improvement: form controls should be inside modal / viewport

# features 

Affinity:

```
  habit tracking <-> recurring tasks <-> tasks 
          |    _________________|_______*  |
          |   /                       \_* projects <-> notes
       missions 
```

## Habit Management 

### Atomic Habits
TODO: think about how to adapt / support / reinforce "Atomic Habits" framework, e.g.:

- focus on systems not goals
- Cue -> Craving -> Response -> Reward
  - inform habit creation UI / guidance
  - support from data model?
- Easy -> Attractive -> Obvious -> Satisfying # form
- Invisible -> Unattractive -> Difficult -> Unsatisfying # break
  - Cue 
    - different "trigger" conditions (e.g. timed reminders vs ...)
    - habit stacking: UI to support "stacks"?
  - Craving
    - ...
    - guidance -> text fields -> ...?

  - Response
    - reduce friction
    - guidance: 2 minute rule
  - Reward
    - gamify / score / meta-currency?
    - reporting & analytics
    - automate somehow: user scripts
    - socials

### (add / edit)
- [ ] edit title
- [ ] ? add optional default for boolean fields. This will set the default state for form fields during entry.

## data model

- a habit can have n fields
- more complex (DSL?) criteria
  - combine per-field criteria
  - and / or, or ... any / all
  - criteria spanning fields (at least n of criteria true)
  - checklist % completion
- recurrence schedules
  - calendar vs x every y (redo from when visible or when completed)
    - for numeric fields, quota over time
  - only show for entry when "due"
- checklists 
  - notes?
  - skippable items
  - skip the whole thing
  - % or x/y success thresholds? optional vs non-optional items?
  - embed into logs / tasks? ad-hoc?

## Scoring

- [ ] time of day : before conditions spanning midnight - how can this be adressed?
  (optional)

## reporting / analytics (necessary, soonish)

- it's a whole thing.
- loosely coupled, but reuse a decent chunk of entry parsing / error handling
- performance requirements
  - time series / caching
  - sqlite (projection; text is truth). 
    - ingestion
  - or, skates; convex; ...
- sequence-dependent evaluation (n every m days; streaks; etc)
- https://github.com/NimbleMarkets/ntcharts
- calendar stuff
- lots of opportunity to build stuff that looks cool
  - cli animations!

## cli bling

- better / updated help
- https://github.com/charmbracelet/fang

## configuration

- theme support
  - steal neovim ones
    - gruvbox
    - material
    - paperthingy
    - catpuccin
    - user defined

- prefs / filter state

## TUI

- full screen tui interface
  - finally memory leaks a possibility
- tabs
- design multi-element UI
- command bar
- viewport

## persistent process

- sway / menu bar app
- server process

## oddities

- flotsam: user messages which show up randomly
  - title, notes
  - can be revised, dismissed, dunked (don't show for a while longer)
  - form of SRS / iterative writing
  - convert to X (task? ...)

- other SRS / incremental writing

## expanded features

- persistent processes: api; mcp server; 
- stuff with time
  - timer; stopwatch
  - pomodoros
  - eyestrain / health & mindfulness timer
  - flowtime log
  - time block planner
  - interstitial journal (timestamped (begin/end) logs)

## separate vault / context support

- work | personal | family
- different data dirs
- \-c / --config-dir / VICE_DATA

## task management / note taking

- recurring tasks
  - kinda like habits
  - lots of wrinkles around scheduling, periodicity
  - due, defer dates, deadlines - lots of semantics to unmuddle
  - need reminders, notifications (cross-platform: notify-send, whatever pound of flesh apple needs to use their notifications API)
  - lots of filtering things to solve for large sets of occasional obligations
  - never start a land war in asia

- standard tasks  
  - don't wanna try to be todoist
- integration: taskwarrior? rem?
- personal kanban
  - ready
  - doing
  - done
  - trough_of_lamentation.yml

- work clean style opinionated thing
  - missions
  - back/front burners
  - capacity

- gtd
  - lists
  - actions
    - title
    - status
    - mission
  - projects
    - title
    - status
    - mission
    - 0-1 next \*action 
      - title
      - status
  - projects and actions mostly interchangeable in UI but projects render as 2 lines if they have a next action
  - agendas
    - people are interesting. or are they just lists?

- bullet journal
  - daily log
  - mix tasks / notes
  - simple migration (move to today / bin; toggle day's entry migrated)
  - collections (named notes)

- zettelkasten
  - don't implement 

- notational velocity
  - the rewrite is vapourware right?
  - this + flotsam seems cool
  - it'd also be good to have fast fuzzy search for the trough of disillusionment

- shittywiki

## integrations

- $EDITOR
- Obsidian
- taskwarrior
- rem
- zsh (prompt - counts, or .. see flotsam)
- neovim / emacs
- user scripts (bash?) / plugin system?
- waybar / macos menu bar
- tmux (plugin?)
- zk
- dnote, et al?

## docs

- better readme
- fancy videos
- simple description of data model / features / variations

## compatibility

- Cross-platform testing / fallback for terminal compatibility (low colour; no emoji support!)
  - non-nerdfonts
  - system TTY; kms
  - all the xterms: foot, wezterm, ghostty, gnome, kde, 
  - Apple: iTerm; Terminal
  - SSH
  - shitty slow computer / vm (docker et al?)

## Testing & QA

- Performance testing for large habit sets (100+ habits)
- Table / cartesian product of test scenarios / supported cases
- Fuzz testing (habits > collectors)
- Logging / headless operation? ask stupid to come up with a plan, or do some research (must be patterns for charm out there)
- Stress testing with complex habit schemas and criteria
- Error recovery testing for interrupted entry sessions
  - dump scratch files?

-- AI Slop hereafter: --

# inspired by Harsh
Looking at Harsh's unique features that Vice lacks, here's a prioritized list of improvements for Vice:

## High Priority

- Track skips separately from failures in analytics
- Visual indicators for skipped days vs failed days

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
- `ask` command for quick habit entry without full forms
- Substring filtering for specific habits (e.g., `vice ask gym`)
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