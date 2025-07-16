package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"davidlee/vice/internal/config"
)

// contextCmd represents the context command
// AIDEV-NOTE: T028/4.1-context-commands; persistent context switching commands for runtime operations
var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage contexts for organizing habit data",
	Long: `Manage contexts to organize your habit data into separate environments.
Contexts allow you to maintain completely separate habit tracking data for different
aspects of your life (e.g., personal, work, travel).

Available contexts are defined in config.toml. Context switching persists the 
active context to your state file, unlike the transient --context flag or 
VICE_CONTEXT environment variable.

Examples:
  vice context list           # Show all available contexts and current active context
  vice context show           # Show only the current active context  
  vice context switch work    # Switch to work context (persistent)
  
Transient context overrides (do not persist):
  vice --context work todo    # Use work context for one command
  VICE_CONTEXT=work vice todo # Use work context via environment variable`,
}

// contextListCmd lists all available contexts
var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available contexts",
	Long: `List all available contexts defined in config.toml and show which one is currently active.
The active context determines where your habit data is stored and loaded from.`,
	RunE: runContextList,
}

// contextShowCmd shows the current context
var contextShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current active context",
	Long: `Show the currently active context. This is the context being used for 
habit data storage and loading.`,
	RunE: runContextShow,
}

// contextSwitchCmd switches to a different context
var contextSwitchCmd = &cobra.Command{
	Use:   "switch <context-name>",
	Short: "Switch to a different context",
	Long: `Switch to a different context and persist the change to the state file.
The context must be defined in your config.toml file.

This is different from the --context flag which is transient (temporary).
Context switching with this command persists the change between sessions.`,
	Args: cobra.ExactArgs(1),
	RunE: runContextSwitch,
}

func init() {
	rootCmd.AddCommand(contextCmd)
	contextCmd.AddCommand(contextListCmd)
	contextCmd.AddCommand(contextShowCmd)
	contextCmd.AddCommand(contextSwitchCmd)
}

func runContextList(_ *cobra.Command, _ []string) error {
	// AIDEV-NOTE: T028/4.1-context-list; displays all available contexts with current marker
	// Get the resolved environment
	env := GetViceEnv()

	// Show available contexts
	fmt.Printf("Available contexts (defined in %s):\n", env.GetConfigTomlPath())
	for i, ctx := range env.Contexts {
		marker := "  "
		if ctx == env.Context {
			marker = "* "
		}
		fmt.Printf("%s%d. %s", marker, i+1, ctx)
		if ctx == env.Context {
			fmt.Printf(" (current)")
		}
		fmt.Println()
	}

	if len(env.Contexts) == 0 {
		fmt.Println("  No contexts defined in config.toml")
	}

	fmt.Printf("\nCurrent active context: %s\n", env.Context)
	fmt.Printf("Data directory: %s\n", env.ContextData)

	return nil
}

func runContextShow(_ *cobra.Command, _ []string) error {
	// Get the resolved environment
	env := GetViceEnv()

	fmt.Printf("Current context: %s\n", env.Context)
	fmt.Printf("Data directory: %s\n", env.ContextData)
	
	return nil
}

func runContextSwitch(_ *cobra.Command, args []string) error {
	// AIDEV-NOTE: T028/4.1-context-switch; persistent context switching with validation and state persistence
	newContext := args[0]
	
	// Get the resolved environment
	env := GetViceEnv()

	// Check if context exists in available contexts
	contextExists := false
	for _, ctx := range env.Contexts {
		if ctx == newContext {
			contextExists = true
			break
		}
	}

	if !contextExists {
		return fmt.Errorf("context '%s' not found in available contexts: %s", 
			newContext, strings.Join(env.Contexts, ", "))
	}

	// Check if already active
	if env.Context == newContext {
		fmt.Printf("Already using context '%s'\n", newContext)
		return nil
	}

	// Switch to the new context
	if err := config.SwitchContext(env, newContext); err != nil {
		return fmt.Errorf("failed to switch context: %w", err)
	}

	fmt.Printf("Switched to context '%s'\n", newContext)
	fmt.Printf("Data directory: %s\n", env.ContextData)

	return nil
}