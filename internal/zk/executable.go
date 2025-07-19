// Package zk provides abstraction for external command-line tools integration.
// This file implements the simplified ZKExecutable for T041/4.1 basic abstraction.
// AIDEV-NOTE: T041/4.1-zk-abstraction; composition-based tool interface, replaces complex tool.go
package zk

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CommandLineTool provides generic interface for external command-line tools.
// This interface supports zk, taskwarrior, remind, and other Unix tools.
type CommandLineTool interface {
	Name() string
	Available() bool
	Execute(args ...string) (*ToolResult, error)
}

//revive:disable-next-line:exported // ZKTool naming is intentional to avoid conflicts with existing ZK package
// ZKTool extends CommandLineTool with zk-specific operations for note management.
type ZKTool interface {
	CommandLineTool
	List(filters ...string) ([]string, error)
	Edit(paths ...string) error
	GetLinkedNotes(path string) ([]string, []string, error) // backlinks, outbound
}

// ToolResult represents the result of a command-line tool execution.
type ToolResult struct {
	Stdout   string        // Standard output
	Stderr   string        // Standard error
	ExitCode int           // Process exit code
	Duration time.Duration // Execution time
}

//revive:disable-next-line:exported // ZKExecutable naming is intentional to match existing patterns
// ZKExecutable implements ZKTool interface for zk command-line tool integration.
// AIDEV-NOTE: simplified implementation for T041/4.1 - basic shell-out abstraction with graceful degradation
type ZKExecutable struct {
	path      string // Resolved zk binary path
	available bool   // Runtime availability status
	warned    bool   // Track if user has been warned about missing zk
}

// NewZKExecutable creates a new ZK tool instance with runtime detection.
// It searches for zk in PATH and initializes availability status.
func NewZKExecutable() *ZKExecutable {
	path, err := exec.LookPath("zk")
	available := err == nil && path != ""
	
	return &ZKExecutable{
		path:      path,
		available: available,
		warned:    false,
	}
}

// Name returns the tool name identifier.
func (z *ZKExecutable) Name() string {
	return "zk"
}

// Available returns true if zk is available in the system PATH.
func (z *ZKExecutable) Available() bool {
	return z.available
}

// Execute runs a zk command with the specified arguments.
// Returns ToolResult with stdout, stderr, exit code, and execution time.
func (z *ZKExecutable) Execute(args ...string) (*ToolResult, error) {
	if !z.available {
		return nil, fmt.Errorf("zk not available - install from https://github.com/zk-org/zk")
	}

	start := time.Now()
	
	//nolint:gosec // Command arguments are controlled by vice, not user input
	cmd := exec.Command(z.path, args...)
	
	// Capture both stdout and stderr
	stdout, err := cmd.Output()
	stderr := ""
	var exitCode int
	
	// Get stderr from ExitError if command failed
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute zk command: %w", err)
		}
	}
	
	return &ToolResult{
		Stdout:   string(stdout),
		Stderr:   stderr,
		ExitCode: exitCode,
		Duration: time.Since(start),
	}, nil
}

// List executes 'zk list' with optional filters and returns note paths.
// Filters are passed directly as arguments to 'zk list'.
func (z *ZKExecutable) List(filters ...string) ([]string, error) {
	args := append([]string{"list"}, filters...)
	
	result, err := z.Execute(args...)
	if err != nil {
		return nil, fmt.Errorf("zk list failed: %w", err)
	}
	
	if result.ExitCode != 0 {
		return nil, fmt.Errorf("zk list failed with exit code %d: %s", result.ExitCode, result.Stderr)
	}
	
	// Parse output into note paths (one per line)
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	var paths []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}
	
	return paths, nil
}

// Edit executes 'zk edit' with the specified note paths.
// This delegates to zk's editor integration for note editing.
func (z *ZKExecutable) Edit(paths ...string) error {
	if len(paths) == 0 {
		return fmt.Errorf("no paths specified for editing")
	}
	
	args := append([]string{"edit"}, paths...)
	
	result, err := z.Execute(args...)
	if err != nil {
		return fmt.Errorf("zk edit failed: %w", err)
	}
	
	if result.ExitCode != 0 {
		return fmt.Errorf("zk edit failed with exit code %d: %s", result.ExitCode, result.Stderr)
	}
	
	return nil
}

// GetLinkedNotes returns backlinks and outbound links for a note path.
// Returns (backlinks, outbound, error) where both slices contain note paths.
func (z *ZKExecutable) GetLinkedNotes(path string) ([]string, []string, error) {
	// Get backlinks: notes that link TO this note
	backlinks, err := z.List("--linked-by", path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get backlinks for %s: %w", path, err)
	}
	
	// Get outbound links: notes this note links TO
	outbound, err := z.List("--link-to", path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get outbound links for %s: %w", path, err)
	}
	
	return backlinks, outbound, nil
}

// WarnIfUnavailable prints a warning message if zk is unavailable.
// Only warns once per session to avoid spam.
func (z *ZKExecutable) WarnIfUnavailable() {
	if !z.available && !z.warned {
		fmt.Fprintf(os.Stderr, "Warning: zk not found in PATH. Install from https://github.com/zk-org/zk for full flotsam functionality.\n")
		z.warned = true
	}
}

// ValidateZKConfig validates .zk/config.toml for compatibility.
// Currently a NOOP placeholder for future enhancement (T046).
func ValidateZKConfig(configPath string) error {
	// AIDEV-NOTE: NOOP for T041/4.1 - placeholder for future validation in T046
	// Future: Parse TOML, check for incompatible settings
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config doesn't exist - this is fine, zk will use defaults
		return nil
	}
	
	// Config exists - assume it's valid for now
	// TODO: Parse and validate in T046
	return nil
}

// FindZKNotebook searches for .zk directory in current and parent directories.
// Returns the notebook root directory or error if not found.
func FindZKNotebook(startDir string) (string, error) {
	dir := startDir
	
	for {
		zkDir := filepath.Join(dir, ".zk")
		if info, err := os.Stat(zkDir); err == nil && info.IsDir() {
			return dir, nil
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}
	
	return "", fmt.Errorf("no .zk directory found in %s or parent directories", startDir)
}