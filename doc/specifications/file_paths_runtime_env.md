# Specification: File paths & runtime environment

feature: config.toml and contexts

dependencies: pelletier/go-toml library 

support a user config.toml (https://toml.io/en/) in the vice data directory. 
This is where keybindings, colour theme settings, emoji to habit status mappings, etc will be defined. 

If missing, on vice invocation a settings.toml with default settings will be created (ideally, commented out).

For now there is only one config setting:

```
[core]
contexts = ["personal", "work"] # default values
```

file paths adhere to https://specifications.freedesktop.org/basedir-spec/latest/:

```

$VICE_CONFIG = ($XDG_CONFIG_HOME || $HOME/.config)/vice # default: ~/.config/vice
$VICE_DATA = ($XDG_DATA_HOME || $HOME/.local/share)/vice # default: ~/.local/share/vice
$VICE_STATE = ($XDG_STATE_HOME || $HOME/.local/state)/vice # default: ~/.local/state/vice
$VICE_CACHE = ($XDG_CACHE_HOME || $HOME/.cache)/vice # default: ~/.cache/vice
```

these are all expanded to absolute paths. empty directories will be recursively created (mkdir -p) if missing.


*config location*
config.toml resides in $VICE_CONFIG

*data location*
`core.contexts` defines an array of folders under $VICE_DATA to be created (which is where YAML data lives); 
by default:
- ~/.local/share/vice/personal
- ~/.local/share/vice/work

these will be created if they don't exist for the active context, and minimal but non-empty files created inside them.

only one context is ever active at once (by default the first in the array); the context compartmentalises all data.

the active context can be overridden by `$VICE_CONTEXT`, and is stored between invocations in `$VICE_STATE/vice.yml`

---

These values are read by vice into a struct available at runtime:
```
ViceEnv {
  config: $VICE_CONFIG
  data: $VICE_DATA,
  state: $VICE_STATE,
  cache: $VICE_CACHE,
  context: $VICE_CONTEXT,
  context_data: "$VICE_DATA/$VICE_CONTEXT",
  ...
}
```

which is used everywhere these values are referenced.
`context` may change at runtime (in which case the state file will be written, `context_data` will be recomputed, and any data or cache reloaded.)

the override priority (example for config dir) is:

`$HOME -> $XDG_CONFIG_HOME -> $VICE_CONFIG -> (command line flag --config-dir or -c) ->  ViceEnv.config`

---

## Implementation Status

**Current Implementation**: See [T028 File Paths & Runtime Environment](../kanban/in-progress/T028_file_paths_runtime_env.md)

- Phase 1 Complete: ViceEnv struct with full XDG compliance and TOML configuration
- Phase 2 In Progress: Context-aware data loading with Repository Pattern
- Architectural decision: Repository Pattern with "turn off and on again" context switching

**Key Design Decisions**:
- Repository Pattern chosen over lazy loading for simplicity and clear migration path
- Context switching performs complete data unload/reload for race condition avoidance
- Interface-based design allows future evolution to sophisticated caching/lazy loading
