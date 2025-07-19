package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/davidlee/vice/internal/config"
)

// doctorCmd represents the doctor command for system health checks
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health and dependencies",
	Long: `Check system health and dependencies for vice functionality.

The doctor command validates:
- Vice configuration and directory structure
- External tool dependencies (zk, etc.)
- Database connectivity and integrity
- Context configuration and availability

This helps diagnose common setup and configuration issues.`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

// runDoctor performs comprehensive system health checks
func runDoctor(_ *cobra.Command, _ []string) error {
	fmt.Println("🔍 Running vice system diagnostics...")
	fmt.Println()

	// Initialize environment
	env, err := config.GetViceEnvWithOverrides(getDirectoryOverrides())
	if err != nil {
		return fmt.Errorf("failed to initialize environment: %w", err)
	}

	allOK := true
	
	// Check vice configuration
	allOK = checkViceConfiguration(env) && allOK
	
	// Check external dependencies
	allOK = checkExternalDependencies(env) && allOK
	
	// Check database connectivity
	allOK = checkDatabases(env) && allOK
	
	// Summary
	fmt.Println()
	if allOK {
		fmt.Println("✅ All systems healthy!")
		return nil
	}
	fmt.Println("⚠️  Some issues detected. See details above.")
	return nil // Don't return error, just inform user
}

// checkViceConfiguration validates vice directory structure and config
func checkViceConfiguration(env *config.ViceEnv) bool {
	fmt.Println("📁 Checking vice configuration...")
	
	allOK := true
	
	// Check XDG directories
	directories := map[string]string{
		"Config": env.ConfigDir,
		"Data":   env.DataDir,
		"State":  env.StateDir,
		"Cache":  env.CacheDir,
	}
	
	for name, dir := range directories {
		if info, err := os.Stat(dir); err == nil {
			if info.IsDir() {
				fmt.Printf("   ✅ %s directory: %s\n", name, dir)
			} else {
				fmt.Printf("   ❌ %s path exists but is not a directory: %s\n", name, dir)
				allOK = false
			}
		} else {
			fmt.Printf("   ⚠️  %s directory missing (will be created): %s\n", name, dir)
		}
	}
	
	// Check context configuration
	fmt.Printf("   ✅ Active context: %s\n", env.Context)
	fmt.Printf("   ✅ Context data directory: %s\n", env.ContextData)
	
	// Check available contexts
	if len(env.Contexts) > 0 {
		fmt.Printf("   ✅ Available contexts: %v\n", env.Contexts)
	} else {
		fmt.Printf("   ⚠️  No contexts configured (using defaults)\n")
	}
	
	fmt.Println()
	return allOK
}

// checkExternalDependencies validates external tool availability
func checkExternalDependencies(env *config.ViceEnv) bool {
	fmt.Println("🔧 Checking external dependencies...")
	
	allOK := true
	
	// Check ZK availability
	if env.IsZKAvailable() {
		fmt.Printf("   ✅ zk tool: available at %s\n", env.ZK.Name())
		
		// Try to get version if possible
		if result, err := env.ZK.Execute("--version"); err == nil && result.ExitCode == 0 {
			version := result.Stdout
			if len(version) > 50 { // Truncate long version strings
				version = version[:50] + "..."
			}
			fmt.Printf("   ✅ zk version: %s\n", version)
		}
		
		// Check ZK notebook configuration
		if err := env.ValidateZKNotebook(); err != nil {
			fmt.Printf("   ⚠️  zk configuration issue: %v\n", err)
		} else {
			fmt.Printf("   ✅ zk notebook configuration: valid\n")
		}
	} else {
		fmt.Printf("   ❌ zk tool: not found in PATH\n")
		fmt.Printf("       Install from: https://github.com/zk-org/zk\n")
		fmt.Printf("       Note: Some flotsam features require zk\n")
		allOK = false
	}
	
	fmt.Println()
	return allOK
}

// checkDatabases validates database connectivity and structure
func checkDatabases(env *config.ViceEnv) bool {
	fmt.Println("💾 Checking databases...")
	
	allOK := true
	
	// Check SRS database (if it exists)
	srsDBPath := filepath.Join(env.GetFlotsamDir(), ".vice", "flotsam.db")
	if info, err := os.Stat(srsDBPath); err == nil {
		if info.IsDir() {
			fmt.Printf("   ❌ SRS database path is a directory: %s\n", srsDBPath)
			allOK = false
		} else {
			fmt.Printf("   ✅ SRS database: %s\n", srsDBPath)
			fmt.Printf("   ✅ Database size: %d bytes\n", info.Size())
		}
	} else {
		fmt.Printf("   ℹ️  SRS database not found (will be created when needed): %s\n", srsDBPath)
	}
	
	// Check habits/entries files in context
	habitsFile := filepath.Join(env.ContextData, "habits.yml")
	if _, err := os.Stat(habitsFile); err == nil {
		fmt.Printf("   ✅ Habits file: %s\n", habitsFile)
	} else {
		fmt.Printf("   ℹ️  Habits file not found (will be created when needed): %s\n", habitsFile)
	}
	
	entriesFile := filepath.Join(env.ContextData, "entries.yml")
	if _, err := os.Stat(entriesFile); err == nil {
		fmt.Printf("   ✅ Entries file: %s\n", entriesFile)
	} else {
		fmt.Printf("   ℹ️  Entries file not found (will be created when needed): %s\n", entriesFile)
	}
	
	fmt.Println()
	return allOK
}

// getDirectoryOverrides extracts directory overrides from cobra flags
func getDirectoryOverrides() config.DirectoryOverrides {
	return config.DirectoryOverrides{
		ConfigDir: configDir,
		DataDir:   dataDir,
		StateDir:  stateDir,
		CacheDir:  cacheDir,
		Context:   contextFlag,
	}
}