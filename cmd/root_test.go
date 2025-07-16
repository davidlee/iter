package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	t.Run("executes without error", func(t *testing.T) {
		cmd := rootCmd
		cmd.SetArgs([]string{"--help"})
		err := cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("supports config-dir flag", func(t *testing.T) {
		cmd := rootCmd
		cmd.SetArgs([]string{"--config-dir", "/tmp/test"})

		// Parse flags
		err := cmd.ParseFlags([]string{"--config-dir", "/tmp/test"})
		require.NoError(t, err)

		// Check that the flag was set
		flag := cmd.Flag("config-dir")
		require.NotNil(t, flag)
		assert.Equal(t, "/tmp/test", flag.Value.String())
	})
}

func TestInitializeViceEnv(t *testing.T) {
	t.Run("uses default paths when no config-dir flag", func(t *testing.T) {
		// Setup
		tempDir := t.TempDir()
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		require.NoError(t, os.Setenv("XDG_CONFIG_HOME", tempDir))

		// Cleanup
		defer func() {
			if originalXDG == "" {
				require.NoError(t, os.Unsetenv("XDG_CONFIG_HOME"))
			} else {
				require.NoError(t, os.Setenv("XDG_CONFIG_HOME", originalXDG))
			}
		}()

		// Reset configDir to ensure we test default behavior
		originalConfigDir := configDir
		configDir = ""
		defer func() {
			configDir = originalConfigDir
		}()

		// Test
		err := initializeViceEnv(rootCmd, []string{})
		require.NoError(t, err)

		// Verify environment was set correctly
		env := GetViceEnv()
		require.NotNil(t, env)
		expectedConfig := filepath.Join(tempDir, "vice")
		assert.Equal(t, expectedConfig, env.ConfigDir)
		// Data files go in DataDir/Context, not ConfigDir
		assert.Contains(t, env.GetHabitsFile(), "habits.yml")
		assert.Contains(t, env.GetEntriesFile(), "entries.yml")

		// Verify directory was created
		info, err := os.Stat(expectedConfig)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("uses custom config-dir when flag is set", func(t *testing.T) {
		// Setup
		tempDir := t.TempDir()
		customConfigDir := filepath.Join(tempDir, "custom-vice")

		// Set configDir global variable (simulating flag being set)
		originalConfigDir := configDir
		configDir = customConfigDir
		defer func() {
			configDir = originalConfigDir
		}()

		// Test
		err := initializeViceEnv(rootCmd, []string{})
		require.NoError(t, err)

		// Verify environment was set correctly
		env := GetViceEnv()
		require.NotNil(t, env)
		assert.Equal(t, customConfigDir, env.ConfigDir)
		// Data files go in DataDir/Context, not ConfigDir
		assert.Contains(t, env.GetHabitsFile(), "habits.yml")
		assert.Contains(t, env.GetEntriesFile(), "entries.yml")

		// Verify directory was created
		info, err := os.Stat(customConfigDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}

func TestGetPaths(t *testing.T) {
	t.Run("returns paths after initialization", func(t *testing.T) {
		// Setup - ensure viceEnv is initialized
		tempDir := t.TempDir()

		originalConfigDir := configDir
		configDir = tempDir
		defer func() {
			configDir = originalConfigDir
		}()

		err := initializeViceEnv(rootCmd, []string{})
		require.NoError(t, err)

		// Test
		result := GetPaths()
		require.NotNil(t, result)
		assert.Equal(t, tempDir, result.ConfigDir)
	})
}
