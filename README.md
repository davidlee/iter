# vice

![VICE](./doc/images/vice.png)

*Habit forming. Holds your work.*

A command-line / TUI habit tracker with support for flexible habit types and
text (YAML, Markdown) file persistence.

## Status

**vice** is alpha software, and might destroy your property or seriously injure
you for no reason at all.

## Development Methods

**vice** is written almost entirely by Claude Code, partly as an experiment to
determine whether that's even a good idea. 

So far the answer is: _it depends_. But it sure feels weird.

**vice** makes extensive use of [charm libs](https://charm.sh) because
they're swish af. Golang was a consequence of this decision. The author has no
previous experience with Golang.

Even if you have no interest in the software itself, you might find **vice**
interesting as an engineered system for producing specification-driven
software, via amnesiac idiot-savant LLMs. If so, take a look at:
  - [CLAUDE.md](CLAUDE.md)
  - [.claude/commands/](.claude/commands/)
  - [kanban/](kanban/)
  - [doc/](doc/)


## Planned Features & Design Goals

**vice** is not minimalist software; it aims to be a heady cocktail of complementary
but surprising capabilities.

**vice** steals liberally. It lifted code straight from
[ZK](https://zk-org.github.io/zk/) for tight interop, and
[go-srs](https://github.com/revelaction/go-srs) 
for spaced repetition / incremental writing, and got the GPLv3
[license](./LICENSE.md) all over it in the process.

**vice** wants to be:
- **Addictive**: if it's not habit forming, how's it going to help you form habits?
- **Fast Acting**: almost too easy, once you're used to it.
- **Promiscuous**: permissive morals; plays with others.

## Features

- **Simple Habits**: Boolean pass/fail tracking 
- **Elastic Habits**: Multi-level achievement tracking with mini/midi/maxi levels
- **Informational Habits**: Data collection without pass/fail scoring
- **Automatic Scoring**: Habits can be automatically scored based on defined criteria
- **Context Management**: Separate personal/work contexts with isolated data
- **XDG Compliance**: Follows Unix filesystem conventions with proper directory structure
- **Local Storage**: All data stored in local YAML files for version control and portability
- **Interactive CLI**: User-friendly forms with field-specific input validation

## Ethics

You can take some without asking, **vice** will put it on your tab —
but it isn't carbon neutral, and it's probably not vegan either. 

## Installation

```bash
# Build from source
git clone <repository>
cd vice
just build

# Install to PATH
go install .
```

If you're using nixos-direnv, it'll set up shop for you. Otherwise look at the `flake.nix`.

If you're on windows ... I dunno, go buy yourself a real computer.

## Quick Start

1. **First hit's free**:
   ```bash
   vice --help 
   ```
   See what's on offer. Pick something else, and you get a sample (configuration and data files).

2. **Get some habits**:
   ```bash
   vice habits
   ```
   Tickle the wizard, choose your poison. 

2. **Record today's crimes**:
   ```bash
   vice entry
   ```
   A light interrogation to record how it all really happened.

## Configuration

**vice** follows the [XDG spec](https://specifications.freedesktop.org/basedir-spec/latest/), 
and promises not to vomit in the sink.

### Configuration Files
- **Application config**: `config.toml` in `$XDG_CONFIG_HOME/vice/` (default: `~/.config/vice/`)
- **Context state**: `vice.yml` in `$XDG_STATE_HOME/vice/` (default: `~/.local/state/vice/`)

### Data Files (per context)
- **Habit definitions**: `habits.yml` in `$XDG_DATA_HOME/vice/{context}/` 
- **Daily entries**: `entries.yml` in `$XDG_DATA_HOME/vice/{context}/`
- **Checklists**: `checklists.yml` and `checklist_entries.yml` in `$XDG_DATA_HOME/vice/{context}/`

Default data location: `~/.local/share/vice/{context}/` (where context is "personal" or "work" by default)

### Context Isolation

Contexts compartmentalise your stuff, so don't shit where you sleep:

```toml
# config.toml
[core]
contexts = ["personal", "work"]  # Define available contexts
```

Each context maintains completely separate data files, so you can avoid mixing
business and pleasure unless that's your kink.

## Commands

To start the habit entry TUI, run `vice`. For help, `vice --help`.

## Data Storage

### Global Options

All commands support these global flags:

```bash
--config-dir PATH      # Override $XDG_CONFIG_HOME/vice
--data-dir PATH        # Override $XDG_DATA_HOME/vice  
--state-dir PATH       # Override $XDG_STATE_HOME/vice
--cache-dir PATH       # Override $XDG_CACHE_HOME/vice
--context NAME         # Use context temporarily (no state change)
```

### Directory Structure

```
XDG directories:
├── ~/.config/vice/
│   └── config.toml             # Application configuration
├── ~/.local/state/vice/
│   └── vice.yml                # Active context state
├── ~/.local/share/vice/
│   ├── personal/               # Personal context data
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

## Architecture

- [Architecture Overview](doc/Architecture.md) - start here
- [specifications](specifications/) - more thorough, focused specifications 