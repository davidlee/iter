package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davidlee/vice/internal/config"
	"github.com/davidlee/vice/internal/zk"
)

func TestDoctorCmd_Basic(t *testing.T) {
	// Create temporary environment
	tmpDir := t.TempDir()
	
	// Set up environment variables for test
	originalHome := os.Getenv("HOME")
	originalViceData := os.Getenv("VICE_DATA")
	
	_ = os.Setenv("HOME", tmpDir)
	_ = os.Setenv("VICE_DATA", filepath.Join(tmpDir, "vice-data"))
	
	defer func() {
		_ = os.Setenv("HOME", originalHome)
		if originalViceData != "" {
			_ = os.Setenv("VICE_DATA", originalViceData)
		} else {
			_ = os.Unsetenv("VICE_DATA")
		}
	}()
	
	// Capture stdout since doctor prints to stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Run doctor command
	rootCmd.SetArgs([]string{"doctor"})
	err := rootCmd.Execute()
	
	// Close write end and restore stdout
	_ = w.Close()
	os.Stdout = originalStdout
	
	if err != nil {
		t.Errorf("doctor command should not return error, got: %v", err)
	}
	
	// Read captured output
	output := make([]byte, 2000)
	n, _ := r.Read(output)
	outputStr := string(output[:n])
	
	// Check that output contains expected sections
	expectedSections := []string{
		"Running vice system diagnostics",
		"Checking vice configuration",
		"Checking external dependencies",
		"Checking databases",
	}
	
	for _, section := range expectedSections {
		if !strings.Contains(outputStr, section) {
			t.Errorf("Output should contain '%s', got:\n%s", section, outputStr)
		}
	}
}

func TestCheckViceConfiguration(_ *testing.T) {
	// Create test environment
	env := &config.ViceEnv{
		ConfigDir:   "/tmp/test-config",
		DataDir:     "/tmp/test-data",
		StateDir:    "/tmp/test-state",
		CacheDir:    "/tmp/test-cache",
		Context:     "test-context",
		ContextData: "/tmp/test-data/test-context",
		Contexts:    []string{"test-context", "another-context"},
	}
	
	// Test with non-existent directories (should handle gracefully)
	result := checkViceConfiguration(env)
	
	// Should not fail completely for missing directories
	_ = result // We don't assert specific result since it depends on system state
}

func TestCheckExternalDependencies_ZKAvailable(t *testing.T) {
	// Create mock ZK tool (available)
	mockZK := &zk.ZKExecutable{}
	
	env := &config.ViceEnv{
		ZK: mockZK,
	}
	
	// Capture output
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	result := checkExternalDependencies(env)
	
	_ = w.Close()
	os.Stdout = originalStdout
	
	// Read captured output
	output := make([]byte, 1000)
	n, _ := r.Read(output)
	outputStr := string(output[:n])
	
	// Should mention zk status
	if !strings.Contains(outputStr, "zk tool") {
		t.Error("Output should mention zk tool status")
	}
	
	_ = result // Result depends on actual zk availability
}

func TestCheckExternalDependencies_ZKUnavailable(t *testing.T) {
	// Create mock environment with no ZK
	env := &config.ViceEnv{
		ZK: nil,
	}
	
	// Capture output to prevent spam during tests
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	result := checkExternalDependencies(env)
	
	_ = w.Close()
	os.Stdout = originalStdout
	
	// Read and verify output mentions installation
	output := make([]byte, 1000)
	n, _ := r.Read(output)
	outputStr := string(output[:n])
	
	if !strings.Contains(outputStr, "https://github.com/zk-org/zk") {
		t.Error("Output should include zk installation URL")
	}
	
	// Should return false for unavailable zk
	if result {
		t.Error("checkExternalDependencies should return false when zk unavailable")
	}
}

func TestCheckDatabases(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	
	// Create test environment with proper method
	env, err := config.GetDefaultViceEnv()
	if err != nil {
		t.Fatalf("Failed to create test environment: %v", err)
	}
	
	// Override the ContextData for our test
	env.ContextData = tmpDir
	
	// Capture output
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	result := checkDatabases(env)
	
	_ = w.Close()
	os.Stdout = originalStdout
	
	// Read output
	output := make([]byte, 1000)
	n, _ := r.Read(output)
	outputStr := string(output[:n])
	
	// Should mention database files
	if !strings.Contains(outputStr, "database") {
		t.Error("Output should mention database status")
	}
	
	// Should complete without error
	_ = result
}

func TestGetDirectoryOverrides(t *testing.T) {
	// Set some test flags
	originalConfigDir := configDir
	originalDataDir := dataDir
	originalContext := contextFlag
	
	configDir = "/test/config"
	dataDir = "/test/data"
	contextFlag = "test-context"
	
	defer func() {
		configDir = originalConfigDir
		dataDir = originalDataDir
		contextFlag = originalContext
	}()
	
	overrides := getDirectoryOverrides()
	
	if overrides.ConfigDir != "/test/config" {
		t.Errorf("ConfigDir override = %q, want %q", overrides.ConfigDir, "/test/config")
	}
	
	if overrides.DataDir != "/test/data" {
		t.Errorf("DataDir override = %q, want %q", overrides.DataDir, "/test/data")
	}
	
	if overrides.Context != "test-context" {
		t.Errorf("Context override = %q, want %q", overrides.Context, "test-context")
	}
}