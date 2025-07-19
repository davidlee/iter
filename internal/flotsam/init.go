// Package flotsam provides auto-initialization for flotsam directories and ZK notebooks.
// AIDEV-NOTE: T041/6.1-auto-init; transparent setup when ZK available and flotsam dir missing
package flotsam

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"

	"github.com/davidlee/vice/internal/config"
	"github.com/davidlee/vice/internal/zk"
)

// EnsureFlotsamEnvironment sets up flotsam directory and ZK notebook if needed.
// This function is idempotent and safe to call multiple times.
// AIDEV-NOTE: auto-init strategy - transparent, graceful, user-friendly setup
func EnsureFlotsamEnvironment(env *config.ViceEnv) error {
	return EnsureFlotsamEnvironmentWithZK(env, env.ZK)
}

// EnsureFlotsamEnvironmentWithZK allows dependency injection for testing.
// AIDEV-NOTE: testable version with ZK tool injection for unit tests
func EnsureFlotsamEnvironmentWithZK(env *config.ViceEnv, zkTool interface{}) error {
	flotsamDir := filepath.Join(env.ContextData, "flotsam")

	// Check if flotsam environment is fully initialized
	// AIDEV-NOTE: T041-directory-fix; now checks complete initialization (dir + .zk) not just directory existence
	if !IsFlotsamInitialized(env) {
		log.Debug("Flotsam environment incomplete, checking ZK availability for auto-init")

		// Determine ZK availability from injected tool
		var zkAvailable bool
		var zkExecutor interface {
			Execute(...string) (*zk.ToolResult, error)
		}

		if zkTool != nil {
			if tool, ok := zkTool.(interface{ Available() bool }); ok {
				zkAvailable = tool.Available()
			}
			if exec, ok := zkTool.(interface {
				Execute(...string) (*zk.ToolResult, error)
			}); ok {
				zkExecutor = exec
			}
		}

		// Only auto-init if ZK is available
		if !zkAvailable {
			log.Warn("ZK tool not available - flotsam directory creation skipped",
				"directory", flotsamDir,
				"install_url", "https://github.com/zk-org/zk")
			return nil // Not an error - graceful degradation
		}

		log.Info("Auto-initializing flotsam environment", "directory", flotsamDir)

		// Create flotsam directory
		if err := os.MkdirAll(flotsamDir, 0o750); err != nil {
			return fmt.Errorf("failed to create flotsam directory: %w", err)
		}

		// Initialize ZK notebook in flotsam directory
		if err := initializeZKNotebookWithTool(flotsamDir, zkExecutor); err != nil {
			return fmt.Errorf("failed to initialize ZK notebook: %w", err)
		}

		log.Info("Flotsam environment initialized successfully")
	}

	return nil
}

// initializeZKNotebookWithTool creates a ZK notebook using the provided ZK executor.
// AIDEV-NOTE: testable version with ZK executor injection
func initializeZKNotebookWithTool(flotsamDir string, zkExecutor interface {
	Execute(...string) (*zk.ToolResult, error)
}) error {
	// Change to flotsam directory for ZK init
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if err := os.Chdir(flotsamDir); err != nil {
		return fmt.Errorf("failed to change to flotsam directory: %w", err)
	}

	// Ensure we return to original directory
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			log.Error("Failed to return to original directory", "error", err)
		}
	}()

	// Run zk init to create .zk directory and config
	if zkExecutor == nil {
		return fmt.Errorf("ZK executor not available")
	}

	// AIDEV-NOTE: T041-interactive-fix; use --no-input to avoid hanging on interactive prompts
	result, err := zkExecutor.Execute("init", "--no-input")
	if err != nil {
		return fmt.Errorf("zk init failed: %w", err)
	}

	log.Debug("ZK notebook initialized", "output", result.Stdout)
	return nil
}

// IsFlotsamInitialized checks if flotsam environment is properly set up.
// AIDEV-NOTE: validation helper for checking complete setup state
func IsFlotsamInitialized(env *config.ViceEnv) bool {
	flotsamDir := filepath.Join(env.ContextData, "flotsam")
	zkDir := filepath.Join(flotsamDir, ".zk")

	// Check if both flotsam and .zk directories exist
	flotsamExists := dirExists(flotsamDir)
	zkExists := dirExists(zkDir)

	return flotsamExists && zkExists
}

// dirExists checks if a directory exists.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
