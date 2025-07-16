package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"davidlee/vice/internal/debug"
	"github.com/spf13/cobra"
)

// prototypeCmd represents the prototype command for T024 modal investigation
var prototypeCmd = &cobra.Command{
	Use:   "prototype",
	Short: "Run T024 modal investigation prototype",
	Long: `Run the T024 modal investigation prototype to test modal behavior.

This command executes the test_modal_prototype.go file which contains
incremental complexity integration for debugging the auto-closing modal bug.

The prototype includes:
- Field Input Factory integration
- EntryMenuModel integration layer  
- Entry Collection Context with complex state

Use this command to test modal behavior without interfering with main application builds.`,
	RunE: runPrototype,
}

func init() {
	rootCmd.AddCommand(prototypeCmd)
}

//revive:disable-next-line:unused-parameter -- cmd required by cobra.Command interface
func runPrototype(cmd *cobra.Command, args []string) error {
	// Initialize debug logging if requested
	if debugMode {
		err := debug.GetInstance().Initialize(paths.ConfigDir)
		if err != nil {
			return fmt.Errorf("failed to initialize debug logging: %w", err)
		}
		defer func() {
			_ = debug.GetInstance().Close()
		}()
	}

	// Get the root directory of the project
	rootDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Path to the prototype file
	prototypePath := filepath.Join(rootDir, "prototype", "test_modal_prototype.go")

	// Check if prototype file exists
	if _, err := os.Stat(prototypePath); os.IsNotExist(err) {
		return fmt.Errorf("prototype file not found: %s", prototypePath)
	}

	// Execute the prototype
	fmt.Println("🔬 Running T024 modal investigation prototype...")
	fmt.Printf("📂 Prototype path: %s\n", prototypePath)
	if debugMode {
		fmt.Printf("📝 Debug logging enabled: %s/vice-debug.log\n", paths.ConfigDir)
	}
	fmt.Println()

	// Run the prototype using go run
	// #nosec G204 -- prototypePath is constructed from trusted cwd + literal path
	execCmd := exec.Command("go", "run", prototypePath)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin

	return execCmd.Run()
}
