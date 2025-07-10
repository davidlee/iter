package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultPaths(t *testing.T) {
	t.Run("uses XDG_CONFIG_HOME when set", func(t *testing.T) {
		// Setup
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		testConfigDir := "/tmp/test-config"
		require.NoError(t, os.Setenv("XDG_CONFIG_HOME", testConfigDir))

		// Cleanup
		defer func() {
			if originalXDG == "" {
				require.NoError(t, os.Unsetenv("XDG_CONFIG_HOME"))
			} else {
				require.NoError(t, os.Setenv("XDG_CONFIG_HOME", originalXDG))
			}
		}()

		// Test
		paths, err := GetDefaultPaths()
		require.NoError(t, err)

		expected := filepath.Join(testConfigDir, "iter")
		assert.Equal(t, expected, paths.ConfigDir)
		assert.Equal(t, filepath.Join(expected, "goals.yml"), paths.GoalsFile)
		assert.Equal(t, filepath.Join(expected, "entries.yml"), paths.EntriesFile)
	})

	t.Run("uses ~/.config when XDG_CONFIG_HOME not set", func(t *testing.T) {
		// Setup
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		require.NoError(t, os.Unsetenv("XDG_CONFIG_HOME"))

		// Cleanup
		defer func() {
			if originalXDG != "" {
				require.NoError(t, os.Setenv("XDG_CONFIG_HOME", originalXDG))
			}
		}()

		// Test
		paths, err := GetDefaultPaths()
		require.NoError(t, err)

		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		expected := filepath.Join(homeDir, ".config", "iter")
		assert.Equal(t, expected, paths.ConfigDir)
		assert.Equal(t, filepath.Join(expected, "goals.yml"), paths.GoalsFile)
		assert.Equal(t, filepath.Join(expected, "entries.yml"), paths.EntriesFile)
	})

	t.Run("ignores empty XDG_CONFIG_HOME", func(t *testing.T) {
		// Setup
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		require.NoError(t, os.Setenv("XDG_CONFIG_HOME", ""))

		// Cleanup
		defer func() {
			if originalXDG == "" {
				require.NoError(t, os.Unsetenv("XDG_CONFIG_HOME"))
			} else {
				require.NoError(t, os.Setenv("XDG_CONFIG_HOME", originalXDG))
			}
		}()

		// Test
		paths, err := GetDefaultPaths()
		require.NoError(t, err)

		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		expected := filepath.Join(homeDir, ".config", "iter")
		assert.Equal(t, expected, paths.ConfigDir)
	})
}

func TestGetPathsWithConfigDir(t *testing.T) {
	customDir := "/custom/config/path"

	paths := GetPathsWithConfigDir(customDir)

	assert.Equal(t, customDir, paths.ConfigDir)
	assert.Equal(t, filepath.Join(customDir, "goals.yml"), paths.GoalsFile)
	assert.Equal(t, filepath.Join(customDir, "entries.yml"), paths.EntriesFile)
}

func TestEnsureConfigDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	testConfigDir := filepath.Join(tempDir, "test-iter-config")

	paths := GetPathsWithConfigDir(testConfigDir)

	// Directory should not exist initially
	_, err := os.Stat(testConfigDir)
	assert.True(t, os.IsNotExist(err))

	// EnsureConfigDir should create it
	err = paths.EnsureConfigDir()
	require.NoError(t, err)

	// Directory should now exist
	info, err := os.Stat(testConfigDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Should have correct permissions
	assert.Equal(t, os.FileMode(0o750), info.Mode().Perm())
}

func TestGetXDGConfigDir(t *testing.T) {
	t.Run("returns XDG_CONFIG_HOME when set", func(t *testing.T) {
		// Setup
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		testPath := "/test/xdg/config"
		require.NoError(t, os.Setenv("XDG_CONFIG_HOME", testPath))

		// Cleanup
		defer func() {
			if originalXDG == "" {
				require.NoError(t, os.Unsetenv("XDG_CONFIG_HOME"))
			} else {
				require.NoError(t, os.Setenv("XDG_CONFIG_HOME", originalXDG))
			}
		}()

		// Test
		result, err := getXDGConfigDir()
		require.NoError(t, err)
		assert.Equal(t, testPath, result)
	})

	t.Run("returns ~/.config when XDG_CONFIG_HOME not set", func(t *testing.T) {
		// Setup
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		require.NoError(t, os.Unsetenv("XDG_CONFIG_HOME"))

		// Cleanup
		defer func() {
			if originalXDG != "" {
				require.NoError(t, os.Setenv("XDG_CONFIG_HOME", originalXDG))
			}
		}()

		// Test
		result, err := getXDGConfigDir()
		require.NoError(t, err)

		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)
		expected := filepath.Join(homeDir, ".config")

		assert.Equal(t, expected, result)
	})
}
